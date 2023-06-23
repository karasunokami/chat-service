package chatsrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/chat"
	"github.com/karasunokami/chat-service/internal/store/problem"
	"github.com/karasunokami/chat-service/internal/types"
)

var ErrNotFound = errors.New("chat not found")

func (r *Repo) CreateIfNotExists(ctx context.Context, userID types.UserID) (types.ChatID, error) {
	chatID, err := r.db.Chat(ctx).Create().
		SetClientID(userID).
		OnConflictColumns(chat.FieldClientID).Ignore().
		ID(ctx)
	if err != nil {
		return types.ChatIDNil, fmt.Errorf("create new chat: %v", err)
	}

	return chatID, nil
}

func (r *Repo) GetManagerOpened(ctx context.Context, managerID types.UserID) ([]Chat, error) {
	chats, err := r.db.Chat(ctx).Query().
		Where(chat.HasProblemsWith(
			problem.ManagerIDEQ(managerID),
			problem.ResolvedAtIsNil(),
		)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query chats, err=%v", err)
	}

	return storeChatsToRepoChats(chats), nil
}

func (r *Repo) GetClientID(ctx context.Context, chatID types.ChatID) (types.UserID, error) {
	c, err := r.db.Chat(ctx).Query().
		Where(chat.IDEQ(chatID)).
		Select(chat.FieldClientID).
		First(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return types.UserIDNil, ErrNotFound
		}

		return types.UserIDNil, fmt.Errorf("fetch manager id by problem id, err=%v", err)
	}

	return c.ClientID, nil
}
