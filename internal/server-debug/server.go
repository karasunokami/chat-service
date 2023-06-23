package serverdebug

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
	"time"

	"github.com/karasunokami/chat-service/internal/buildinfo"
	"github.com/karasunokami/chat-service/internal/logger"
	"github.com/karasunokami/chat-service/internal/middlewares"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

const (
	readHeaderTimeout = time.Second
	shutdownTimeout   = 3 * time.Second
)

//go:generate options-gen -out-filename=server_options.gen.go -from-struct=Options
type Options struct {
	addr                string      `option:"mandatory" validate:"required,hostname_port"`
	clientV1Swagger     *openapi3.T `option:"mandatory" validate:"required"`
	managerV1Swagger    *openapi3.T `option:"mandatory" validate:"required"`
	clientEventsSwagger *openapi3.T `option:"mandatory" validate:"required"`
}

type Server struct {
	lg                  *zap.Logger
	srv                 *http.Server
	clientV1Swagger     *openapi3.T
	managerV1Swagger    *openapi3.T
	clientEventsSwagger *openapi3.T
}

func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("opts validate, err=%w", err)
	}

	lg := zap.L().Named("server-debug")

	e := echo.New()
	e.Use(
		middlewares.NewLoggerMiddleware(lg),
		middleware.RecoverWithConfig(middleware.RecoverConfig{
			Skipper:           nil,
			StackSize:         0,
			DisableStackAll:   false,
			DisablePrintStack: false,
			LogLevel:          0,
			LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
				lg.Error("Recovered", zap.ByteString("stack", stack), zap.Error(err))

				return nil
			},
		}),
	)

	s := &Server{
		lg:                  lg,
		clientV1Swagger:     opts.clientV1Swagger,
		managerV1Swagger:    opts.managerV1Swagger,
		clientEventsSwagger: opts.clientEventsSwagger,
		srv: &http.Server{
			Addr:              opts.addr,
			Handler:           e,
			ReadHeaderTimeout: readHeaderTimeout,
		},
	}

	e.GET("/version", s.Version)
	e.GET("/debug/*", echo.WrapHandler(http.DefaultServeMux))
	e.GET("/debug/error", s.DebugError)
	e.GET("/schema/client", s.SchemaClient)
	e.GET("/schema/manager", s.SchemaManager)
	e.GET("/schema/client_events", s.SchemaClientEvents)

	e.PUT("/log/level", s.LogLevel)

	index := newIndexPage()
	index.addPage("/version", "Get build information")
	index.addPage("/debug/pprof", "Go std profiler")
	index.addPage("/debug/pprof/profile?seconds=30", "Take half-min profile")
	index.addPage("/debug/error", "Debug Sentry error event")
	index.addPage("/schema/client", "Get client Open API specification")
	index.addPage("/schema/manager", "Get manager Open API specification")
	e.GET("/", index.handler)

	return s, nil
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

func (s *Server) Version(c echo.Context) error {
	err := c.JSON(http.StatusOK, buildinfo.BuildInfo)
	if err != nil {
		return fmt.Errorf("encode build info to response, err=%v", err)
	}

	return nil
}

func (s *Server) LogLevel(c echo.Context) error {
	level := c.FormValue("level")

	l, err := zapcore.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("parse level from request, err=%v", err)
	}

	logger.Al.SetLevel(l)

	return nil
}

func (s *Server) DebugError(c echo.Context) error {
	s.lg.Error("Look for me in the Sentry")

	err := c.String(http.StatusOK, "event sent")
	if err != nil {
		return fmt.Errorf("c string")
	}

	return nil
}

func (s *Server) SchemaClient(c echo.Context) error {
	err := c.JSON(http.StatusOK, s.clientV1Swagger)
	if err != nil {
		return fmt.Errorf("echo context json, err=%v", err)
	}

	return nil
}

func (s *Server) SchemaManager(c echo.Context) error {
	err := c.JSON(http.StatusOK, s.managerV1Swagger)
	if err != nil {
		return fmt.Errorf("echo context json, err=%v", err)
	}

	return nil
}

func (s *Server) SchemaClientEvents(c echo.Context) error {
	err := c.JSON(http.StatusOK, s.clientEventsSwagger)
	if err != nil {
		return fmt.Errorf("echo context json, err=%v", err)
	}

	return nil
}
