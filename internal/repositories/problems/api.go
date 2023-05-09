package problemsrepo

import (
	"context"
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/predicate"
	"github.com/karasunokami/chat-service/internal/store/problem"
	"github.com/karasunokami/chat-service/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, chatID types.ChatID) (types.ProblemID, error) {
	id, err := r.db.Problem(ctx).Query().Where([]predicate.Problem{
		problem.ChatID(chatID),
		problem.ResolvedAtIsNil(),
	}...).FirstID(ctx)
	if err != nil && !store.IsNotFound(err) {
		return types.ProblemID{}, fmt.Errorf("query problem id by chatID=%s, err=%v", chatID, err)
	}

	if !id.IsZero() {
		return id, nil
	}

	newProblem, err := r.db.Problem(ctx).Create().SetChatID(chatID).Save(ctx)
	if err != nil {
		return types.ProblemID{}, fmt.Errorf("create new problem, err=%v", err)
	}

	return newProblem.ID, nil
}
