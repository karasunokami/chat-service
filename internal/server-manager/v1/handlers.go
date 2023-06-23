package managerv1

import (
	"context"
	"fmt"

	canreceiveproblems "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"
	closechat "github.com/karasunokami/chat-service/internal/usecases/manager/close-chat"
	freehands "github.com/karasunokami/chat-service/internal/usecases/manager/free-hands"
	getchats "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/manager/get-history"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/manager/send-message"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/handlers_mocks.gen.go -package=managerv1mocks

type canReceiveProblemsUseCase interface {
	Handle(ctx context.Context, req canreceiveproblems.Request) (canreceiveproblems.Response, error)
}

type freeHandsUseCase interface {
	Handle(ctx context.Context, req freehands.Request) error
}

type getChatsUseCase interface {
	Handle(ctx context.Context, req getchats.Request) (getchats.Response, error)
}

type getHistoryUseCase interface {
	Handle(ctx context.Context, req gethistory.Request) (gethistory.Response, error)
}

type sendMessageUseCase interface {
	Handle(ctx context.Context, req sendmessage.Request) (sendmessage.Response, error)
}

type closeChatUseCase interface {
	Handle(ctx context.Context, req closechat.Request) error
}

//go:generate options-gen --out-filename=handlers_options.gen.go --from-struct=Options
type Options struct {
	canReceiveProblems canReceiveProblemsUseCase `option:"mandatory" validate:"required"`
	freeHands          freeHandsUseCase          `option:"mandatory" validate:"required"`
	getChats           getChatsUseCase           `option:"mandatory" validate:"required"`
	getHistory         getHistoryUseCase         `option:"mandatory" validate:"required"`
	sendMessage        sendMessageUseCase        `option:"mandatory" validate:"required"`
	closeChat          closeChatUseCase          `option:"mandatory" validate:"required"`
}

type Handlers struct {
	Options
}

func NewHandlers(opts Options) (Handlers, error) {
	if err := opts.Validate(); err != nil {
		return Handlers{}, fmt.Errorf("validate options, err=%v", err)
	}

	return Handlers{opts}, nil
}
