package managerload

import (
	"context"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/problems_repository_mock.gen.go -package=mocks

type problemsRepository interface {
	GetManagerOpenProblemsCount(ctx context.Context, managerID types.UserID) (int, error)
}

//go:generate options-gen -out-filename=service_options.gen.go -from-struct=Options
type Options struct {
	maxProblemsAtTime int                `option:"mandatory" validation:"required,gte=1,lte=30"`
	problemsRepo      problemsRepository `option:"mandatory" validate:"required"`
}

type Service struct {
	Options
}

func New(opts Options) (*Service, error) {
	err := opts.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Service{opts}, nil
}
