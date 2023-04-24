package clientv1

import (
	"fmt"

	"go.uber.org/zap"
)

//go:generate options-gen -out-filename=handler_options.gen.go -from-struct=Options
type Options struct {
	logger *zap.Logger `option:"mandatory" validate:"required"`
	// Ждут своего часа.
}

type Handlers struct {
	Options
}

func NewHandlers(opts Options) (Handlers, error) {
	err := opts.Validate()
	if err != nil {
		return Handlers{}, fmt.Errorf("validate options, err=%v", err)
	}

	return Handlers{Options: opts}, nil
}
