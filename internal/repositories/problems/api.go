package problemsrepo

import (
	"context"
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/problem"
	"github.com/karasunokami/chat-service/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, chatID types.ChatID) (types.ProblemID, error) {
	id, err := r.db.Problem(ctx).Query().Where(
		problem.ChatID(chatID),
		problem.ResolvedAtIsNil(),
	).FirstID(ctx)
	if nil == err {
		return id, nil
	}

	if !store.IsNotFound(err) {
		return types.ProblemID{}, fmt.Errorf("query problem id by chatID=%s, err=%v", chatID, err)
	}

	newProblem, err := r.db.Problem(ctx).Create().SetChatID(chatID).Save(ctx)
	if err != nil {
		return types.ProblemID{}, fmt.Errorf("create new problem, err=%v", err)
	}

	return newProblem.ID, nil
}

func (r *Repo) GetManagerOpenProblemsCount(ctx context.Context, managerID types.UserID) (int, error) {
	count, err := r.db.Problem(ctx).Query().Where(
		problem.ManagerID(managerID),
		problem.ResolvedAtIsNil(),
	).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("get problems count by manager id, err=%v", err)
	}

	return count, nil
}
