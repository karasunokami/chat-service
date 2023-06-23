package tokenexpiration_test

import (
	"context"
	"testing"
	"time"

	tokenexpiration "github.com/karasunokami/chat-service/internal/services/token-expiration"

	"github.com/stretchr/testify/require"
)

func TestExpire(t *testing.T) {
	s := tokenexpiration.New()

	ctx := context.Background()
	userID := "1"
	deadline := time.Now().Add(time.Millisecond * 20)

	expireContext, err := s.NewExpireContext(ctx, userID, deadline)
	require.NoError(t, err)

	waitContextClose(expireContext, t, time.Millisecond*30)
}

func TestExtendExpire(t *testing.T) {
	s := tokenexpiration.New()

	ctx := context.Background()
	userID := "1"
	deadline := time.Now().Add(time.Millisecond * 100)

	expireContext, err := s.NewExpireContext(ctx, userID, deadline)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 20)

	err = s.Extend(userID, time.Now().Add(time.Millisecond*100))
	require.NoError(t, err)

	err = s.Extend(userID, time.Now().Add(time.Millisecond*100))
	require.NoError(t, err)

	waitContextNotClose(expireContext, t, time.Millisecond*50)
	waitContextClose(expireContext, t, time.Millisecond*301)
}

func waitContextClose(ctx context.Context, t *testing.T, timeout time.Duration) {
	t.Helper()

	select {
	case <-ctx.Done():
	case <-time.After(timeout):
		t.Error("Context no closed")
		t.Fail()
	}
}

func waitContextNotClose(ctx context.Context, t *testing.T, timeout time.Duration) {
	t.Helper()

	select {
	case <-ctx.Done():
		t.Error("Context closed")
		t.Fail()
	case <-time.After(timeout):
	}
}
