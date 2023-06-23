package sendmessage

import (
	"context"
	"errors"
	"fmt"
	"time"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	problemsrepo "github.com/karasunokami/chat-service/internal/repositories/problems"
	sendmanagermessagejob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/send-manager-message"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=sendmessagemocks

var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrProblemNotFound = errors.New("problem not created")
)

type messagesRepository interface {
	CreateFullVisible(
		ctx context.Context,
		reqID types.RequestID,
		problemID types.ProblemID,
		chatID types.ChatID,
		authorID types.UserID,
		msgBody string,
	) (*messagesrepo.Message, error)
}

type outboxService interface {
	Put(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error)
}

type problemsRepository interface {
	GetAssignedProblemID(ctx context.Context, managerID types.UserID, chatID types.ChatID) (types.ProblemID, error)
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	messagesRepository messagesRepository `option:"mandatory" validate:"required"`
	outboxService      outboxService      `option:"mandatory" validate:"required"`
	problemsRepository problemsRepository `option:"mandatory" validate:"required"`
	txtor              transactor         `option:"mandatory" validate:"required"`
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

	problemID, err := u.problemsRepository.GetAssignedProblemID(ctx, req.ManagerID, req.ChatID)
	if err != nil {
		if errors.Is(err, problemsrepo.ErrNotFound) {
			return Response{}, ErrProblemNotFound
		}

		return Response{}, fmt.Errorf("problems repository, get assigned problem id, err=%v", err)
	}

	var (
		msgID        types.MessageID
		msgCreatedAt time.Time
	)

	err = u.txtor.RunInTx(ctx, func(ctx context.Context) error {
		msg, err := u.messagesRepository.CreateFullVisible(ctx, req.ID, problemID, req.ChatID, req.ManagerID, req.MessageBody)
		if err != nil {
			return fmt.Errorf("messages repository, create full visible, err=%w", err)
		}

		pl, err := sendmanagermessagejob.MarshalPayload(msg.ID, req.ManagerID)
		if err != nil {
			return fmt.Errorf("marshal message id payload, err=%w", err)
		}

		_, err = u.outboxService.Put(ctx, sendmanagermessagejob.Name, pl, time.Now())
		if err != nil {
			return fmt.Errorf("put send manager message job to outbox service, err=%w", err)
		}

		msgID = msg.ID
		msgCreatedAt = msg.CreatedAt

		return nil
	})
	if err != nil {
		return Response{}, fmt.Errorf("create manager message in transaction, err=%w", err)
	}

	return Response{
		MessageID: msgID,
		CreatedAt: msgCreatedAt,
	}, nil
}
