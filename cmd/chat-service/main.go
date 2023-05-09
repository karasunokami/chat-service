package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	keycloakclient "github.com/karasunokami/chat-service/internal/clients/keycloak"
	"github.com/karasunokami/chat-service/internal/config"
	"github.com/karasunokami/chat-service/internal/logger"
	chatsrepo "github.com/karasunokami/chat-service/internal/repositories/chats"
	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	problemsrepo "github.com/karasunokami/chat-service/internal/repositories/problems"
	"github.com/karasunokami/chat-service/internal/server-client/errhandler"
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"
	serverdebug "github.com/karasunokami/chat-service/internal/server-debug"
	"github.com/karasunokami/chat-service/internal/store"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
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

	// init deps
	configureZap(cfg.Log.Level, cfg.Global.Env, cfg.Sentry.Dsn)

	deps, err := initServerDeps(ctx, cfg)
	if err != nil {
		return fmt.Errorf("init server deps, err=%v", err)
	}

	defer stopDeps(deps)

	// init servers
	srvDebug, err := serverdebug.New(serverdebug.NewOptions(cfg.Servers.Debug.Addr, deps.swagger))
	if err != nil {
		return fmt.Errorf("init debug server: %v", err)
	}

	srvClient, err := initServerClient(deps, cfg.Servers.Client)
	if err != nil {
		return fmt.Errorf("init client server: %v", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	// run servers
	eg.Go(func() error { return srvDebug.Run(ctx) })
	eg.Go(func() error { return srvClient.Run(ctx) })

	// wait for command line signal
	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("wait app stop: %v", err)
	}

	return nil
}

type serverDeps struct {
	swagger      *openapi3.T
	logger       *zap.Logger
	psqlClient   *store.Client
	db           *store.Database
	msgRepo      *messagesrepo.Repo
	chatRepo     *chatsrepo.Repo
	problemsRepo *problemsrepo.Repo
	kcClient     *keycloakclient.Client
	errHandler   errhandler.Handler
}

func initServerDeps(ctx context.Context, cfg config.Config) (serverDeps, error) {
	var (
		d   serverDeps
		err error
	)

	// init swagger client
	d.swagger, err = clientv1.GetSwagger()
	if err != nil {
		return serverDeps{}, fmt.Errorf("client v1 get swagger, err=%v", err)
	}

	// init logger
	d.logger = zap.L().Named(nameServerClient)

	// init psql client
	d.psqlClient, err = store.NewPSQLClient(store.NewPSQLOptions(
		cfg.Clients.PSQLClient.Address,
		cfg.Clients.PSQLClient.Username,
		cfg.Clients.PSQLClient.Password,
		cfg.Clients.PSQLClient.Database,
		store.WithDebug(cfg.Clients.PSQLClient.DebugMode),
	))
	if err != nil {
		return serverDeps{}, fmt.Errorf("create psql client, err=%v", err)
	}

	if cfg.Clients.PSQLClient.DebugMode && cfg.Global.IsInProdEnv() {
		d.logger.Warn("Attention! PSQL client is in debug mode and env is prod")
	}

	if err = d.psqlClient.Schema.Create(ctx); err != nil {
		return serverDeps{}, fmt.Errorf("psql client schema create, err=%v", err)
	}

	// init database client
	d.db = store.NewDatabase(d.psqlClient, d.logger)

	// init repositories
	d.msgRepo, err = messagesrepo.New(messagesrepo.NewOptions(d.db))
	if err != nil {
		return serverDeps{}, fmt.Errorf("init messages repo, err=%v", err)
	}

	d.chatRepo, err = chatsrepo.New(chatsrepo.NewOptions(d.db))
	if err != nil {
		return serverDeps{}, fmt.Errorf("init chats repo, err=%v", err)
	}

	d.problemsRepo, err = problemsrepo.New(problemsrepo.NewOptions(d.db))
	if err != nil {
		return serverDeps{}, fmt.Errorf("init problems repo, err=%v", err)
	}

	// init keycloak client
	d.kcClient, err = initKeyCloakClient(d.logger, cfg.Clients.KeycloakClient, cfg.Global.IsInProdEnv())
	if err != nil {
		return serverDeps{}, fmt.Errorf("init init keycloak client, err=%v", err)
	}

	// init server resp errors handler
	d.errHandler, err = errhandler.New(errhandler.NewOptions(d.logger, cfg.Global.IsInProdEnv(), errhandler.ResponseBuilder))
	if err != nil {
		return serverDeps{}, fmt.Errorf("init error handler, err=%v", err)
	}

	return d, nil
}

func stopDeps(deps serverDeps) {
	err := deps.psqlClient.Close()
	if err != nil {
		deps.logger.Error("stop psql client", zap.Error(err))
	}
}

func initKeyCloakClient(logger *zap.Logger, cfg config.KeycloakClientConfig, isProdEnv bool) (*keycloakclient.Client, error) {
	kcClient, err := keycloakclient.New(keycloakclient.NewOptions(
		cfg.BasePath,
		cfg.Realm,
		cfg.ClientID,
		cfg.ClientSecret,
		keycloakclient.WithDebugMode(cfg.DebugMode),
	))
	if err != nil {
		return nil, fmt.Errorf("cretae new keycloak client, err=%v", err)
	}

	if cfg.DebugMode && isProdEnv {
		logger.Warn("Attention! Keycloak client is in debug mode and env is prod")
	}

	return kcClient, nil
}

func configureZap(logLevel, env, dsn string) {
	logger.MustInit(logger.NewOptions(
		logLevel,
		logger.WithEnv(env),
		logger.WithSentryDSN(dsn),
	))
	logger.Sync()
}
