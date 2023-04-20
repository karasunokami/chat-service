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

	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/sync/errgroup"
)

var (
	configPath          = flag.String("config", "./configs/config.example.toml", "Path to config file")
	swaggerClientV1Path = flag.String("v1client", "./api/client.v1.swagger.yml", "Path to swagger client v1 file")
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}

func run() (errReturned error) {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.ParseAndValidate(*configPath)
	if err != nil {
		return fmt.Errorf("parse and validate config %q: %v", *configPath, err)
	}

	logger.MustInit(logger.NewOptions(
		cfg.Log.Level,
		logger.WithEnv(cfg.Global.Env),
		logger.WithSentryDSN(cfg.Sentry.Dsn),
	))
	logger.Sync()

	t, err := openapi3.NewLoader().LoadFromFile(*swaggerClientV1Path)
	if err != nil {
		return fmt.Errorf("load openapi from file, err=%v", err)
	}

	srvDebug, err := serverdebug.New(serverdebug.NewOptions(
		cfg.Servers.Debug.Addr,
		t,
	))
	if err != nil {
		return fmt.Errorf("init debug server: %v", err)
	}

	srvClient, err := initServerClient(
		cfg.Servers.Client,
		t,
		cfg.Clients.KeycloakClient,
		cfg.Servers.Client.RequiredAccess,
		cfg.Global,
	)
	if err != nil {
		return fmt.Errorf("init client server: %v", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	// Run servers.
	eg.Go(func() error { return srvDebug.Run(ctx) })
	eg.Go(func() error { return srvClient.Run(ctx) })

	// Run services.
	// Ждут своего часа.
	// ...

	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("wait app stop: %v", err)
	}

	return nil
}
