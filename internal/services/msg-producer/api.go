package msgproducer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Message struct {
	ID         types.MessageID
	ChatID     types.ChatID
	Body       string
	FromClient bool
}

func (s *Service) ProduceMessage(ctx context.Context, msg Message) error {
	data, err := msgToJSON(msg)
	if err != nil {
		return fmt.Errorf("marshal json, err=%v", err)
	}

	if s.cipher != nil {
		data, err = s.encryptData(data)
		if err != nil {
			return fmt.Errorf("encrypt data, err=%v", err)
		}
	}

	err = s.wr.WriteMessages(ctx, kafka.Message{
		Key:   []byte(msg.ChatID.String()),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("write data to kafka writer, err=%v", err)
	}

	s.logger.Debug("Message produced", zap.Stringer("messageId", msg.ID), zap.String("body", msg.Body))

	return nil
}

func (s *Service) encryptData(data []byte) ([]byte, error) {
	nonce, err := s.nonceFactory(s.cipher.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("generage nonce, err=%v", err)
	}

	return s.cipher.Seal(nonce, nonce, data, nil), nil
}

func msgToJSON(msg Message) ([]byte, error) {
	return json.Marshal(struct {
		ID         string `json:"id"`
		ChatID     string `json:"chatId"`
		Body       string `json:"body"`
		FromClient bool   `json:"fromClient"`
	}{
		ID:         msg.ID.String(),
		ChatID:     msg.ChatID.String(),
		Body:       msg.Body,
		FromClient: msg.FromClient,
	})
}

func (s *Service) Close() error {
	err := s.wr.Close()
	if err != nil {
		return fmt.Errorf("close kafka writer, err=%v", err)
	}

	return nil
}
