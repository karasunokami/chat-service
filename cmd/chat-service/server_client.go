package main

import (
	"fmt"

	keycloakclient "github.com/karasunokami/chat-service/internal/clients/keycloak"
	"github.com/karasunokami/chat-service/internal/config"
	chatsrepo "github.com/karasunokami/chat-service/internal/repositories/chats"
	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	problemsrepo "github.com/karasunokami/chat-service/internal/repositories/problems"
	serverclient "github.com/karasunokami/chat-service/internal/server-client"
	"github.com/karasunokami/chat-service/internal/server-client/errhandler"
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"
	"github.com/karasunokami/chat-service/internal/store"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/client/get-history"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/client/send-message"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
)

const nameServerClient = "server-client"

func initServerClient(
	lg *zap.Logger,
	clientServerConfig config.ClientServerConfig,
	v1Swagger *openapi3.T,
	kcClientConfig config.KeycloakClientConfig,
	reqAccConfig config.RequiredAccessConfig,
	globalConfig config.GlobalConfig,
	messagesRepo *messagesrepo.Repo,
	chatsRepo *chatsrepo.Repo,
	problemsRepo *problemsrepo.Repo,
	db *store.Database,
) (*serverclient.Server, error) {
	getHistoryUseCase, err := gethistory.New(gethistory.NewOptions(messagesRepo))
	if err != nil {
		return nil, fmt.Errorf("init get history usecase: %v", err)
	}

	sendMessageUseCase, err := sendmessage.New(sendmessage.NewOptions(chatsRepo, messagesRepo, problemsRepo, db))
	if err != nil {
		return nil, fmt.Errorf("init send message usecase: %v", err)
	}

	v1Handlers, err := clientv1.NewHandlers(clientv1.NewOptions(lg, getHistoryUseCase, sendMessageUseCase))
	if err != nil {
		return nil, fmt.Errorf("create v1 handlers: %v", err)
	}

	kcClient, err := keycloakclient.New(keycloakclient.NewOptions(
		kcClientConfig.BasePath,
		kcClientConfig.Realm,
		kcClientConfig.ClientID,
		kcClientConfig.ClientSecret,
		keycloakclient.WithDebugMode(kcClientConfig.DebugMode),
	))
	if err != nil {
		return nil, fmt.Errorf("cretae new keycloak client, err=%v", err)
	}

	if kcClientConfig.DebugMode && globalConfig.IsInProdEnv() {
		lg.Warn("Attention! Keycloak client is in debug mode and env is prod")
	}

	errHandler, err := errhandler.New(errhandler.NewOptions(lg, globalConfig.IsInProdEnv(), errhandler.ResponseBuilder))
	if err != nil {
		return nil, fmt.Errorf("init error handler, err=%v", err)
	}

	srv, err := serverclient.New(serverclient.NewOptions(
		lg,
		clientServerConfig.Addr,
		clientServerConfig.AllowOrigins,
		v1Swagger,
		v1Handlers,
		kcClient,
		reqAccConfig.Resource,
		reqAccConfig.Role,
		errHandler.Handle,
	))
	if err != nil {
		return nil, fmt.Errorf("build server: %v", err)
	}

	return srv, nil
}
