package sendmessage

import (
	"context"
	"errors"
	"fmt"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=sendmessagemocks

var (
	ErrInvalidRequest    = errors.New("invalid request")
	ErrChatNotCreated    = errors.New("chat not created")
	ErrProblemNotCreated = errors.New("problem not created")
)

type chatsRepository interface {
	CreateIfNotExists(ctx context.Context, userID types.UserID) (types.ChatID, error)
}

type messagesRepository interface {
	GetMessageByRequestID(ctx context.Context, reqID types.RequestID) (*messagesrepo.Message, error)
	CreateClientVisible(
		ctx context.Context,
		reqID types.RequestID,
		problemID types.ProblemID,
		chatID types.ChatID,
		authorID types.UserID,
		msgBody string,
	) (*messagesrepo.Message, error)
}

type problemsRepository interface {
	CreateIfNotExists(ctx context.Context, chatID types.ChatID) (types.ProblemID, error)
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	chatRepo     chatsRepository    `option:"mandatory" validate:"required"`
	msgRepo      messagesRepository `option:"mandatory" validate:"required"`
	problemsRepo problemsRepository `option:"mandatory" validate:"required"`
	txtor        transactor         `option:"mandatory" validate:"required"`
}

type UseCase struct {
	Options
}

func New(opts Options) (UseCase, error) {
	return UseCase{Options: opts}, opts.Validate()
}

func (u UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	err := req.Validate()
	if err != nil {
		return Response{}, ErrInvalidRequest
	}

	var newMessage *messagesrepo.Message

	err = u.txtor.RunInTx(ctx, func(ctx context.Context) error {
		msg, err := u.msgRepo.GetMessageByRequestID(ctx, req.ID)
		if err != nil && !errors.Is(err, messagesrepo.ErrMsgNotFound) {
			return fmt.Errorf("msg repo get message by requeset id, err=%v", err)
		}

		if msg != nil {
			newMessage = msg

			return nil
		}

		chatID, err := u.chatRepo.CreateIfNotExists(ctx, req.ClientID)
		if err != nil {
			return ErrChatNotCreated
		}

		problemID, err := u.problemsRepo.CreateIfNotExists(ctx, chatID)
		if err != nil {
			return ErrProblemNotCreated
		}

		msg, err = u.msgRepo.CreateClientVisible(ctx, req.ID, problemID, chatID, req.ClientID, req.MessageBody)
		if err != nil {
			return fmt.Errorf("msg repo crate client visible message, err=%v", err)
		}

		newMessage = msg

		return nil
	})
	if err != nil {
		return Response{}, fmt.Errorf("run in tx, err=%w", err)
	}

	return Response{
		AuthorID:  newMessage.AuthorID,
		MessageID: newMessage.ID,
		CreatedAt: newMessage.CreatedAt,
	}, nil
}
