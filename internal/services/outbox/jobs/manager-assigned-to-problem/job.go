package managerassignedtoproblemjob

import (
	"context"
	"fmt"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	msgproducer "github.com/karasunokami/chat-service/internal/services/msg-producer"
	"github.com/karasunokami/chat-service/internal/services/outbox"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/job_mock.gen.go -package=clientmessagesentjobmocks

const (
	Name = "manager-assigned-to-problem"

	ServiceMessageTpl = "Manager %s will answer you"
)

//go:generate options-gen -out-filename=job_options.gen.go -from-struct=Options
type Options struct {
	msgProducer messageProducer   `option:"mandatory" validate:"required"`
	eventStream eventStream       `option:"mandatory" validate:"required"`
	msgRepo     messageRepository `option:"mandatory" validate:"required"`
}

type eventStream interface {
	Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error
}

type messageProducer interface {
	ProduceMessage(ctx context.Context, message msgproducer.Message) error
}

type messageRepository interface {
	GetFirstProblemMessage(ctx context.Context, problemID types.ProblemID) (*messagesrepo.Message, error)
	CreateService(
		ctx context.Context,
		problemID types.ProblemID,
		chatID types.ChatID,
		msgBody string,
	) (*messagesrepo.Message, error)
}

type Job struct {
	outbox.DefaultJob
	eventStream eventStream
	msgRepo     messageRepository
	msgProducer messageProducer
}

func New(opts Options) (*Job, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Job{
		eventStream: opts.eventStream,
		msgRepo:     opts.msgRepo,
		msgProducer: opts.msgProducer,
	}, nil
}

func (j *Job) Name() string {
	return Name
}

func (j *Job) Handle(ctx context.Context, payload string) error {
	pl, err := unmarshalPayload(payload)
	if err != nil {
		return fmt.Errorf("unmarshal payload, err=%v", err)
	}

	body := fmt.Sprintf(ServiceMessageTpl, pl.ManagerID)

	msg, err := j.msgRepo.GetFirstProblemMessage(ctx, pl.ProblemID)
	if err != nil {
		return fmt.Errorf("add service message to problem, err=%v", err)
	}

	serviceMsg, err := j.msgRepo.CreateService(
		ctx,
		pl.ProblemID,
		msg.ChatID,
		body,
	)
	if err != nil {
		return fmt.Errorf("msg repo create service message, err=%v", err)
	}

	err = j.msgProducer.ProduceMessage(ctx, repoMsgToProducerMsg(serviceMsg))
	if err != nil {
		return fmt.Errorf("send message to producer, err=%v", err)
	}

	err = j.eventStream.Publish(ctx, msg.AuthorID, eventstream.NewNewMessageEvent(
		types.NewEventID(),
		msg.InitialRequestID,
		serviceMsg.ChatID,
		serviceMsg.ID,
		serviceMsg.CreatedAt,
		serviceMsg.Body,
		types.UserIDNil,
		serviceMsg.IsService,
	))
	if err != nil {
		return fmt.Errorf("publish message to event stream, err=%v", err)
	}

	err = j.eventStream.Publish(ctx, pl.ManagerID, eventstream.NewNewChatEvent(
		pl.CanTakeMoreProblems,
		types.NewEventID(),
		serviceMsg.InitialRequestID,
		serviceMsg.ChatID,
		msg.AuthorID,
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
		FromClient: !msg.IsService,
	}
}
