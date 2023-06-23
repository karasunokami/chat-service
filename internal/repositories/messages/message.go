package messagesrepo

import (
	"time"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/types"
)

type Message struct {
	ID               types.MessageID
	ChatID           types.ChatID
	AuthorID         types.UserID
	InitialRequestID types.RequestID
	ProblemID        types.ProblemID

	Body string

	CreatedAt time.Time

	IsVisibleForClient  bool
	IsVisibleForManager bool
	IsBlocked           bool
	IsService           bool
}

func storeMessageToRepoMessage(m *store.Message) *Message {
	return &Message{
		ID:                  m.ID,
		ChatID:              m.ChatID,
		AuthorID:            m.AuthorID,
		Body:                m.Body,
		CreatedAt:           m.CreatedAt,
		IsVisibleForClient:  m.IsVisibleForClient,
		IsVisibleForManager: m.IsVisibleForManager,
		IsBlocked:           m.IsBlocked,
		IsService:           m.IsService,
		InitialRequestID:    m.InitialRequestID,
		ProblemID:           m.ProblemID,
	}
}

func storeMessagesToRepoMessages(result []*store.Message) []Message {
	msgs := make([]Message, len(result))
	for i, m := range result {
		msgs[i] = *storeMessageToRepoMessage(m)
	}

	return msgs
}
