package messagesrepo

import (
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"

	"go.uber.org/zap"
)

const serviceName = "messagesRepo"

//go:generate options-gen -out-filename=repo_options.gen.go -from-struct=Options
type Options struct {
	db *store.Database `option:"mandatory" validate:"required"`
}

type Repo struct {
	Options
	logger *zap.Logger
}

func New(opts Options) (*Repo, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options err=%v", err)
	}

	return &Repo{
		Options: opts,
		logger:  zap.L().Named(serviceName),
	}, nil
}
