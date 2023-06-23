package sendmanagermessagejob

import (
	"encoding/json"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
	"github.com/karasunokami/chat-service/internal/validator"
)

type Payload struct {
	ManagerID types.UserID    `json:"managerId" validate:"required"`
	MessageID types.MessageID `json:"id" validate:"required"`
}

func (p Payload) validate() error {
	return validator.Validator.Struct(p)
}

func MarshalPayload(
	messageID types.MessageID,
	managerID types.UserID,
) (string, error) {
	p := Payload{
		MessageID: messageID,
		ManagerID: managerID,
	}

	if err := p.validate(); err != nil {
		return "", fmt.Errorf("validate job payload, err=%v", err)
	}

	d, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("json marshal, err=%v", err)
	}

	return string(d), nil
}

func UnmarshalPayload(payload string) (Payload, error) {
	var jp Payload

	err := json.Unmarshal([]byte(payload), &jp)
	if err != nil {
		return Payload{}, fmt.Errorf("unmarshal payload, err=%v", err)
	}

	return jp, nil
}
