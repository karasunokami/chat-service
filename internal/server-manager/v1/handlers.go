package managerv1

import (
	"context"
	"fmt"

	canreceiveproblems "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/handlers_mocks.gen.go -package=managerv1mocks

type canReceiveProblemsUseCase interface {
	Handle(ctx context.Context, req canreceiveproblems.Request) (canreceiveproblems.Response, error)
}

//go:generate options-gen --out-filename=handlers_options.gen.go --from-struct=Options
type Options struct {
	canReceiveProblems canReceiveProblemsUseCase `option:"mandatory" validate:"required"`
}

type Handlers struct {
	Options
}

func NewHandlers(opts Options) (Handlers, error) {
	err := opts.Validate()
	if err != nil {
		return Handlers{}, fmt.Errorf("validate options, err=%v", err)
	}

	return Handlers{opts}, nil
}
