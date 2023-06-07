package afcverdictsprocessor

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"time"

	clientmessageblockedjob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/client-message-blocked"
	clientmessagesentjob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/client-message-sent"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang-jwt/jwt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	serviceName = "afc-verdicts-processor"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/service_mocks.gen.go -package=afcverdictsprocessormocks

type messagesRepository interface {
	MarkAsVisibleForManager(ctx context.Context, msgID types.MessageID) error
	BlockMessage(ctx context.Context, msgID types.MessageID) error
}

type outboxService interface {
	Put(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error)
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

//go:generate options-gen -out-filename=service_options.gen.go -from-struct=Options
type Options struct {
	backoffInitialInterval time.Duration `default:"100ms" validate:"min=50ms,max=1s"`
	backoffMaxElapsedTime  time.Duration `default:"5s" validate:"min=500ms,max=1m"`
	backoffExpFactor       float64       `default:"2" validate:"min=1.1,max=5"`

	brokers          []string `option:"mandatory" validate:"min=1"`
	consumers        int      `option:"mandatory" validate:"min=1,max=16"`
	consumerGroup    string   `option:"mandatory" validate:"required"`
	verdictsTopic    string   `option:"mandatory" validate:"required"`
	verdictsSignKey  string
	processBatchSize int

	readerFactory KafkaReaderFactory `option:"mandatory" validate:"required"`
	dlqWriter     KafkaDLQWriter     `option:"mandatory" validate:"required"`

	txtor   transactor         `option:"mandatory" validate:"required"`
	msgRepo messagesRepository `option:"mandatory" validate:"required"`
	outBox  outboxService      `option:"mandatory" validate:"required"`
}

type Service struct {
	Options

	signKey *rsa.PublicKey
	logger  *zap.Logger
}

func New(opts Options) (*Service, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%w", err)
	}

	s := &Service{Options: opts}

	if err := s.init(); err != nil {
		return nil, fmt.Errorf("init service, err=%w", err)
	}

	return s, nil
}

func (s *Service) Run(ctx context.Context) error {
	defer func() {
		err := s.dlqWriter.Close()
		if err != nil {
			s.logger.Error("Close dlq writer", zap.Error(err))
		}
	}()

	eg, egCtx := errgroup.WithContext(ctx)

	for i := 0; i < s.Options.consumers; i++ {
		eg.Go(func() error {
			err := s.consumerLoop(egCtx)
			if err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, io.EOF) {
					s.logger.Error("Consumer loop returned error", zap.Error(err))

					return err
				}
			}

			return nil
		})
	}

	return eg.Wait()
}

func (s *Service) init() error {
	if s.verdictsSignKey != "" {
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(s.verdictsSignKey))
		if err != nil {
			return fmt.Errorf("parse verdict sign key, err=%w", err)
		}

		s.signKey = key
	}

	s.logger = zap.L().Named(serviceName)

	return nil
}

func (s *Service) consumerLoop(ctx context.Context) error {
	r := s.readerFactory(s.brokers, s.consumerGroup, s.verdictsTopic)
	defer func() {
		err := r.Close()
		if err != nil {
			s.logger.Error("Close consumer reader", zap.Error(err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			m, err := r.FetchMessage(ctx)
			if err != nil {
				return fmt.Errorf("fetch message, err=%w", err)
			}

			s.logger.Debug("Message fetched", zap.Any("message", m))

			err = s.handleMessage(ctx, m)
			if err != nil {
				s.logger.Debug("Handle message error", zap.Error(err))

				s.writeMessageToDlq(ctx, m, err.Error())
			}

			err = r.CommitMessages(ctx, m)
			if err != nil {
				s.logger.Error("Commit message", zap.Error(err))
			}
		}
	}
}

func (s *Service) getDelay(lastDelay time.Duration) time.Duration {
	return lastDelay * time.Duration(s.backoffExpFactor)
}

func (s *Service) handleMessage(ctx context.Context, m kafka.Message) error {
	mp, err := s.parseMessage(m.Value)
	if err != nil {
		return fmt.Errorf("parse message, err=%w", err)
	}

	err = mp.Validate()
	if err != nil {
		return fmt.Errorf("validate message payload, err=%w", err)
	}

	err = s.handleWithRetries(ctx, mp)
	if err != nil {
		return fmt.Errorf("handle with retries, err=%w", err)
	}

	return nil
}

func (s *Service) handleWithRetries(ctx context.Context, mp messagePayload) error {
	var (
		lastError error
		lastDelay = s.backoffInitialInterval
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if lastError != nil {
				delay := s.getDelay(lastDelay)

				if delay > s.backoffMaxElapsedTime {
					return fmt.Errorf("handle message timeout, err=%w", lastError)
				}

				s.logger.Debug("Sleep before next handle retry", zap.Duration("duration", delay))

				lastDelay = delay
				time.Sleep(delay)
			}

			err := s.handleMessageByStatus(ctx, mp)
			if err != nil {
				s.logger.Debug("Handle message by status returned error", zap.Error(err))

				lastError = multierr.Append(lastError, err)

				if he, converted := convertToHandleErr(err); converted && he.IsTemporary() {
					// retry
					continue
				}

				return err
			}

			return nil
		}
	}
}

func (s *Service) handleMessageByStatus(ctx context.Context, mp messagePayload) error {
	switch mp.Status {
	case statusOk:
		if err := s.handleMessageOk(ctx, mp.MessageID); err != nil {
			return newHandleMessageError(err, true)
		}
	case statusSuspicious:
		if err := s.handleMessageSuspicious(ctx, mp.MessageID); err != nil {
			return newHandleMessageError(err, true)
		}
	default:
		return newHandleMessageError(fmt.Errorf("unknown message status, status=%v", mp.Status), false)
	}

	return nil
}

func (s *Service) handleMessageOk(ctx context.Context, msgID types.MessageID) error {
	err := s.msgRepo.MarkAsVisibleForManager(ctx, msgID)
	if err != nil {
		return fmt.Errorf("msg repo mark as visible for manager, err=%v", err)
	}

	payload, err := clientmessagesentjob.MarshalPayload(msgID)
	if err != nil {
		return fmt.Errorf("marshal client message sent job payload, err=%v", err)
	}

	_, err = s.outBox.Put(ctx, clientmessagesentjob.Name, payload, time.Now())
	if err != nil {
		return fmt.Errorf("outbox svc put, err=%v", err)
	}

	return nil
}

func (s *Service) handleMessageSuspicious(ctx context.Context, msgID types.MessageID) error {
	err := s.msgRepo.BlockMessage(ctx, msgID)
	if err != nil {
		return fmt.Errorf("msg repo block message, err=%v", err)
	}

	payload, err := clientmessageblockedjob.MarshalPayload(msgID)
	if err != nil {
		return fmt.Errorf("marshal client message blocked job payload, err=%v", err)
	}

	_, err = s.outBox.Put(ctx, clientmessageblockedjob.Name, payload, time.Now())
	if err != nil {
		return fmt.Errorf("outbox svc put, err=%v", err)
	}

	return nil
}

func (s *Service) writeMessageToDlq(ctx context.Context, m kafka.Message, lastErrorText string) {
	dlqMessage := kafka.Message{
		Key:   m.Key,
		Value: m.Value,
		Headers: []protocol.Header{
			{
				Key:   "LAST_ERROR",
				Value: []byte(lastErrorText),
			},
			{
				Key:   "ORIGINAL_PARTITION",
				Value: []byte{byte(m.Partition)},
			},
		},
	}

	err := s.dlqWriter.WriteMessages(ctx, dlqMessage)
	if err != nil {
		s.logger.Error("Write message to DLQ", zap.Error(err), zap.Any("message", dlqMessage))
	}
}

func (s *Service) parseMessage(data []byte) (messagePayload, error) {
	if s.signKey != nil {
		token, err := jwt.ParseWithClaims(string(data), &messagePayload{}, func(token *jwt.Token) (interface{}, error) {
			return s.signKey, nil
		})
		if err != nil {
			return messagePayload{}, fmt.Errorf("jwt parse with claims, err=%v", err)
		}

		return *token.Claims.(*messagePayload), nil
	}

	mp, err := unmarshalPayload(data)
	if err != nil {
		return messagePayload{}, fmt.Errorf("unmarshal payload, err=%v", err)
	}

	return mp, nil
}
