package getchats

import (
	"context"
	"errors"
	"fmt"

	chatsrepo "github.com/karasunokami/chat-service/internal/repositories/chats"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=getchatsmocks

var ErrInvalidRequest = errors.New("invalid request")

type chatsRepo interface {
	GetManagerOpened(ctx context.Context, managerID types.UserID) ([]*chatsrepo.Chat, error)
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	chatsRepo chatsRepo `option:"mandatory" validate:"required"`
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
	if err := req.Validate(); err != nil {
		return Response{}, ErrInvalidRequest
	}

	chats, err := u.chatsRepo.GetManagerOpened(ctx, req.ManagerID)
	if err != nil {
		return Response{}, fmt.Errorf("get manager opened chats, err=%w", err)
	}

	result := make([]Chat, 0, len(chats))
	for _, m := range chats {
		result = append(result, Chat{
			ID:       m.ID,
			ClientID: m.ClientID,
		})
	}

	return Response{Chats: result}, nil
}
