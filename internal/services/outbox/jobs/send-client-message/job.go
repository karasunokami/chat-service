package sendclientmessagejob

import (
	"context"
	"fmt"
	"time"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	msgproducer "github.com/karasunokami/chat-service/internal/services/msg-producer"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/job_mock.gen.go -package=sendclientmessagejobmocks

const (
	Name = "send-client-message"

	executionTimeout = time.Second * 3
	maxAttempts      = 5
)

type messageProducer interface {
	ProduceMessage(ctx context.Context, message msgproducer.Message) error
}

type messageRepository interface {
	GetMessageByID(ctx context.Context, msgID types.MessageID) (*messagesrepo.Message, error)
}

//go:generate options-gen -out-filename=job_options.gen.go -from-struct=Options
type Options struct {
	msgProducer messageProducer   `option:"mandatory" validate:"required"`
	msgRepo     messageRepository `option:"mandatory" validate:"required"`
}

type Job struct {
	msgProducer messageProducer
	msgRepo     messageRepository
}

func New(opts Options) (*Job, error) {
	err := opts.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Job{
		msgProducer: opts.msgProducer,
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

	err = j.msgProducer.ProduceMessage(ctx, repoMsgToProducerMsg(msg))
	if err != nil {
		return fmt.Errorf("send message to producer, err=%v", err)
	}

	return nil
}

func (j *Job) ExecutionTimeout() time.Duration {
	return executionTimeout
}

func (j *Job) MaxAttempts() int {
	return maxAttempts
}

func repoMsgToProducerMsg(msg *messagesrepo.Message) msgproducer.Message {
	return msgproducer.Message{
		ID:         msg.ID,
		ChatID:     msg.ChatID,
		Body:       msg.Body,
		FromClient: !msg.IsService,
	}
}
