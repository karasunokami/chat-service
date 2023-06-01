package clientv1

import (
	"context"
	"fmt"

	gethistory "github.com/karasunokami/chat-service/internal/usecases/client/get-history"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/client/send-message"

	"go.uber.org/zap"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/handlers_mocks.gen.go -package=clientv1mocks
type getHistoryUseCase interface {
	Handle(ctx context.Context, req gethistory.Request) (gethistory.Response, error)
}

//go:generate mockgen -source=$GOFILE -destination=mocks/handlers_mocks.gen.go -package=clientv1mocks
type sendMessageUseCase interface {
	Handle(ctx context.Context, req sendmessage.Request) (sendmessage.Response, error)
}

//go:generate options-gen -out-filename=handler_options.gen.go -from-struct=Options
type Options struct {
	logger      *zap.Logger        `option:"mandatory" validate:"required"`
	getHistory  getHistoryUseCase  `option:"mandatory" validate:"required"`
	sendMessage sendMessageUseCase `option:"mandatory" validate:"required"`
}

type Handlers struct {
	Options
}

func NewHandlers(opts Options) (Handlers, error) {
	if err := opts.Validate(); err != nil {
		return Handlers{}, fmt.Errorf("validate options, err=%v", err)
	}

	return Handlers{Options: opts}, nil
}
