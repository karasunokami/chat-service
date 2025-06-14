package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	jobsrepo "github.com/karasunokami/chat-service/internal/repositories/jobs"
	"github.com/karasunokami/chat-service/internal/types"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const serviceName = "outbox"

type jobsRepository interface {
	CreateJob(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error)
	FindAndReserveJob(ctx context.Context, until time.Time) (jobsrepo.Job, error)
	CreateFailedJob(ctx context.Context, name, payload, reason string) error
	DeleteJob(ctx context.Context, jobID types.JobID) error
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

//go:generate options-gen -out-filename=service_options.gen.go -from-struct=Options
type Options struct {
	workers    int            `option:"mandatory" validate:"min=1,max=32"`
	idleTime   time.Duration  `option:"mandatory" validate:"min=100ms,max=10s"`
	reserveFor time.Duration  `option:"mandatory" validate:"min=1s,max=10m"`
	jobsRepo   jobsRepository `option:"mandatory"`
	database   transactor     `option:"mandatory"`
}

type Service struct {
	Options

	lg *zap.Logger

	executeJobsCh chan jobsrepo.Job

	jobs map[string]Job
}

func New(opts Options) (*Service, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &Service{
		Options:       opts,
		jobs:          make(map[string]Job),
		lg:            zap.L().Named(serviceName),
		executeJobsCh: make(chan jobsrepo.Job),
	}, nil
}

func (s *Service) RegisterJob(job Job) error {
	return s.registerJobInService(job)
}

func (s *Service) RegisterJobs(jobs ...Job) error {
	for _, job := range jobs {
		if err := s.registerJobInService(job); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) MustRegisterJob(job Job) {
	err := s.RegisterJob(job)
	if err != nil {
		panic(err)
	}
}

func (s *Service) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	for i := 0; i < s.workers; i++ {
		eg.Go(func() error {
			s.runWorker(ctx)

			return nil
		})
	}

	eg.Go(func() error {
		s.findAndReserveJobs(ctx)

		return nil
	})

	return eg.Wait()
}

func (s *Service) findAndReserveJobs(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			j, err := s.jobsRepo.FindAndReserveJob(ctx, time.Now().Add(s.reserveFor))
			if err != nil {
				if errors.Is(err, jobsrepo.ErrNoJobs) {
					time.Sleep(s.idleTime)

					continue
				}

				s.lg.Error("Find and reserve job", zap.Error(err))

				continue
			}

			s.pushJob(ctx, j)
		}
	}
}

func (s *Service) pushJob(ctx context.Context, j jobsrepo.Job) {
	select {
	case s.executeJobsCh <- j:
	case <-ctx.Done():
		return
	}
}

func (s *Service) runWorker(ctx context.Context) {
	for {
		select {
		case j := <-s.executeJobsCh:
			err := s.handleJob(ctx, j)
			if err != nil {
				var jobFailedErr *jobFailedError
				if ok := errors.As(err, &jobFailedErr); ok {
					s.moveJobToDLQ(ctx, j, jobFailedErr.getReason())

					continue
				}

				s.lg.Error("Handle job, err=%v", zap.String("job", j.Name), zap.Error(err))
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) handleJob(ctx context.Context, j jobsrepo.Job) error {
	serviceJob, ex := s.getServiceJob(j.Name)
	if !ex {
		return newJobFailedError(fmt.Sprintf("job with name %s not registered", j.Name))
	}

	err := s.executeJob(ctx, serviceJob, j.Payload)
	if err != nil {
		if serviceJob.MaxAttempts() <= j.Attempts {
			return newJobFailedError(
				fmt.Sprintf("max attempts for job exceeded, job=%s, max_attempts=%d", j.Name, serviceJob.MaxAttempts()),
			)
		}

		return fmt.Errorf("execute job, err=%w", err)
	}

	err = s.jobsRepo.DeleteJob(ctx, j.ID)
	if err != nil {
		return fmt.Errorf("delete job from db, err=%v", err)
	}

	return nil
}

func (s *Service) executeJob(ctx context.Context, serviceJob Job, payload string) error {
	ctx, cancel := context.WithTimeout(ctx, serviceJob.ExecutionTimeout())
	defer cancel()

	return serviceJob.Handle(ctx, payload)
}

func (s *Service) moveJobToDLQ(ctx context.Context, j jobsrepo.Job, reason string) {
	err := s.database.RunInTx(ctx, func(ctx context.Context) error {
		err := s.jobsRepo.CreateFailedJob(ctx, j.Name, j.Payload, reason)
		if err != nil {
			return fmt.Errorf("create failed job, err=%v", err)
		}

		err = s.jobsRepo.DeleteJob(ctx, j.ID)
		if err != nil {
			return fmt.Errorf("delete job, err=%v", err)
		}

		return nil
	})
	if err != nil {
		s.lg.Error("Move job to dlq", zap.Error(err))
	}
}

func (s *Service) registerJobInService(job Job) error {
	if _, ex := s.jobs[job.Name()]; ex {
		return ErrJobAlreadyExists
	}

	s.jobs[job.Name()] = job

	return nil
}

func (s *Service) getServiceJob(name string) (Job, bool) {
	j, ex := s.jobs[name]
	jobCopy := j

	return jobCopy, ex
}
