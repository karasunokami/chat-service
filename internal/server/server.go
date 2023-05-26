package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	keycloakclient "github.com/karasunokami/chat-service/internal/clients/keycloak"
	"github.com/karasunokami/chat-service/internal/middlewares"
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"
	managerv1 "github.com/karasunokami/chat-service/internal/server-manager/v1"

	oapimdlwr "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	readHeaderTimeout = time.Second
	shutdownTimeout   = 3 * time.Second
)

//go:generate options-gen -out-filename=server_options.gen.go -from-struct=Options
type Options struct {
	addr           string                 `option:"mandatory" validate:"required,hostname_port"`
	allowOrigins   []string               `option:"mandatory" validate:"min=1"`
	resource       string                 `option:"mandatory" validate:"required"`
	role           string                 `option:"mandatory" validate:"required"`
	errorHandler   echo.HTTPErrorHandler  `option:"mandatory" validate:"required"`
	logger         *zap.Logger            `option:"mandatory" validate:"required"`
	swagger        *openapi3.T            `option:"mandatory" validate:"required"`
	keycloakClient *keycloakclient.Client `option:"mandatory" validate:"required"`
}

type Server struct {
	lg  *zap.Logger
	srv *http.Server

	serverGroup *echo.Group
}

func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate server options, err=%v", err)
	}

	echoServer := echo.New()
	echoServer.HTTPErrorHandler = opts.errorHandler

	s := Server{
		lg: opts.logger,
		srv: &http.Server{
			Addr:              opts.addr,
			Handler:           echoServer,
			ReadHeaderTimeout: readHeaderTimeout,
		},
	}

	echoServer.Use(
		middlewares.NewLoggerMiddleware(s.lg),
		middleware.RecoverWithConfig(middleware.RecoverConfig{
			DisableStackAll: true,
			LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
				s.lg.Error("recovered", zap.ByteString("stack", stack), zap.Error(err))

				return err
			},
		}),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: opts.allowOrigins,
			AllowMethods: []string{echo.POST},
		}),
		middlewares.NewKeyCloakTokenAuth(opts.keycloakClient, opts.resource, opts.role),

		// max length of message is 3000 utf-8 symbols 3000. 4 bytes each = 12000 bytes / 1024 = 11.78 kB ~= 12 kB
		middleware.BodyLimit("12K"),
	)

	s.serverGroup = echoServer.Group("v1", oapimdlwr.OapiRequestValidatorWithOptions(opts.swagger, &oapimdlwr.Options{
		Options: openapi3filter.Options{
			ExcludeRequestBody:  false,
			ExcludeResponseBody: true,
			AuthenticationFunc:  openapi3filter.NoopAuthenticationFunc,
		},
	}))

	return &s, nil
}

func (s *Server) RegisterClientHandlers(
	f func(router clientv1.EchoRouter, si clientv1.ServerInterface),
	handlers clientv1.ServerInterface,
) {
	f(s.serverGroup, handlers)
}

func (s *Server) RegisterManagerHandlers(
	f func(router managerv1.EchoRouter, si managerv1.ServerInterface),
	handlers managerv1.ServerInterface,
) {
	f(s.serverGroup, handlers)
}

func (s *Server) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return s.srv.Shutdown(ctx)
	})

	eg.Go(func() error {
		s.lg.Info("listen and serve", zap.String("addr", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}
