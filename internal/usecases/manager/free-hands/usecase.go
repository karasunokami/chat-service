package freehands

import (
	"context"
	"errors"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=mocks

var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrManagerOverload = errors.New("manager overload")
)

type managerLoadService interface {
	CanManagerTakeProblem(ctx context.Context, managerID types.UserID) (bool, error)
}

type managerPool interface {
	Put(ctx context.Context, managerID types.UserID) error
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	managerLoadSvc managerLoadService `option:"mandatory" validate:"required"`
	managerPool    managerPool        `option:"mandatory" validate:"required"`
}

type UseCase struct {
	Options
}

func New(opts Options) (UseCase, error) {
	if err := opts.Validate(); err != nil {
		return UseCase{}, fmt.Errorf("validate options, err=%v", err)
	}

	return UseCase{opts}, nil
}

func (u UseCase) Handle(ctx context.Context, req Request) error {
	err := req.Validate()
	if err != nil {
		return fmt.Errorf("validate request, err=%w", ErrInvalidRequest)
	}

	can, err := u.managerLoadSvc.CanManagerTakeProblem(ctx, req.ManagerID)
	if err != nil {
		return fmt.Errorf("managers load service can manager take problem, err=%v", err)
	}

	if !can {
		return ErrManagerOverload
	}

	err = u.managerPool.Put(ctx, req.ManagerID)
	if err != nil {
		return fmt.Errorf("put manager in managers pool, err=%v", err)
	}

	return nil
}
