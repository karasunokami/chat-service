package jobsrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/job"
	"github.com/karasunokami/chat-service/internal/store/predicate"
	"github.com/karasunokami/chat-service/internal/types"
)

var ErrNoJobs = errors.New("no jobs found")

type Job struct {
	ID       types.JobID
	Name     string
	Payload  string
	Attempts int
}

func (r *Repo) FindAndReserveJob(ctx context.Context, until time.Time) (Job, error) {
	var rj Job

	err := r.db.RunInTx(ctx, func(ctx context.Context) error {
		now := time.Now()
		predicates := []predicate.Job{
			job.Or([]predicate.Job{job.ReservedUntilIsNil(), job.ReservedUntilLT(now)}...),
			job.AvailableAtLTE(now),
		}

		j, err := r.db.Job(ctx).Query().Where(predicates...).ForUpdate().First(ctx)
		if err != nil {
			if store.IsNotFound(err) {
				return ErrNoJobs
			}

			return fmt.Errorf("select first, err=%v", err)
		}

		j, err = r.db.Job(ctx).UpdateOne(j).SetAttempts(j.Attempts + 1).SetReservedUntil(until).Save(ctx)
		if err != nil {
			return fmt.Errorf("update one, err=%v", err)
		}

		rj = Job{
			ID:       j.ID,
			Name:     j.Name,
			Payload:  j.Payload,
			Attempts: j.Attempts,
		}

		return nil
	})
	if err != nil {
		return Job{}, fmt.Errorf("run in tx, err=%w", err)
	}

	return rj, nil
}

func (r *Repo) CreateJob(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error) {
	j, err := r.db.Job(ctx).
		Create().
		SetName(name).
		SetPayload(payload).
		SetAvailableAt(availableAt).
		SetReservedUntil(availableAt).
		Save(ctx)
	if err != nil {
		return types.JobIDNil, fmt.Errorf("save in db, err=%v", err)
	}

	return j.ID, nil
}

func (r *Repo) CreateFailedJob(ctx context.Context, name, payload, reason string) error {
	_, err := r.db.FailedJob(ctx).Create().SetName(name).SetPayload(payload).SetReason(reason).Save(ctx)
	if err != nil {
		return fmt.Errorf("save in db, err=%v", err)
	}

	return nil
}

func (r *Repo) DeleteJob(ctx context.Context, jobID types.JobID) error {
	err := r.db.Job(ctx).DeleteOneID(jobID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete in db, err=%v", err)
	}

	return nil
}
