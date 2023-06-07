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

type (
	clientChMap map[int]chan eventstream.Event
	subsMap     map[string]clientChMap
)

type Service struct {
	mu sync.RWMutex
	wg sync.WaitGroup

	subs   subsMap
	logger *zap.Logger
}

func New() *Service {
	return &Service{
		subs:   make(subsMap),
		logger: zap.L().Named(serviceName),
	}
}

func (s *Service) Subscribe(ctx context.Context, userID types.UserID) (<-chan eventstream.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	uid := userID.String()

	if _, ex := s.subs[uid]; !ex {
		s.subs[uid] = make(clientChMap, 0)
	}

	ch := make(chan eventstream.Event)

	ind := len(s.subs[uid])
	s.subs[uid][ind] = ch

	s.wg.Add(1)
	go s.closeChanOnCtxDone(ctx, ch, uid, ind)

	s.logger.Debug("Subscriber added to event stream", zap.String("userID", uid), zap.Int("subsCount", len(s.subs[uid])))

	return ch, nil
}

func (s *Service) Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("validate event, err=%v", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	chs, ex := s.subs[userID.String()]
	if !ex || len(chs) == 0 {
		return nil
	}

	s.logger.Debug("Publishing message for client", zap.String("clientID", userID.String()), zap.Any("event", event))

	for ind, c := range chs {
		err := s.safeSendToCh(ctx, c, event)
		if err != nil {
			delete(chs, ind)
		}
	}

	return nil
}

func (s *Service) Close() error {
	s.logger.Info("Stopping service...")

	s.wg.Wait()

	s.logger.Info("Stopping service done")

	return nil
}

func (s *Service) closeChanOnCtxDone(ctx context.Context, ch chan eventstream.Event, userID string, ind int) {
	defer s.wg.Done()

	<-ctx.Done()

	s.logger.Debug("Stopping event stream subscriber...")

	s.mu.Lock()
	defer s.mu.Unlock()

	close(ch)
	delete(s.subs[userID], ind)

	s.logger.Debug("Stopping event stream subscriber done")
}

func (s *Service) safeSendToCh(ctx context.Context, ch chan eventstream.Event, event eventstream.Event) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if re, ok := e.(error); ok {
				err = re
			}
		}
	}()

	select {
	case <-ctx.Done():
	case ch <- event:
	}

	return
}
