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
	addr string `option:"mandatory" validate:"required,hostname_port"`
}

type Server struct {
	lg  *zap.Logger
	srv *http.Server
}

func New(opts Options) (*Server, error) {
	err := opts.Validate()
	if err != nil {
		return nil, fmt.Errorf("opts validate, err=%w", err)
	}

	lg := zap.L().Named("server-debug")

	e := echo.New()
	e.Use(middleware.Recover())

	s := &Server{
		lg: lg,
		srv: &http.Server{
			Addr:              opts.addr,
			Handler:           e,
			ReadHeaderTimeout: readHeaderTimeout,
		},
	}

	e.GET("/version", s.Version)
	e.GET("/debug/*", echo.WrapHandler(http.DefaultServeMux))
	e.GET("/debug/error", s.DebugError)

	e.PUT("/log/level", s.LogLevel)

	index := newIndexPage()
	index.addPage("/version", "Get build information")
	index.addPage("/debug/pprof", "Go std profiler")
	index.addPage("/debug/pprof/profile?seconds=30", "Take half-min profile")
	index.addPage("/debug/error", "Debug Sentry error event")
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
		s.lg.Info("listen and serve", zap.String("addr", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}

func (s *Server) Version(eCtx echo.Context) error {
	err := eCtx.JSON(200, buildinfo.BuildInfo)
	if err != nil {
		return fmt.Errorf("encode build info to response, err=%v", err)
	}

	return nil
}

func (s *Server) LogLevel(eCtx echo.Context) error {
	level := eCtx.FormValue("level")

	l, err := zapcore.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("parse level from request, err=%v", err)
	}

	logger.Al.SetLevel(l)

	return nil
}

func (s *Server) DebugError(eCtx echo.Context) error {
	s.lg.Error("look for me in the Sentry")

	err := eCtx.String(200, "event sent")
	if err != nil {
		return fmt.Errorf("ectx string")
	}

	return nil
}
