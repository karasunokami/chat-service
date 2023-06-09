package problemsrepo

import (
	"context"
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/message"
	"github.com/karasunokami/chat-service/internal/store/problem"
	"github.com/karasunokami/chat-service/internal/types"
)

func (r *Repo) GetProblemsWithoutManagers(ctx context.Context, limit int) ([]*store.Problem, error) {
	problems, err := r.db.Problem(ctx).Query().
		Where(
			problem.ManagerIDIsNil(),
			problem.HasMessagesWith(message.IsVisibleForManager(true)),
			problem.ResolvedAtIsNil(),
		).
		Order(store.Asc(problem.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch problems, err=%v", err)
	}

	return problems, nil
}

func (r *Repo) SetManagerToProblem(ctx context.Context, problemID types.ProblemID, managerID types.UserID) error {
	return r.db.Problem(ctx).UpdateOneID(problemID).SetManagerID(managerID).Exec(ctx)
}
