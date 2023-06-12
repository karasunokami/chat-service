package messagesrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/chat"
	"github.com/karasunokami/chat-service/internal/store/message"
	"github.com/karasunokami/chat-service/internal/store/predicate"
	"github.com/karasunokami/chat-service/internal/store/problem"
	"github.com/karasunokami/chat-service/internal/types"
)

var (
	ErrInvalidPageSize        = errors.New("invalid page size")
	ErrInvalidCursor          = errors.New("invalid cursor")
	ErrEmptyPageSizeAndCursor = errors.New("empty page size and cursor")
)

type Cursor struct {
	LastCreatedAt time.Time
	PageSize      int
}

// GetClientChatMessages returns Nth page of messages in the chat for client side.
func (r *Repo) GetClientChatMessages(
	ctx context.Context,
	clientID types.UserID,
	pageSize int,
	cursor *Cursor,
) ([]Message, *Cursor, error) {
	err := validateParams(pageSize, cursor)
	if err != nil {
		return nil, nil, fmt.Errorf("validate params, err=%w", err)
	}

	limit := pageSize
	if cursor != nil {
		limit = cursor.PageSize
	}

	query := r.buildMessagesQuery(ctx, limit, cursor)
	query = query.Where(message.HasChatWith(chat.ClientID(clientID)))

	msgs, err := query.All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("query messages, err=%v", err)
	}

	if len(msgs) <= limit {
		return storeMessagesToRepoMessages(msgs), nil, nil
	}

	return storeMessagesToRepoMessages(msgs[:limit]), &Cursor{
		LastCreatedAt: msgs[limit-1].CreatedAt,
		PageSize:      limit,
	}, nil
}

// GetManagerChatMessages returns Nth page of messages in the chat for client side.
func (r *Repo) GetManagerChatMessages(
	ctx context.Context,
	chatID types.ChatID,
	managerID types.UserID,
	pageSize int,
	cursor *Cursor,
) ([]Message, *Cursor, error) {
	err := validateParams(pageSize, cursor)
	if err != nil {
		return nil, nil, fmt.Errorf("validate params, err=%w", err)
	}

	limit := pageSize
	if cursor != nil {
		limit = cursor.PageSize
	}

	query := r.buildMessagesQuery(ctx, limit, cursor)

	query = query.Where(
		message.ChatIDEQ(chatID),
		message.IsVisibleForManager(true),
		message.HasChatWith(
			chat.HasProblemsWith(
				problem.ManagerIDEQ(managerID),
			),
		),
	)

	msgs, err := query.All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("query messages, err=%v", err)
	}

	if len(msgs) <= limit {
		return storeMessagesToRepoMessages(msgs), nil, nil
	}

	return storeMessagesToRepoMessages(msgs[:limit]), &Cursor{
		LastCreatedAt: msgs[limit-1].CreatedAt,
		PageSize:      limit,
	}, nil
}

func (r *Repo) buildMessagesQuery(ctx context.Context, limit int, cursor *Cursor) *store.MessageQuery {
	predicates := []predicate.Message{
		message.IsVisibleForClient(true),
	}

	if cursor != nil {
		predicates = append(predicates, message.CreatedAtLT(cursor.LastCreatedAt))
	}

	return r.db.Message(ctx).Query().Where(predicates...).Order(store.Desc(message.FieldCreatedAt)).Limit(limit + 1)
}

func validateParams(pageSize int, cursor *Cursor) error {
	if pageSize == 0 && cursor == nil {
		return ErrEmptyPageSizeAndCursor
	}

	err := validatePageSize(pageSize)
	if err != nil {
		return fmt.Errorf("validate page size, err=%w", err)
	}

	err = validateCursor(cursor)
	if err != nil {
		return fmt.Errorf("validate cursor page size, err=%w, err=%w", ErrInvalidCursor, err) // 1.20 is sick!
	}

	return nil
}

func validateCursor(cursor *Cursor) error {
	if cursor == nil {
		return nil
	}

	err := validatePageSize(cursor.PageSize)
	if err != nil {
		return fmt.Errorf("validate cursor page size")
	}

	if cursor.LastCreatedAt.Second() == 0 {
		return errors.New("validate cursor createdAt, create at must be provided")
	}

	return nil
}

func validatePageSize(size int) error {
	if size != 0 && (size < 10 || size > 100) {
		return fmt.Errorf("page size must be between 10 and 100, err:=%w", ErrInvalidPageSize)
	}

	return nil
}
