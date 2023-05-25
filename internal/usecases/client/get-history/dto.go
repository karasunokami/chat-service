package gethistory

import (
	"encoding/json"
	"errors"
	"time"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	"github.com/karasunokami/chat-service/internal/types"
	"github.com/karasunokami/chat-service/internal/validator"
)

type Request struct {
	ID       types.RequestID `validate:"required"`
	ClientID types.UserID    `validate:"required"`
	PageSize int             `validate:"omitempty,gte=10,lte=100"`
	Cursor   string          `validate:"omitempty,base64url"`
}

func (r Request) Validate() error {
	if r.PageSize == 0 && r.Cursor == "" {
		return errors.New("page size or cursor must be provided")
	}

	if r.PageSize != 0 && r.Cursor != "" {
		return errors.New("page size or cursor must be provided")
	}

	return validator.Validator.Struct(r)
}

type Response struct {
	NextCursor string    `json:"next"`
	Messages   []Message `json:"messages"`
}

type Message struct {
	ID         types.MessageID
	AuthorID   types.UserID
	Body       string
	CreatedAt  time.Time
	IsReceived bool
	IsBlocked  bool
	IsService  bool
}

func (m Message) MarshalJSON() ([]byte, error) {
	t := struct {
		ID         types.MessageID `json:"id"`
		AuthorID   *types.UserID   `json:"authorId,omitempty" `
		Body       string          `json:"body"`
		CreatedAt  time.Time       `json:"createdAt"`
		IsReceived bool            `json:"isReceived"`
		IsBlocked  bool            `json:"isBlocked"`
		IsService  bool            `json:"isService"`
	}{
		ID:         m.ID,
		Body:       m.Body,
		CreatedAt:  m.CreatedAt,
		IsReceived: m.IsReceived,
		IsBlocked:  m.IsBlocked,
		IsService:  m.IsService,
	}

	if !m.AuthorID.IsZero() {
		t.AuthorID = &m.AuthorID
	}

	return json.Marshal(t)
}

func adoptMessages(messages []messagesrepo.Message) []Message {
	msgs := make([]Message, 0, len(messages))

	for _, message := range messages {
		msgs = append(msgs, adoptMessage(message))
	}

	return msgs
}

func adoptMessage(m messagesrepo.Message) Message {
	return Message{
		ID:         m.ID,
		AuthorID:   m.AuthorID,
		Body:       m.Body,
		CreatedAt:  m.CreatedAt,
		IsReceived: m.IsVisibleForManager && !m.IsBlocked,
		IsBlocked:  m.IsBlocked,
		IsService:  m.IsService,
	}
}
