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

	query := r.buildMessagesQuery(ctx, limit, cursor, clientID)

	msgs, err := query.All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("query messages, err=%v", err)
	}

	repoMessages := storeMessagesToRepoMessages(msgs)

	// if count of selected messages lower than limit we know
	// that there are no more messages in db
	if len(msgs) < limit {
		return repoMessages, nil, nil
	}

	crs, err := r.createNextCursor(ctx, clientID, msgs[limit-1].CreatedAt, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("create next cursor, err=%v", err)
	}

	return repoMessages, crs, nil
}

func (r *Repo) createNextCursor(
	ctx context.Context,
	clientID types.UserID,
	lastMessageCreatedAt time.Time,
	limit int,
) (*Cursor, error) {
	nextCursor := &Cursor{
		LastCreatedAt: lastMessageCreatedAt,
		PageSize:      limit,
	}

	exists, err := r.buildMessagesQuery(ctx, limit, nextCursor, clientID).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("query messages exists, err=%v", err)
	}

	if !exists {
		return (*Cursor)(nil), nil
	}

	return nextCursor, nil
}

func (r *Repo) buildMessagesQuery(ctx context.Context, limit int, cursor *Cursor, clientID types.UserID) *store.MessageQuery {
	predicates := []predicate.Message{
		message.HasChatWith(chat.ClientID(clientID)),
		message.IsVisibleForClient(true),
	}

	if cursor != nil {
		predicates = append(predicates, message.CreatedAtLT(cursor.LastCreatedAt))
	}

	return r.db.Message(ctx).Query().Where(predicates...).Order(store.Desc(message.FieldCreatedAt)).Limit(limit)
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
