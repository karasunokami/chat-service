package closechat

import (
	"context"
	"errors"
	"fmt"
	"time"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	problemsrepo "github.com/karasunokami/chat-service/internal/repositories/problems"
	chatclosed "github.com/karasunokami/chat-service/internal/services/outbox/jobs/chat-closed"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=mocks

const messageResolvedMessageBody = `Your question has been marked as resolved.
Thank you for being with us!`

var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrProblemNotFound = errors.New("problem not found")
)

type problemsRepo interface {
	MarkProblemAsResolved(ctx context.Context, problemID types.ProblemID) error
	GetAssignedProblemID(
		ctx context.Context,
		managerID types.UserID,
		chatID types.ChatID,
	) (types.ProblemID, error)
}

type messagesRepo interface {
	CreateClientService(
		ctx context.Context,
		problemID types.ProblemID,
		chatID types.ChatID,
		msgBody string,
	) (*messagesrepo.Message, error)
}

type outboxService interface {
	Put(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error)
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

//go:generate options-gen -out-filename=service_options.gen.go -from-struct=Options
type Options struct {
	outboxService outboxService `option:"mandatory" validate:"required"`
	problemsRepo  problemsRepo  `option:"mandatory" validate:"required"`
	messagesRepo  messagesRepo  `option:"mandatory" validate:"required"`
	transactor    transactor    `option:"mandatory" validate:"required"`
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

	problemID, err := u.problemsRepo.GetAssignedProblemID(ctx, req.ManagerID, req.ChatID)
	if err != nil {
		if errors.Is(err, problemsrepo.ErrNotFound) {
			return ErrProblemNotFound
		}

		return fmt.Errorf("problems repo, get assigned problem id, err=%w", err)
	}

	err = u.transactor.RunInTx(ctx, func(ctx context.Context) error {
		err = u.problemsRepo.MarkProblemAsResolved(ctx, problemID)
		if err != nil {
			if errors.Is(err, problemsrepo.ErrNotFound) {
				return ErrProblemNotFound
			}

			return fmt.Errorf("problems repo, mark problem as resolved, err=%w", err)
		}

		msg, err := u.messagesRepo.CreateClientService(ctx, problemID, req.ChatID, messageResolvedMessageBody)
		if err != nil {
			return fmt.Errorf("messages repo, create service, err=%w", err)
		}

		payload, err := chatclosed.MarshalPayload(req.ManagerID, msg.ID, req.ID)
		if err != nil {
			return fmt.Errorf("marshal chat closed job payload, err=%v", err)
		}

		_, err = u.outboxService.Put(ctx, chatclosed.Name, payload, time.Now())
		if err != nil {
			return fmt.Errorf("put job to outbox service, err=%w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("mark problem as resolved in transaction, err=%w", err)
	}

	return nil
}
