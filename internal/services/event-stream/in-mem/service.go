package inmemeventstream

import (
	"context"
	"fmt"
	"sync"

	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/internal/types"

	"go.uber.org/zap"
)

const serviceName = "event-stream"

type Service struct {
	mu sync.RWMutex
	wg sync.WaitGroup

	subs   map[string][]chan eventstream.Event
	logger *zap.Logger
}

func New() *Service {
	return &Service{
		subs:   make(map[string][]chan eventstream.Event),
		logger: zap.L().Named(serviceName),
	}
}

func (s *Service) Subscribe(ctx context.Context, userID types.UserID) (<-chan eventstream.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	uid := userID.String()

	if _, ex := s.subs[uid]; !ex {
		s.subs[uid] = make([]chan eventstream.Event, 0)
	}

	ch := make(chan eventstream.Event)

	s.subs[uid] = append(s.subs[uid], ch)

	s.wg.Add(1)
	go s.closeChanOnCtxDone(ctx, ch)

	s.logger.Debug("Subscriber added to event stream", zap.String("userID", uid), zap.Int("subsCount", len(s.subs[uid])))

	return ch, nil
}

func (s *Service) Publish(_ context.Context, userID types.UserID, event eventstream.Event) error {
	s.logger.Debug("Publish event called")

	if err := event.Validate(); err != nil {
		return fmt.Errorf("validate event, err=%v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	chs, ex := s.subs[userID.String()]
	if !ex || len(chs) == 0 {
		return nil
	}

	for _, c := range chs {
		s.safeSendToCh(c, event)
	}

	return nil
}

func (s *Service) Close() error {
	s.logger.Info("Stopping service...")

	s.wg.Wait()

	s.logger.Info("Stopping service done")

	return nil
}

func (s *Service) closeChanOnCtxDone(ctx context.Context, ch chan eventstream.Event) {
	s.wg.Done()

	<-ctx.Done()

	s.mu.Lock()
	defer s.mu.Unlock()

	close(ch)
}

func (s *Service) safeSendToCh(ch chan eventstream.Event, event eventstream.Event) {
	defer func() {
		recover()
	}()

	ch <- event
}
