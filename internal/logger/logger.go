package logger

import (
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"syscall"

	"github.com/karasunokami/chat-service/internal/buildinfo"
	"github.com/karasunokami/chat-service/internal/config"

	"github.com/TheZeroSlave/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate options-gen -out-filename=logger_options.gen.go -from-struct=Options
type Options struct {
	level     string `option:"mandatory" validate:"required,oneof=debug info warn error"`
	sentryDSN string `validate:"unix_addr"`
	env       string
}

var Al zap.AtomicLevel

func MustInit(opts Options) {
	if err := Init(opts); err != nil {
		panic(err)
	}
}

func Init(opts Options) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("validate options: %v", err)
	}

	al, err := zap.ParseAtomicLevel(opts.level)
	if err != nil {
		return fmt.Errorf("zap parse atomic level, err=%w", err)
	}

	Al = al

	err = configureLogger(al, opts)
	if err != nil {
		return fmt.Errorf("create logger, err=%w", err)
	}

	return nil
}

func Sync() {
	if err := zap.L().Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
		stdlog.Printf("cannot sync logger: %v", err)
	}
}

func configureLogger(al zap.AtomicLevel, opts Options) error {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.TimeKey = "T"
	cfg.NameKey = "component"
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder

	encoder := zapcore.NewJSONEncoder(cfg)

	if opts.env != config.GlobalEnvProd {
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(cfg)
	}

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), al),
	}

	if opts.sentryDSN != "" {
		sentryCore, err := createSentryCore(opts.sentryDSN, opts.env)
		if err != nil {
			return fmt.Errorf("create sentry zap core, err=%w", err)
		}

		cores = append(cores, sentryCore)
	}

	l := zap.New(zapcore.NewTee(cores...))
	zap.ReplaceGlobals(l)

	return nil
}

func createSentryCore(dsn, env string) (zapcore.Core, error) {
	sentryClient, err := NewSentryClient(dsn, env, buildinfo.BuildInfo.GoVersion)
	if err != nil {
		// in case of err it will return noop core. so we can safely attach it
		return nil, fmt.Errorf("create new sentry client, err=%v", err)
	}

	coreCfg := zapsentry.Configuration{
		Level:             zapcore.WarnLevel, // when to send message to sentry
		EnableBreadcrumbs: true,              // enable sending breadcrumbs to Sentry
		BreadcrumbLevel:   zapcore.InfoLevel, // at what level should we sent breadcrumbs to sentry
	}

	core, err := zapsentry.NewCore(coreCfg, zapsentry.NewSentryClientFromClient(sentryClient))
	if err != nil {
		// in case of err it will return noop core. so we can safely attach it
		return nil, fmt.Errorf("zapsentry new core, err=%v", err)
	}

	return core, nil
}
