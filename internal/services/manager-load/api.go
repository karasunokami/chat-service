package managerload

import (
	"context"
	"fmt"

	"github.com/karasunokami/chat-service/internal/types"
)

func (s *Service) CanManagerTakeProblem(ctx context.Context, managerID types.UserID) (bool, error) {
	count, err := s.problemsRepo.GetManagerOpenProblemsCount(ctx, managerID)
	if err != nil {
		if err != nil {
			return false, fmt.Errorf("problems repo, get manager open problems count, err=%w", err)
		}
	}

	return count < s.maxProblemsAtTime, nil
}
