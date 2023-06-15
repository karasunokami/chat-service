package main

import (
	"fmt"

	"github.com/karasunokami/chat-service/internal/config"
	"github.com/karasunokami/chat-service/internal/server"
	servermanager "github.com/karasunokami/chat-service/internal/server-manager"
	managerevents "github.com/karasunokami/chat-service/internal/server-manager/events"
	managerv1 "github.com/karasunokami/chat-service/internal/server-manager/v1"
	canreceiveproblems "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"
	closechat "github.com/karasunokami/chat-service/internal/usecases/manager/close-chat"
	freehands "github.com/karasunokami/chat-service/internal/usecases/manager/free-hands"
	getchats "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/manager/get-history"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/manager/send-message"
)

const nameServerManager = "server-manager"

func initServerManager(
	deps serverDeps,
	managerServerConfig config.ManagerServerConfig,
) (*server.Server, error) {
	serverHandlers, err := initManagerServerHandlers(deps)
	if err != nil {
		return nil, fmt.Errorf("init server hanlders, err=%v", err)
	}

	// build manager server client
	srv, err := server.New(server.NewOptions(
		deps.managerLogger,
		managerServerConfig.Addr,
		managerServerConfig.AllowOrigins,
		managerServerConfig.SecWsProtocol,
		managerServerConfig.RequiredAccess.Resource,
		managerServerConfig.RequiredAccess.Role,
		servermanager.NewHandlersRegistrar(deps.managerSwagger, serverHandlers, deps.errHandler.Handle),
		deps.kcClient,
		deps.eventsStream,
		managerevents.Adapter{},
	))
	if err != nil {
		return nil, fmt.Errorf("build server: %v", err)
	}

	return srv, nil
}

func initManagerServerHandlers(deps serverDeps) (managerv1.Handlers, error) {
	// create use cases
	canReceiveProblemsUseCase, err := canreceiveproblems.New(canreceiveproblems.NewOptions(
		deps.managerLoad,
		deps.managerPool,
	))
	if err != nil {
		return managerv1.Handlers{}, fmt.Errorf("init can receive problems usecase: %v", err)
	}

	freeHandsUseCase, err := freehands.New(freehands.NewOptions(
		deps.managerLoad,
		deps.managerPool,
	))
	if err != nil {
		return managerv1.Handlers{}, fmt.Errorf("init free hands usecase: %v", err)
	}

	getChatsUseCase, err := getchats.New(getchats.NewOptions(deps.chatRepo))
	if err != nil {
		return managerv1.Handlers{}, fmt.Errorf("init get chats usecase: %v", err)
	}

	getHistoryUseCase, err := gethistory.New(gethistory.NewOptions(deps.msgRepo))
	if err != nil {
		return managerv1.Handlers{}, fmt.Errorf("init get history usecase: %v", err)
	}

	sendMessageUseCase, err := sendmessage.New(sendmessage.NewOptions(deps.msgRepo, deps.outboxService, deps.problemsRepo, deps.db))
	if err != nil {
		return managerv1.Handlers{}, fmt.Errorf("init send message usecase: %v", err)
	}

	closeChatUseCase, err := closechat.New(closechat.NewOptions(
		deps.outboxService,
		deps.problemsRepo,
		deps.msgRepo,
		deps.db,
	))
	if err != nil {
		return managerv1.Handlers{}, fmt.Errorf("init resolve problem usecase: %v", err)
	}

	// create manager handlers
	serverV1Handlers, err := managerv1.NewHandlers(managerv1.NewOptions(
		canReceiveProblemsUseCase,
		freeHandsUseCase,
		getChatsUseCase,
		getHistoryUseCase,
		sendMessageUseCase,
		closeChatUseCase,
	))
	if err != nil {
		return managerv1.Handlers{}, fmt.Errorf("create v1 handlers: %v", err)
	}

	return serverV1Handlers, nil
}
