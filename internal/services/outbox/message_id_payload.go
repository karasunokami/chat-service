package outbox

import (
	"encoding/json"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
	"github.com/karasunokami/chat-service/internal/validator"
)

type MessageIDPayload struct {
	MessageID types.MessageID `json:"id" validate:"required"`
}

func (p MessageIDPayload) validate() error {
	return validator.Validator.Struct(p)
}

func MarshalMessageIDPayload(
	messageID types.MessageID,
) (string, error) {
	p := MessageIDPayload{
		MessageID: messageID,
	}

	if err := p.validate(); err != nil {
		return "", fmt.Errorf("validate job payload, err=%v", err)
	}

	d, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("json marshal MessageIDPayload, err=%v", err)
	}

	return string(d), nil
}

func UnmarshalMessageIDPayload(payload string) (MessageIDPayload, error) {
	var jp MessageIDPayload

	err := json.Unmarshal([]byte(payload), &jp)
	if err != nil {
		return MessageIDPayload{}, fmt.Errorf("unmarshal message id payload, err=%v", err)
	}

	return jp, nil
}
