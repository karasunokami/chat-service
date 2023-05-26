package errhandler

import (
	"fmt"
	"net/http"

	"github.com/karasunokami/chat-service/internal/errors"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

//go:generate options-gen -out-filename=errhandler_options.gen.go -from-struct=Options
type Options struct {
	logger          *zap.Logger                                    `option:"mandatory" validate:"required"`
	productionMode  bool                                           `option:"mandatory"`
	responseBuilder func(code int, msg string, details string) any `option:"mandatory" validate:"required"`
}

type Handler struct {
	lg              *zap.Logger
	productionMode  bool
	responseBuilder func(code int, msg string, details string) any
}

func New(opts Options) (Handler, error) {
	if err := opts.Validate(); err != nil {
		return Handler{}, fmt.Errorf("validate options, err=%v", err)
	}

	return Handler{
		lg:              opts.logger,
		productionMode:  opts.productionMode,
		responseBuilder: opts.responseBuilder,
	}, nil
}

func (h Handler) Handle(err error, eCtx echo.Context) {
	code, msg, details := errors.ProcessServerError(err)

	if h.productionMode {
		details = ""
	}

	eCtxErr := eCtx.JSON(http.StatusOK, h.responseBuilder(code, msg, details))
	if eCtxErr != nil {
		h.lg.Error("put json to echo context", zap.Error(eCtxErr))
	}
}
