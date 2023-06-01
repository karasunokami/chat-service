package sendclientmessagejob

import (
	"encoding/json"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
)

type jobPayload struct {
	MessageID types.MessageID `json:"id"`
}

func MarshalPayload(messageID types.MessageID) (string, error) {
	if err := messageID.Validate(); err != nil {
		return "", fmt.Errorf("validate message, err=%v", err)
	}
	p := jobPayload{MessageID: messageID}

	d, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("json marshal jobPayload, err=%v", err)
	}

	return string(d), nil
}

func unmarshalPayload(payload string) (jobPayload, error) {
	var jp jobPayload

	err := json.Unmarshal([]byte(payload), &jp)
	if err != nil {
		return jobPayload{}, fmt.Errorf("unmarshal payload to job payload, err=%v", err)
	}

	return jp, nil
}
