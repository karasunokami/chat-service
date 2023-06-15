package sendmanagermessagejob

import (
	"context"
	"fmt"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	msgproducer "github.com/karasunokami/chat-service/internal/services/msg-producer"
	"github.com/karasunokami/chat-service/internal/services/outbox"
	"github.com/karasunokami/chat-service/internal/types"
)

const Name = "send-manager-message"

//go:generate mockgen -source=$GOFILE -destination=mocks/job_mock.gen.go -package=sendmanagermessagejobmocks

type eventStream interface {
	Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error
}

type messageProducer interface {
	ProduceMessage(ctx context.Context, message msgproducer.Message) error
}

type messagesRepo interface {
	GetMessageByID(ctx context.Context, id types.MessageID) (*messagesrepo.Message, error)
	GetFirstProblemMessage(ctx context.Context, problemID types.ProblemID) (*messagesrepo.Message, error)
}

//go:generate options-gen -out-filename=job_options.gen.go -from-struct=Options
type Options struct {
	eventStream  eventStream     `option:"mandatory" validate:"required"`
	msgProducer  messageProducer `option:"mandatory" validate:"required"`
	messagesRepo messagesRepo    `option:"mandatory" validate:"required"`
}

type Job struct {
	Options
	outbox.DefaultJob
}

func New(opts Options) (*Job, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Job{Options: opts}, nil
}

func (j *Job) Name() string {
	return Name
}

func (j *Job) Handle(ctx context.Context, payload string) error {
	pl, err := UnmarshalPayload(payload)
	if err != nil {
		return fmt.Errorf("unmarshal message id payload, err=%v", err)
	}

	msg, err := j.messagesRepo.GetMessageByID(ctx, pl.MessageID)
	if err != nil {
		return fmt.Errorf("messages repo, get message by id, err=%v", err)
	}

	firstMessage, err := j.messagesRepo.GetFirstProblemMessage(ctx, msg.ProblemID)
	if err != nil {
		return fmt.Errorf("messages repo, get first problem message, err=%v", err)
	}

	err = j.msgProducer.ProduceMessage(ctx, repoMsgToProducerMsg(msg))
	if err != nil {
		return fmt.Errorf("msg producer, produce message, err=%v", err)
	}

	err = j.eventStream.Publish(ctx, pl.ManagerID, eventstream.NewNewManagerMessageEvent(
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

	err = j.eventStream.Publish(ctx, firstMessage.AuthorID, eventstream.NewNewMessageEvent(
		types.NewEventID(),
		msg.InitialRequestID,
		msg.ChatID,
		msg.ID,
		msg.CreatedAt,
		msg.Body,
		msg.AuthorID,
		msg.IsService,
	))
	if err != nil {
		return fmt.Errorf("publish message to event stream, err=%v", err)
	}

	return nil
}

func repoMsgToProducerMsg(msg *messagesrepo.Message) msgproducer.Message {
	return msgproducer.Message{
		ID:         msg.ID,
		ChatID:     msg.ChatID,
		Body:       msg.Body,
		FromClient: false,
	}
}
