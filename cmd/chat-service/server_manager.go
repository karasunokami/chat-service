package main

import (
	"fmt"

	"github.com/karasunokami/chat-service/internal/config"
	server "github.com/karasunokami/chat-service/internal/server-manager"
	managerv1 "github.com/karasunokami/chat-service/internal/server-manager/v1"
	canreceiveproblems "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"
	freehands "github.com/karasunokami/chat-service/internal/usecases/manager/free-hands"
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
		managerServerConfig.Addr,
		managerServerConfig.AllowOrigins,
		managerServerConfig.SecWsProtocol,
		managerServerConfig.RequiredAccess.Resource,
		managerServerConfig.RequiredAccess.Role,
		deps.errHandler.Handle,
		deps.managerLogger,
		deps.managerSwagger,
		deps.kcClient,
		serverHandlers,
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

	// create manager handlers
	serverV1Handlers, err := managerv1.NewHandlers(managerv1.NewOptions(
		canReceiveProblemsUseCase,
		freeHandsUseCase,
	))
	if err != nil {
		return managerv1.Handlers{}, fmt.Errorf("create v1 handlers: %v", err)
	}

	return serverV1Handlers, nil
}
