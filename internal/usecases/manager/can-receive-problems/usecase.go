package canreceiveproblems

import (
	"context"
	"errors"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=mocks

var ErrInvalidRequest = errors.New("invalid request")

type managerLoadService interface {
	CanManagerTakeProblem(ctx context.Context, managerID types.UserID) (bool, error)
}

type managerPool interface {
	Contains(ctx context.Context, managerID types.UserID) (bool, error)
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

func (u UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	err := req.Validate()
	if err != nil {
		return Response{}, fmt.Errorf("validate request, err=%w", ErrInvalidRequest)
	}

	ex, err := u.managerPool.Contains(ctx, req.ManagerID)
	if err != nil {
		return Response{}, fmt.Errorf("manager pool contains, err=%v", err)
	}

	// manager already in pool
	if ex {
		return Response{
			Result: false,
		}, nil
	}

	res, err := u.managerLoadSvc.CanManagerTakeProblem(ctx, req.ManagerID)
	if err != nil {
		return Response{}, fmt.Errorf("managers load service can manager take problem, err=%v", err)
	}

	return Response{Result: res}, nil
}
