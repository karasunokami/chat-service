package serverclient

import (
	"context"
	"fmt"

	keycloakclient "github.com/karasunokami/chat-service/internal/clients/keycloak"
	"github.com/karasunokami/chat-service/internal/server"
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"
	inmemeventstream "github.com/karasunokami/chat-service/internal/services/event-stream/in-mem"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Server struct {
	innerServer *server.Server
}

//go:generate options-gen -out-filename=server_options.gen.go -from-struct=Options
type Options struct {
	addr           string                    `option:"mandatory" validate:"required,hostname_port"`
	allowOrigins   []string                  `option:"mandatory" validate:"min=1"`
	secWsProtocol  string                    `option:"mandatory" validate:"required"`
	resource       string                    `option:"mandatory" validate:"required"`
	role           string                    `option:"mandatory" validate:"required"`
	errorHandler   echo.HTTPErrorHandler     `option:"mandatory" validate:"required"`
	logger         *zap.Logger               `option:"mandatory" validate:"required"`
	swagger        *openapi3.T               `option:"mandatory" validate:"required"`
	keycloakClient *keycloakclient.Client    `option:"mandatory" validate:"required"`
	v1Handlers     clientv1.ServerInterface  `option:"mandatory" validate:"required"`
	eventsStream   *inmemeventstream.Service `option:"mandatory" validate:"required"`
}

func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate server options, err=%v", err)
	}

	innerServer, err := server.New(server.NewOptions(
		opts.addr,
		opts.allowOrigins,
		opts.secWsProtocol,
		opts.resource,
		opts.role,
		opts.errorHandler,
		opts.logger,
		opts.swagger,
		opts.keycloakClient,
		opts.eventsStream,
	))
	if err != nil {
		return nil, fmt.Errorf("init inner server, err=%v", err)
	}

	s := &Server{
		innerServer: innerServer,
	}

	s.innerServer.RegisterClientHandlers(clientv1.RegisterHandlers, opts.v1Handlers)

	return s, nil
}

func (s *Server) Run(ctx context.Context) error {
	return s.innerServer.Run(ctx)
}
