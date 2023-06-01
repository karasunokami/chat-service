package afcverdictsprocessor

import (
	"encoding/json"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
	"github.com/karasunokami/chat-service/internal/validator"
)

type messageStatus string

const (
	statusOk         messageStatus = "ok"
	statusSuspicious messageStatus = "suspicious"
)

type messagePayload struct {
	ChatID    types.ChatID    `json:"chatId" validate:"required"`
	MessageID types.MessageID `json:"messageId" validate:"required"`
	Status    messageStatus   `json:"status"`
}

func (m *messagePayload) Validate() error {
	return validator.Validator.Struct(m)
}

func (m *messagePayload) Valid() error {
	return m.Validate()
}

func unmarshalPayload(data []byte) (messagePayload, error) {
	p := messagePayload{}

	err := json.Unmarshal(data, &p)
	if err != nil {
		return messagePayload{}, fmt.Errorf("json unmarshal to message payload, err=%v", err)
	}

	return p, nil
}
