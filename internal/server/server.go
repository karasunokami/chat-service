package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/karasunokami/chat-service/internal/middlewares"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	websocketstream "github.com/karasunokami/chat-service/internal/websocket-stream"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	bodyLimit = "12KB" // ~ 3000 characters * 4 bytes.

	readHeaderTimeout = time.Second
	shutdownTimeout   = 3 * time.Second
)

//go:generate options-gen -out-filename=server_options.gen.go -from-struct=Options
type Options struct {
	logger            *zap.Logger                  `option:"mandatory" validate:"required"`
	addr              string                       `option:"mandatory" validate:"required,hostname_port"`
	allowOrigins      []string                     `option:"mandatory" validate:"min=1"`
	wsSecProtocol     string                       `option:"mandatory" validate:"required"`
	requiredResource  string                       `option:"mandatory" validate:"required"`
	requiredRole      string                       `option:"mandatory" validate:"required"`
	handlersRegistrar func(e *echo.Echo)           `option:"mandatory" validate:"required"`
	introspector      middlewares.Introspector     `option:"mandatory" validate:"required"`
	eventStream       eventstream.EventStream      `option:"mandatory" validate:"required"`
	eventsAdapter     websocketstream.EventAdapter `option:"mandatory" validate:"required"`
}

type Server struct {
	lg  *zap.Logger
	srv *http.Server
}

func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	e := echo.New()

	e.Use(
		middlewares.NewLoggerMiddleware(opts.logger),
		middleware.RecoverWithConfig(middleware.RecoverConfig{
			DisableStackAll: true,
			LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
				opts.logger.Error("Recovered", zap.ByteString("stack", stack), zap.Error(err))

				return err
			},
		}),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: opts.allowOrigins,
			AllowMethods: []string{echo.POST},
		}),
		middlewares.NewKeyCloakTokenAuth(opts.introspector, opts.requiredResource, opts.requiredRole),

		// max length of message is 3000 utf-8 symbols 3000. 4 bytes each = 12000 bytes / 1024 = 11.78 kB ~= 12 kB
		middleware.BodyLimit(bodyLimit),
	)

	opts.handlersRegistrar(e)

	srv := &http.Server{
		Addr:              opts.addr,
		Handler:           e,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	shutdownCh := make(chan struct{})
	srv.RegisterOnShutdown(func() {
		close(shutdownCh)
	})

	wsHandler, err := websocketstream.NewHTTPHandler(websocketstream.NewOptions(
		opts.logger,
		opts.eventStream,
		opts.eventsAdapter,
		websocketstream.JSONEventWriter{},
		websocketstream.NewUpgrader(opts.allowOrigins, opts.wsSecProtocol),
		shutdownCh,
	))
	if err != nil {
		return nil, fmt.Errorf("create ws handler, err=%v", err)
	}

	e.GET("/ws", wsHandler.Serve)

	return &Server{
		lg:  opts.logger,
		srv: srv,
	}, nil
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
		s.lg.Info("Listen and serve", zap.String("addr", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}
