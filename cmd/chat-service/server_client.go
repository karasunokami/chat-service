package main

import (
	"fmt"

	keycloakclient "github.com/karasunokami/chat-service/internal/clients/keycloak"
	"github.com/karasunokami/chat-service/internal/config"
	serverclient "github.com/karasunokami/chat-service/internal/server-client"
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
)

const nameServerClient = "server-client"

func initServerClient(
	clientServerConfig config.ClientServerConfig,
	v1Swagger *openapi3.T,
	kcClientConfig config.KeycloakClientConfig,
	reqAccConfig config.RequiredAccessConfig,
	globalConfig config.GlobalConfig,
) (*serverclient.Server, error) {
	lg := zap.L().Named(nameServerClient)

	v1Handlers, err := clientv1.NewHandlers(clientv1.NewOptions(lg))
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

	srv, err := serverclient.New(serverclient.NewOptions(
		lg,
		clientServerConfig.Addr,
		clientServerConfig.AllowOrigins,
		v1Swagger,
		v1Handlers,
		kcClient,
		reqAccConfig.Resource,
		reqAccConfig.Role,
	))
	if err != nil {
		return nil, fmt.Errorf("build server: %v", err)
	}

	return srv, nil
}
