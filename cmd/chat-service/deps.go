package main

import (
	"context"
	"fmt"

	keycloakclient "github.com/karasunokami/chat-service/internal/clients/keycloak"
	"github.com/karasunokami/chat-service/internal/config"
	"github.com/karasunokami/chat-service/internal/logger"
	chatsrepo "github.com/karasunokami/chat-service/internal/repositories/chats"
	jobsrepo "github.com/karasunokami/chat-service/internal/repositories/jobs"
	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	problemsrepo "github.com/karasunokami/chat-service/internal/repositories/problems"
	"github.com/karasunokami/chat-service/internal/server-client/errhandler"
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"
	msgproducer "github.com/karasunokami/chat-service/internal/services/msg-producer"
	"github.com/karasunokami/chat-service/internal/services/outbox"
	sendclientmessagejob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/send-client-message"
	"github.com/karasunokami/chat-service/internal/store"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
)

type serverDeps struct {
	swagger *openapi3.T
	logger  *zap.Logger

	psqlClient *store.Client
	db         *store.Database

	msgRepo      *messagesrepo.Repo
	chatRepo     *chatsrepo.Repo
	jobsRepo     *jobsrepo.Repo
	problemsRepo *problemsrepo.Repo

	kcClient *keycloakclient.Client

	errHandler errhandler.Handler

	msgProducerService *msgproducer.Service
	outboxService      *outbox.Service
}

func startNewDeps(ctx context.Context, cfg config.Config) (serverDeps, error) {
	var (
		err error
		d   serverDeps
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

	d.jobsRepo, err = jobsrepo.New(jobsrepo.NewOptions(d.db))
	if err != nil {
		return serverDeps{}, fmt.Errorf("init jobs repo, err=%v", err)
	}

	// init keycloak client
	d.kcClient, err = initKeyCloakClient(d.logger, cfg.Clients.KeycloakClient, cfg.Global.IsInProdEnv())
	if err != nil {
		return serverDeps{}, fmt.Errorf("init init keycloak client, err=%v", err)
	}

	// init server resp errors handler
	errHandler, err := errhandler.New(errhandler.NewOptions(d.logger, cfg.Global.IsInProdEnv(), errhandler.ResponseBuilder))
	if err != nil {
		return serverDeps{}, fmt.Errorf("init error handler, err=%v", err)
	}
	d.errHandler = errHandler

	// init services
	d.msgProducerService, err = msgproducer.New(msgproducer.NewOptions(msgproducer.NewKafkaWriter(
		cfg.Services.MessageProducerService.Brokers,
		cfg.Services.MessageProducerService.Topic,
		cfg.Services.MessageProducerService.BatchSize,
	), msgproducer.WithEncryptKey(cfg.Services.MessageProducerService.EncryptKey)))
	if err != nil {
		return serverDeps{}, fmt.Errorf("init message producer service, err=%v", err)
	}

	d.outboxService, err = outbox.New(outbox.NewOptions(
		cfg.Services.OutboxService.Workers,
		cfg.Services.OutboxService.IdleTime,
		cfg.Services.OutboxService.ReserveFor,
		d.jobsRepo,
		d.db,
	))
	if err != nil {
		return serverDeps{}, fmt.Errorf("init outbox service, err=%v", err)
	}

	// register service jobs
	sendClientMessageJob, err := sendclientmessagejob.New(sendclientmessagejob.NewOptions(
		d.msgProducerService,
		d.msgRepo,
	))
	if err != nil {
		return serverDeps{}, fmt.Errorf("create send client message job, err=%v", err)
	}

	err = d.outboxService.RegisterJob(sendClientMessageJob)
	if err != nil {
		return serverDeps{}, fmt.Errorf("register send client message job, err=%v", err)
	}

	err = d.outboxService.Run(ctx)
	if err != nil {
		return serverDeps{}, fmt.Errorf("start outbox servcie, err=%v", err)
	}

	return d, nil
}

func (d serverDeps) stop() {
	if d.psqlClient != nil {
		err := d.psqlClient.Close()
		if err != nil {
			d.logger.Error("stop psql client", zap.Error(err))
		}
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
