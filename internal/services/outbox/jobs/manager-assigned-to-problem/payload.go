package managerassignedtoproblemjob

import (
	"encoding/json"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
	"github.com/karasunokami/chat-service/internal/validator"
)

type jobPayload struct {
	ManagerID           types.UserID    `json:"managerId" validate:"required"`
	ProblemID           types.ProblemID `json:"problemId" validate:"required"`
	CanTakeMoreProblems bool            `json:"canTakeMoreProblems"`
}

func (p jobPayload) validate() error {
	return validator.Validator.Struct(p)
}

func MarshalPayload(
	managerID types.UserID,
	problemID types.ProblemID,
	canTakeMoreProblems bool,
) (string, error) {
	p := jobPayload{
		ManagerID:           managerID,
		ProblemID:           problemID,
		CanTakeMoreProblems: canTakeMoreProblems,
	}

	if err := p.validate(); err != nil {
		return "", fmt.Errorf("validate job payload, err=%v", err)
	}

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
		return jobPayload{}, fmt.Errorf("unmarshal job payload, err=%v", err)
	}

	return jp, nil
}
