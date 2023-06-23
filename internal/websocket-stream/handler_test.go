package websocketstream_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/karasunokami/chat-service/internal/logger"
	"github.com/karasunokami/chat-service/internal/middlewares"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	tokenexpiration "github.com/karasunokami/chat-service/internal/services/token-expiration"
	"github.com/karasunokami/chat-service/internal/types"
	websocketstream "github.com/karasunokami/chat-service/internal/websocket-stream"

	"github.com/golang/mock/gomock"
	gorillaws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	logger.MustInit(logger.NewOptions("debug"))
}

const (
	eventsNum     = 3
	eventInterval = time.Second

	pingInterval = eventInterval / 4

	origin = "http://localhost"

	headerSecWsProtocol = "Sec-WebSocket-Protocol"
	secWsProtocol       = "chat-service-protocol.test"
)

func TestHTTPHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uid := types.NewUserID()
	eventsCh := make(chan eventstream.Event)
	shutdownCh := make(chan struct{})

	log := zap.L().Named("TestHTTPHandler")

	h, err := newHTTPHandler(uid, eventsCh, shutdownCh)
	require.NoError(t, err)

	e := echo.New()
	e.GET("/ws", middlewares.AuthWith(uid)(h.Serve))
	s := httptest.NewServer(e)

	u := url.URL{Scheme: "ws", Host: s.Listener.Addr().String(), Path: "/ws"}
	t.Log(u.String())

	header := http.Header{}
	header.Add(echo.HeaderOrigin, origin)
	header.Add(headerSecWsProtocol, secWsProtocol)

	c, resp, err := gorillaws.DefaultDialer.DialContext(ctx, u.String(), header)
	require.NoError(t, err)
	assert.Equal(t, secWsProtocol, resp.Header.Get(headerSecWsProtocol))
	defer func() {
		require.NoError(t, c.Close())
		require.NoError(t, resp.Body.Close())
	}()

	var pings int
	{
		c.SetPingHandler(nil) // Hack to set default ping handler.
		defaultPingHandler := c.PingHandler()

		c.SetPingHandler(func(appData string) error {
			pings++
			log.Debug("new ping received, send pong")
			return defaultPingHandler(appData)
		})
	}

	events := make([]eventstream.Event, 0, eventsNum)
	for i := 0; i < eventsNum; i++ {
		events = append(events, new(eventstream.MessageSentEvent))
	}

	go func() {
		for _, e := range events {
			eventsCh <- e
			time.Sleep(eventInterval)
		}
	}()

	receivedEvents := make([]*eventstream.MessageSentEvent, 0, len(events))
	for {
		var event eventstream.MessageSentEvent
		if err := c.ReadJSON(&event); err != nil {
			if gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure) {
				break
			}
			require.NoError(t, err)
		}

		receivedEvents = append(receivedEvents, &event)
		log.Debug("new event received")

		if len(receivedEvents) == len(events) {
			close(shutdownCh)
		}
	}

	t.Run("event stream is working properly", func(t *testing.T) {
		require.Len(t, receivedEvents, len(events))
		for i, e := range receivedEvents {
			assert.Equal(t, events[i], e, "i = %d", i)
		}
	})

	t.Run("ping-pong mechanism is working properly", func(t *testing.T) {
		t.Logf("pings: %d", pings)
		assert.InDelta(t, (eventsNum-1)*4, pings, 1.)
	})

	t.Run("shutdown is working properly", func(t *testing.T) {
		_, _, err := c.NextReader()
		assert.Error(t, err)
		assert.True(t, gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure))
	})
}

func TestCloseWsOnTokenExpired(t *testing.T) {
	const tokenLiveTime = time.Second * 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uid := types.NewUserID()

	h, err := newHTTPHandler(uid, make(chan eventstream.Event), make(chan struct{}))
	require.NoError(t, err)

	e := echo.New()
	e.GET("/ws", middlewares.AuthWithExp(uid, time.Now().Add(tokenLiveTime).Unix())(h.Serve))
	s := httptest.NewServer(e)

	u := url.URL{Scheme: "ws", Host: s.Listener.Addr().String(), Path: "/ws"}
	t.Log(u.String())

	header := http.Header{}
	header.Add(echo.HeaderOrigin, origin)
	header.Add(headerSecWsProtocol, secWsProtocol)

	c, resp, err := gorillaws.DefaultDialer.DialContext(ctx, u.String(), header)
	require.NoError(t, err)
	assert.Equal(t, secWsProtocol, resp.Header.Get(headerSecWsProtocol))
	defer func() {
		require.NoError(t, c.Close())
		require.NoError(t, resp.Body.Close())
	}()

	t.Run("ws closed by token expiration", func(t *testing.T) {
		<-time.After(tokenLiveTime + time.Millisecond*5)

		_, _, err := c.NextReader()
		if gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure) {
			return
		}

		t.Fatalf("ws not closed by token expiraiton")
	})

	t.Run("shutdown is working properly", func(t *testing.T) {
		_, _, err := c.NextReader()
		assert.Error(t, err)
		assert.True(t, gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure))
	})
}

func newHTTPHandler(
	uid types.UserID,
	eventsCh chan eventstream.Event,
	shutdownCh chan struct{},
) (*websocketstream.HTTPHandler, error) {
	return websocketstream.NewHTTPHandler(websocketstream.NewOptions(
		zap.L(),
		eventStreamMock{uid: uid, ch: eventsCh},
		eventAdapter{},
		websocketstream.JSONEventWriter{},
		websocketstream.NewUpgrader([]string{origin}, secWsProtocol),
		shutdownCh,
		tokenexpiration.New(),
		websocketstream.WithPingPeriod(pingInterval),
	))
}

type eventStreamMock struct {
	ch  chan eventstream.Event
	uid types.UserID
}

func (e eventStreamMock) Subscribe(_ context.Context, userID types.UserID) (<-chan eventstream.Event, error) {
	if e.uid != userID {
		return nil, fmt.Errorf("unexpected user: %v != %v", e.uid, userID)
	}
	return e.ch, nil
}

type eventAdapter struct{}

func (eventAdapter) Adapt(event eventstream.Event) (any, error) {
	return event, nil
}
