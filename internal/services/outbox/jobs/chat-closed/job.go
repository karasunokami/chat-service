package chatclosed

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
	Name = "chat-closed"

	serviceName = "chat-closed-job"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/job_mock.gen.go -package=sendclientmessagejobmocks

type messageProducer interface {
	ProduceMessage(ctx context.Context, message msgproducer.Message) error
}

type messagesRepository interface {
	GetMessageByID(ctx context.Context, msgID types.MessageID) (*messagesrepo.Message, error)
}

type chatsRepository interface {
	GetClientID(ctx context.Context, chatID types.ChatID) (types.UserID, error)
}

type eventStream interface {
	Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error
}

type managerLoadService interface {
	CanManagerTakeProblem(ctx context.Context, managerID types.UserID) (bool, error)
}

//go:generate options-gen -out-filename=job_options.gen.go -from-struct=Options
type Options struct {
	msgProducer        messageProducer    `option:"mandatory" validate:"required"`
	msgRepo            messagesRepository `option:"mandatory" validate:"required"`
	chatsRepository    chatsRepository    `option:"mandatory" validate:"required"`
	eventStream        eventStream        `option:"mandatory" validate:"required"`
	managerLoadService managerLoadService `option:"mandatory" validate:"required"`
}

type Job struct {
	outbox.DefaultJob
	Options

	logger *zap.Logger
}

func New(opts Options) (*Job, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Job{
		Options: opts,
		logger:  zap.L().Named(serviceName),
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
		return fmt.Errorf("message repo, get by id, err=%v", err)
	}

	err = j.msgProducer.ProduceMessage(ctx, repoMsgToProducerMsg(msg))
	if err != nil {
		return fmt.Errorf("send message to producer, err=%v", err)
	}

	clientID, err := j.chatsRepository.GetClientID(ctx, msg.ChatID)
	if err != nil {
		return fmt.Errorf("chats repo, get client id by chat id, err=%v", err)
	}

	err = j.eventStream.Publish(ctx, clientID, eventstream.NewNewMessageEvent(
		types.NewEventID(),
		jp.RequestID,
		msg.ChatID,
		msg.ID,
		msg.CreatedAt,
		msg.Body,
		types.UserIDNil,
		msg.IsService,
	))
	if err != nil {
		return fmt.Errorf("publish message to event stream, err=%v", err)
	}

	canTakeMoreProblems, err := j.managerLoadService.CanManagerTakeProblem(ctx, jp.ManagerID)
	if err != nil {
		return fmt.Errorf("manager load svc, can manager get more problems, err=%v", err)
	}

	err = j.eventStream.Publish(ctx, jp.ManagerID, eventstream.NewChatClosedEvent(
		canTakeMoreProblems,
		msg.ChatID,
		types.NewEventID(),
		jp.RequestID,
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
