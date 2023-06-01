package inmemmanagerpool

import (
	"context"
	"sync"

	managerpool "github.com/karasunokami/chat-service/internal/services/manager-pool"
	"github.com/karasunokami/chat-service/internal/types"
)

const (
	serviceName = "manager-pool"
	managersMax = 1000
)

type Service struct {
	mu sync.RWMutex

	managers []types.UserID
}

func New() *Service {
	s := &Service{
		managers: make([]types.UserID, 0, managersMax),
	}

	return s
}

func (s *Service) Close() error {
	return nil
}

func (s *Service) Get(_ context.Context) (types.UserID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ml := len(s.managers)

	if ml == 0 {
		return types.UserIDNil, managerpool.ErrNoAvailableManagers
	}

	m := s.managers[0]
	s.managers = s.managers[1:]

	return m, nil
}

func (s *Service) Put(_ context.Context, managerID types.UserID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, manager := range s.managers {
		if manager.Matches(managerID) {
			return nil
		}
	}

	s.managers = append(s.managers, managerID)

	return nil
}

func (s *Service) Contains(_ context.Context, managerID types.UserID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, manager := range s.managers {
		if manager.Matches(managerID) {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.managers)
}
