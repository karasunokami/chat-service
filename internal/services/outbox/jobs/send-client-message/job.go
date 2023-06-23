package sendclientmessagejob

import (
	"context"
	"fmt"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	msgproducer "github.com/karasunokami/chat-service/internal/services/msg-producer"
	"github.com/karasunokami/chat-service/internal/services/outbox"
	"github.com/karasunokami/chat-service/internal/types"

	"go.uber.org/zap"
)

const (
	Name = "send-client-message"

	serviceName = "send-client-message-job"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/job_mock.gen.go -package=sendclientmessagejobmocks

type messageProducer interface {
	ProduceMessage(ctx context.Context, message msgproducer.Message) error
}

type messageRepository interface {
	GetMessageByID(ctx context.Context, msgID types.MessageID) (*messagesrepo.Message, error)
}

type eventStream interface {
	Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error
}

//go:generate options-gen -out-filename=job_options.gen.go -from-struct=Options
type Options struct {
	msgProducer messageProducer   `option:"mandatory" validate:"required"`
	msgRepo     messageRepository `option:"mandatory" validate:"required"`
	eventStream eventStream       `option:"mandatory" validate:"required"`
}

type Job struct {
	outbox.DefaultJob
	msgProducer messageProducer
	msgRepo     messageRepository
	eventStream eventStream

	logger *zap.Logger
}

func New(opts Options) (*Job, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Job{
		msgProducer: opts.msgProducer,
		msgRepo:     opts.msgRepo,
		eventStream: opts.eventStream,

		logger: zap.L().Named(serviceName),
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

	err = j.msgProducer.ProduceMessage(ctx, repoMsgToProducerMsg(msg))
	if err != nil {
		return fmt.Errorf("send message to producer, err=%v", err)
	}

	err = j.eventStream.Publish(ctx, msg.AuthorID, eventstream.NewNewMessageEvent(
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

	j.logger.Debug("Publish message to event stream", zap.Any("msgID", msg.ID))

	return nil
}

func repoMsgToProducerMsg(msg *messagesrepo.Message) msgproducer.Message {
	return msgproducer.Message{
		ID:         msg.ID,
		ChatID:     msg.ChatID,
		Body:       msg.Body,
		FromClient: !msg.IsService,
	}
}
