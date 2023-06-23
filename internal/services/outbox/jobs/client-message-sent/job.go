package clientmessagesentjob

import (
	"context"
	"errors"
	"fmt"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	problemsrepo "github.com/karasunokami/chat-service/internal/repositories/problems"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/internal/services/outbox"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/job_mock.gen.go -package=clientmessagesentjobmocks

const Name = "client-message-sent"

//go:generate options-gen -out-filename=job_options.gen.go -from-struct=Options
type Options struct {
	eventStream  eventStream  `option:"mandatory" validate:"required"`
	msgRepo      messageRepo  `option:"mandatory" validate:"required"`
	problemsRepo problemsRepo `option:"mandatory" validate:"required"`
}

type eventStream interface {
	Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error
}

type messageRepo interface {
	GetMessageByID(ctx context.Context, msgID types.MessageID) (*messagesrepo.Message, error)
}

type problemsRepo interface {
	GetManagerID(ctx context.Context, problemID types.ProblemID) (types.UserID, error)
}

type Job struct {
	outbox.DefaultJob
	eventStream  eventStream
	msgRepo      messageRepo
	problemsRepo problemsRepo
}

func New(opts Options) (*Job, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Job{
		eventStream:  opts.eventStream,
		msgRepo:      opts.msgRepo,
		problemsRepo: opts.problemsRepo,
	}, nil
}

func (j *Job) Name() string {
	return Name
}

func (j *Job) Handle(ctx context.Context, payload string) error {
	jp, err := outbox.UnmarshalMessageIDPayload(payload)
	if err != nil {
		return fmt.Errorf("unmarshal jobPayload, err=%v", err)
	}

	msg, err := j.msgRepo.GetMessageByID(ctx, jp.MessageID)
	if err != nil {
		return fmt.Errorf("msg repo get message by id, err=%v", err)
	}

	err = j.eventStream.Publish(ctx, msg.AuthorID, eventstream.NewMessageSentEvent(
		types.NewEventID(),
		msg.InitialRequestID,
		msg.ID,
	))
	if err != nil {
		return fmt.Errorf("publish message to event stream, err=%v", err)
	}

	if !msg.IsService && !msg.ProblemID.IsZero() {
		managerID, err := j.problemsRepo.GetManagerID(ctx, msg.ProblemID)
		if err != nil {
			if errors.Is(err, problemsrepo.ErrNotFound) {
				return nil
			}

			return fmt.Errorf("get manager id by problem id, err=%v", err)
		}

		err = j.eventStream.Publish(ctx, managerID, eventstream.NewNewManagerMessageEvent(
			types.NewEventID(),
			msg.InitialRequestID,
			msg.ChatID,
			msg.ID,
			msg.CreatedAt,
			msg.Body,
			msg.AuthorID,
		))
		if err != nil {
			return fmt.Errorf("publish message to event stream, err=%v", err)
		}
	}

	return nil
}
