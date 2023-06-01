package clientmessagesentjob

import (
	"context"
	"fmt"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/internal/services/outbox"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/job_mock.gen.go -package=clientmessagesentjobmocks

const Name = "client-message-sent"

//go:generate options-gen -out-filename=job_options.gen.go -from-struct=Options
type Options struct {
	eventStream eventStream       `option:"mandatory" validate:"required"`
	msgRepo     messageRepository `option:"mandatory" validate:"required"`
}

type eventStream interface {
	Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error
}

type messageRepository interface {
	GetMessageByID(ctx context.Context, msgID types.MessageID) (*messagesrepo.Message, error)
}

type Job struct {
	outbox.DefaultJob
	eventStream eventStream
	msgRepo     messageRepository
}

func New(opts Options) (*Job, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Job{
		eventStream: opts.eventStream,
		msgRepo:     opts.msgRepo,
	}, nil
}

func (j *Job) Name() string {
	return Name
}

func (j *Job) Handle(ctx context.Context, payload string) error {
	jp, err := unmarshalPayload(payload)
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

	return nil
}
