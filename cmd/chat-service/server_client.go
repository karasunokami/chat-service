package main

import (
	"fmt"

	"github.com/karasunokami/chat-service/internal/config"
	"github.com/karasunokami/chat-service/internal/server"
	serverclient "github.com/karasunokami/chat-service/internal/server-client"
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/client/get-history"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/client/send-message"
)

const nameServerClient = "server-client"

func initServerClient(
	deps serverDeps,
	clientServerConfig config.ClientServerConfig,
) (*server.Server, error) {
	serverHandlers, err := initServerHandlers(deps)
	if err != nil {
		return nil, fmt.Errorf("init server hanlders, err=%v", err)
	}

	// build server client
	srv, err := server.New(server.NewOptions(
		deps.clientLogger,
		clientServerConfig.Addr,
		clientServerConfig.AllowOrigins,
		clientServerConfig.SecWsProtocol,
		clientServerConfig.RequiredAccess.Resource,
		clientServerConfig.RequiredAccess.Role,
		serverclient.NewHandlersRegistrar(deps.clientSwagger, serverHandlers, deps.errHandler.Handle),
		deps.kcClient,
		deps.eventsStream,
	))
	if err != nil {
		return nil, fmt.Errorf("build server: %v", err)
	}

	return srv, nil
}

func initServerHandlers(deps serverDeps) (clientv1.Handlers, error) {
	// create use cases
	getHistoryUseCase, err := gethistory.New(gethistory.NewOptions(deps.msgRepo))
	if err != nil {
		return clientv1.Handlers{}, fmt.Errorf("init get history usecase: %v", err)
	}

	sendMessageUseCase, err := sendmessage.New(sendmessage.NewOptions(
		deps.chatRepo,
		deps.msgRepo,
		deps.outboxService,
		deps.problemsRepo,
		deps.db,
	))
	if err != nil {
		return clientv1.Handlers{}, fmt.Errorf("init send message usecase: %v", err)
	}

	// create client handlers
	serverV1Handlers, err := clientv1.NewHandlers(clientv1.NewOptions(deps.clientLogger, getHistoryUseCase, sendMessageUseCase))
	if err != nil {
		return clientv1.Handlers{}, fmt.Errorf("create v1 handlers: %v", err)
	}

	return serverV1Handlers, nil
}
