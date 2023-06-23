package managerscheduler

import (
	"context"
	"fmt"
	"time"

	managerassignedtoproblemjob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/manager-assigned-to-problem"
	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/types"

	"go.uber.org/zap"
)

const serviceName = "manager-scheduler"

type problemsRepo interface {
	GetProblemsWithoutManagers(ctx context.Context, limit int) ([]*store.Problem, error)
	SetManagerToProblem(ctx context.Context, problemID types.ProblemID, managerID types.UserID) error
}

type outboxService interface {
	Put(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error)
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

type managersPool interface {
	Get(ctx context.Context) (types.UserID, error)
	Put(ctx context.Context, managerID types.UserID) error
	Size() int
}

//go:generate options-gen -out-filename=service_options.gen.go -from-struct=Options
type Options struct {
	period time.Duration `option:"mandatory" validate:"min=100ms,max=1m"`

	managersPool  managersPool  `option:"mandatory" validate:"required"`
	outboxService outboxService `option:"mandatory" validate:"required"`
	problemsRepo  problemsRepo  `option:"mandatory" validate:"required"`
	transactor    transactor    `option:"mandatory" validate:"required"`
}

type Service struct {
	Options
	logger *zap.Logger
}

func New(opts Options) (*Service, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Service{
		Options: opts,
		logger:  zap.L().Named(serviceName),
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			managersAvailableCount := s.managersPool.Size()
			if managersAvailableCount <= 0 {
				continue
			}

			problems, err := s.problemsRepo.GetProblemsWithoutManagers(ctx, managersAvailableCount)
			if err != nil {
				s.logger.Error("Fetch problems without managers", zap.Error(err))

				continue
			}

			for _, problem := range problems {
				mngID, err := s.managersPool.Get(ctx)
				if err != nil {
					return fmt.Errorf("get manager from managers pool, err=%v", err)
				}

				err = s.setManagerToProblem(ctx, mngID, problem)
				if err != nil {
					err := s.managersPool.Put(ctx, mngID)
					if err != nil {
						s.logger.Warn("Return manager to managers pool", zap.Error(err))
					}

					s.logger.Error("Set manager to problem", zap.Error(err))

					continue
				}
			}
		}
	}
}

func (s *Service) setManagerToProblem(ctx context.Context, mngID types.UserID, problem *store.Problem) error {
	return s.transactor.RunInTx(ctx, func(ctx context.Context) error {
		err := s.problemsRepo.SetManagerToProblem(ctx, problem.ID, mngID)
		if err != nil {
			return fmt.Errorf("set manager to problem, err=%v", err)
		}

		payload, err := managerassignedtoproblemjob.MarshalPayload(
			mngID,
			problem.ID,
		)
		if err != nil {
			return fmt.Errorf("marshal manager assigned to problem job payload, err=%v", err)
		}

		_, err = s.outboxService.Put(
			ctx,
			managerassignedtoproblemjob.Name,
			payload,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("put job to outbox service, err=%v", err)
		}

		return nil
	})
}
