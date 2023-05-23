package msgproducer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"

	"github.com/segmentio/kafka-go"
)

type Message struct {
	ID         types.MessageID `json:"id"`
	ChatID     types.ChatID    `json:"chatId"`
	Body       string          `json:"body,omitempty"`
	FromClient bool            `json:"fromClient,omitempty"`
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

	return nil
}

func (s *Service) encryptData(data []byte) ([]byte, error) {
	nonce, err := s.nonceFactory(s.cipher.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("generage nonce, err=%v", err)
	}

	encrypted := s.cipher.Seal(nil, nonce, data, nil)

	return append(nonce, encrypted...), nil
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
