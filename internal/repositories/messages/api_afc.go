package messagesrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/karasunokami/chat-service/internal/types"
)

func (r *Repo) MarkAsVisibleForManager(ctx context.Context, msgID types.MessageID) error {
	err := r.db.Message(ctx).UpdateOneID(msgID).
		SetIsVisibleForManager(true).
		SetIsVisibleForClient(true).
		SetCheckedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update message, err=%v", err)
	}

	return nil
}

func (r *Repo) BlockMessage(ctx context.Context, msgID types.MessageID) error {
	err := r.db.Message(ctx).UpdateOneID(msgID).
		SetIsBlocked(true).
		SetCheckedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update message, err=%v", err)
	}

	return nil
}
