package problemsrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/problem"
	"github.com/karasunokami/chat-service/internal/types"
)

var ErrNotFound = errors.New("problem not found")

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
		return 0, fmt.Errorf("get active problems count by manager id, err=%v", err)
	}

	return count, nil
}

func (r *Repo) GetManagerID(ctx context.Context, problemID types.ProblemID) (types.UserID, error) {
	p, err := r.db.Problem(ctx).Query().
		Where(problem.IDEQ(problemID)).
		Select(problem.FieldManagerID).
		First(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return types.UserIDNil, ErrNotFound
		}

		return types.UserIDNil, fmt.Errorf("fetch manager id by problem id, err=%v", err)
	}

	return p.ManagerID, nil
}
