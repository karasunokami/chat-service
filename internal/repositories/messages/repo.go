package messagesrepo

import (
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"
)

//go:generate options-gen -out-filename=repo_options.gen.go -from-struct=Options
type Options struct {
	db *store.Database `option:"mandatory" validate:"required"`
}

type Repo struct {
	Options
}

func New(opts Options) (*Repo, error) {
	err := opts.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate options err=%v", err)
	}

	return &Repo{Options: opts}, nil
}
