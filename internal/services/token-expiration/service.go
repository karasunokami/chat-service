package tokenexpiration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Service struct {
	mu sync.Mutex

	timers map[string]*time.Timer
}

func New() *Service {
	return &Service{
		timers: make(map[string]*time.Timer, 0),
	}
}

func (s *Service) NewExpireContext(ctx context.Context, id string, deadline time.Time) (context.Context, error) {
	duration := time.Until(deadline)
	if duration <= 0 {
		return nil, fmt.Errorf("negative duration")
	}

	ctx, cancel := context.WithCancel(ctx)

	go s.cancelAfter(ctx, cancel, id, duration)

	return ctx, nil
}

func (s *Service) Extend(id string, deadline time.Time) error {
	duration := time.Until(deadline)
	if duration <= 0 {
		return fmt.Errorf("negative duration")
	}

	t, ok := s.getTimer(id)
	if !ok {
		return fmt.Errorf("timer with id %s not found", id)
	}

	t.Reset(duration)

	return nil
}

func (s *Service) cancelAfter(ctx context.Context, cancel context.CancelFunc, id string, duration time.Duration) {
	defer cancel()
	defer s.removeTimer(id)

	t := time.NewTimer(duration)
	defer t.Stop()

	s.addTimer(id, t)

	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

func (s *Service) removeTimer(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.timers, id)
}

func (s *Service) addTimer(id string, t *time.Timer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.timers[id] = t
}

func (s *Service) getTimer(id string) (*time.Timer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.timers[id]

	return t, ok
}
