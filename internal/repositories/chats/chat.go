package chatsrepo

import (
	"time"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/types"
)

type Chat struct {
	ID        types.ChatID
	ClientID  types.UserID
	CreatedAt time.Time
}

func storeChatsToRepoChats(chats []*store.Chat) []Chat {
	chs := make([]Chat, 0, len(chats))

	for _, chat := range chats {
		chs = append(chs, storeChatToRepoChat(chat))
	}

	return chs
}

func storeChatToRepoChat(m *store.Chat) Chat {
	return Chat{
		ID:        m.ID,
		ClientID:  m.ClientID,
		CreatedAt: m.CreatedAt,
	}
}
