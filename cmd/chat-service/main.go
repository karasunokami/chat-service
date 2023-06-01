package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/karasunokami/chat-service/internal/config"
	"github.com/karasunokami/chat-service/internal/logger"
	serverdebug "github.com/karasunokami/chat-service/internal/server-debug"

	"golang.org/x/sync/errgroup"
)

var configPath = flag.String("config", "./configs/config.example.toml", "Path to config file")

func main() {
	if err := run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}

func run() (errReturned error) {
	// parsing command line params
	flag.Parse()

	// creating signal notify context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// parsing configs
	cfg, err := config.ParseAndValidate(*configPath)
	if err != nil {
		return fmt.Errorf("parse and validate config %q: %v", *configPath, err)
	}

	// configure logger
	configureZap(cfg.Log.Level, cfg.Global.Env, cfg.Sentry.Dsn)
	defer logger.Sync()

	// init deps
	deps, err := startNewDeps(ctx, cfg)
	if err != nil {
		return fmt.Errorf("init server deps, err=%v", err)
	}

	defer deps.stop()

	// init servers
	srvDebug, err := serverdebug.New(serverdebug.NewOptions(
		cfg.Servers.Debug.Addr,
		deps.clientSwagger,
		deps.managerSwagger,
		deps.clientEventsSwagger,
	))
	if err != nil {
		return fmt.Errorf("init debug server: %v", err)
	}

	srvClient, err := initServerClient(deps, cfg.Servers.Client)
	if err != nil {
		return fmt.Errorf("init client server: %v", err)
	}

	srvManager, err := initServerManager(deps, cfg.Servers.Manager)
	if err != nil {
		return fmt.Errorf("init manager server: %v", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	// run servers
	eg.Go(func() error { return srvDebug.Run(ctx) })
	eg.Go(func() error { return srvClient.Run(ctx) })
	eg.Go(func() error { return srvManager.Run(ctx) })

	// run services
	eg.Go(func() error { return deps.outboxService.Run(ctx) })
	eg.Go(func() error { return deps.afcVerdictsProcessorService.Run(ctx) })

	// wait for command line signal
	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("wait app stop: %v", err)
	}

	return nil
}
