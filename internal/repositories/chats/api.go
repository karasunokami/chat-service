package chatsrepo

import (
	"context"
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/chat"
	"github.com/karasunokami/chat-service/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, userID types.UserID) (types.ChatID, error) {
	id, err := r.db.Chat(ctx).Query().Where(chat.ClientID(userID)).FirstID(ctx)
	if err != nil && !store.IsNotFound(err) {
		return types.ChatID{}, fmt.Errorf("query chat id by clientId=%s, err=%v", userID, err)
	}

	if !id.IsZero() {
		return id, nil
	}

	newChat, err := r.db.Chat(ctx).Create().SetClientID(userID).Save(ctx)
	if err != nil {
		return types.ChatID{}, fmt.Errorf("create new chat, err=%v", err)
	}

	return newChat.ID, nil
}
