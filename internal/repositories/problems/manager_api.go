package problemsrepo

import (
	"context"
	"fmt"
	"time"

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

func (r *Repo) GetAssignedProblemID(
	ctx context.Context,
	managerID types.UserID,
	chatID types.ChatID,
) (types.ProblemID, error) {
	id, err := r.db.Problem(ctx).Query().
		Where(
			problem.ChatIDEQ(chatID),
			problem.ManagerIDEQ(managerID),
			problem.ResolvedAtIsNil(),
		).
		FirstID(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return types.ProblemIDNil, ErrNotFound
		}

		return types.ProblemIDNil, fmt.Errorf(
			"fetch problem id by manager id and chat id, err=%v",
			err,
		)
	}

	return id, nil
}

func (r *Repo) MarkProblemAsResolved(ctx context.Context, problemID types.ProblemID) error {
	err := r.db.Problem(ctx).UpdateOneID(problemID).
		SetResolvedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("update problem by chat id and manager id, err=%v", err)
	}

	return nil
}
