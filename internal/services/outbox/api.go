package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/karasunokami/chat-service/internal/types"
)

func (s *Service) Put(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error) {
	jobID, err := s.jobsRepo.CreateJob(ctx, name, payload, availableAt)
	if err != nil {
		return types.JobIDNil, fmt.Errorf("jobs repo create job, err=%v", err)
	}

	return jobID, nil
}
