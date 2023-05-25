package main

import (
	"fmt"

	"github.com/karasunokami/chat-service/internal/config"
	server "github.com/karasunokami/chat-service/internal/server-manager"
	managerv1 "github.com/karasunokami/chat-service/internal/server-manager/v1"
)

const nameServerManager = "server-manager"

func initServerManager(
	deps serverDeps,
	clientServerConfig config.ClientServerConfig,
) (*server.Server, error) {
	var serverHandlers managerv1.ServerInterface
	// serverHandlers, err := initManagerServerHandlers(deps)
	// if err != nil {
	//	return nil, fmt.Errorf("init server hanlders, err=%v", err)
	//}

	// build server client
	srv, err := server.New(server.NewOptions(
		clientServerConfig.Addr,
		clientServerConfig.AllowOrigins,
		clientServerConfig.RequiredAccess.Resource,
		clientServerConfig.RequiredAccess.Role,
		deps.errHandler.Handle,
		deps.managerLogger,
		deps.clientSwagger,
		deps.kcClient,
		serverHandlers,
	))
	if err != nil {
		return nil, fmt.Errorf("build server: %v", err)
	}

	return srv, nil
}

// func initManagerServerHandlers(deps serverDeps) (managerv1.Handlers, error) {
//	// create use cases
//	getHistoryUseCase, err := gethistory.New(gethistory.NewOptions(deps.msgRepo))
//	if err != nil {
//		return managerv1.Handlers{}, fmt.Errorf("init get history usecase: %v", err)
//	}
//
//	sendMessageUseCase, err := sendmessage.New(sendmessage.NewOptions(
//		deps.chatRepo,
//		deps.msgRepo,
//		deps.outboxService,
//		deps.problemsRepo,
//		deps.db,
//	))
//	if err != nil {
//		return managerv1.Handlers{}, fmt.Errorf("init send message usecase: %v", err)
//	}
//
//	// create client handlers
//	serverV1Handlers, err := managerv1.NewHandlers(managerv1.NewOptions(deps.managerLogger, getHistoryUseCase, sendMessageUseCase))
//	if err != nil {
//		return managerv1.Handlers{}, fmt.Errorf("create v1 handlers: %v", err)
//	}
//
//	return serverV1Handlers, nil
//}
