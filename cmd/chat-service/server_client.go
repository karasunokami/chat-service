package main

import (
	"fmt"

	"github.com/karasunokami/chat-service/internal/config"
	serverclient "github.com/karasunokami/chat-service/internal/server-client"
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/client/get-history"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/client/send-message"
)

const nameServerClient = "server-client"

func initServerClient(
	deps serverDeps,
	clientServerConfig config.ClientServerConfig,
) (*serverclient.Server, error) {
	serverHandlers, err := initServerHandlers(deps)
	if err != nil {
		return nil, fmt.Errorf("init server hanlders, err=%v", err)
	}

	// build server client
	srv, err := serverclient.New(serverclient.NewOptions(
		clientServerConfig.Addr,
		clientServerConfig.AllowOrigins,
		clientServerConfig.RequiredAccess.Resource,
		clientServerConfig.RequiredAccess.Role,
		deps.errHandler.Handle,
		deps.logger,
		deps.swagger,
		serverHandlers,
		deps.kcClient,
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

	sendMessageUseCase, err := sendmessage.New(sendmessage.NewOptions(deps.chatRepo, deps.msgRepo, deps.problemsRepo, deps.db))
	if err != nil {
		return clientv1.Handlers{}, fmt.Errorf("init send message usecase: %v", err)
	}

	// create client handlers
	serverV1Handlers, err := clientv1.NewHandlers(clientv1.NewOptions(deps.logger, getHistoryUseCase, sendMessageUseCase))
	if err != nil {
		return clientv1.Handlers{}, fmt.Errorf("create v1 handlers: %v", err)
	}

	return serverV1Handlers, nil
}
