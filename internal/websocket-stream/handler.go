package websocketstream

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/karasunokami/chat-service/internal/middlewares"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/internal/types"

	gorillaws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	writeTimeout = time.Second
)

type eventStream interface {
	Subscribe(ctx context.Context, userID types.UserID) (<-chan eventstream.Event, error)
}

//go:generate options-gen -out-filename=handler_options.gen.go -from-struct=Options
type Options struct {
	pingPeriod time.Duration `default:"3s" validate:"omitempty,min=100ms,max=30s"`

	logger       *zap.Logger     `option:"mandatory" validate:"required"`
	eventStream  eventStream     `option:"mandatory" validate:"required"`
	eventAdapter EventAdapter    `option:"mandatory" validate:"required"`
	eventWriter  EventWriter     `option:"mandatory" validate:"required"`
	upgrader     Upgrader        `option:"mandatory" validate:"required"`
	shutdownCh   <-chan struct{} `option:"mandatory" validate:"required"`
}

type HTTPHandler struct {
	Options

	mu sync.Mutex
}

func NewHTTPHandler(opts Options) (*HTTPHandler, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	return &HTTPHandler{Options: opts}, nil
}

func (h *HTTPHandler) Serve(eCtx echo.Context) error {
	ws, err := h.upgrader.Upgrade(eCtx.Response().Writer, eCtx.Request(), eCtx.Response().Header())
	if err != nil {
		return fmt.Errorf("upgrade connection, err=%v", err)
	}

	go h.listenShutdown(ws)

	err = h.readLoop(ws)
	if err != nil {
		return fmt.Errorf("run read loop, err=%v", err)
	}

	userID := middlewares.MustUserID(eCtx)

	ctx := context.Background()
	eventsCh, err := h.eventStream.Subscribe(ctx, userID)
	if err != nil {
		return fmt.Errorf("event stream subscribe, err=%v", err)
	}

	err = h.writeLoop(eventsCh, ws)
	if err != nil {
		return fmt.Errorf("run write loop, err=%v", err)
	}

	err = h.writePingsLoop(ws)
	if err != nil {
		return fmt.Errorf("run write pings loop, err=%v", err)
	}

	h.logger.Debug("WS handler started for user", zap.String("userID", userID.String()))

	return nil
}

// readLoop listen PONGs.
func (h *HTTPHandler) readLoop(ws Websocket) error {
	go func() {
		for {
			msgType, _, err := ws.NextReader()
			if err != nil {
				h.logger.Error("ws next reader", zap.Error(err))

				return
			}

			if msgType == gorillaws.PongMessage {
				err = h.setReadDeadline(writeTimeout, ws)
				if err != nil {
					h.logger.Error("ws set read deadline", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

// writeLoop listen events and writes them into Websocket.
func (h *HTTPHandler) writeLoop(events <-chan eventstream.Event, ws Websocket) error {
	go func() {
		for ev := range events {
			err := h.setWriteDeadline(writeTimeout, ws)
			if err != nil {
				h.logger.Error("ws, set write timeout", zap.Error(err))
			}

			w, err := ws.NextWriter(gorillaws.TextMessage)
			if err != nil {
				h.logger.Error("write loop, get next writer", zap.Error(err))

				return
			}

			adapted, err := h.eventAdapter.Adapt(ev)
			if err != nil {
				h.logger.Error(fmt.Sprintf("adapt event with type %T", ev), zap.Error(err))

				return
			}

			err = h.eventWriter.Write(adapted, w)
			if err != nil {
				h.logger.Error("write loop, write event", zap.Error(err))
			}

			err = h.closeWriter(w)
			if err != nil {
				h.logger.Error("write loop, close writer", zap.Error(err))
			}
		}
	}()

	return nil
}

func (h *HTTPHandler) writePingsLoop(ws Websocket) error {
	go func() {
		t := time.NewTicker(h.pingPeriod)
		defer t.Stop()

		for range t.C {
			err := ws.WriteControl(gorillaws.PingMessage, []byte{}, time.Now().Add(writeTimeout))
			if err != nil {
				h.logger.Error("ws, write control", zap.Error(err))

				return
			}

			err = h.setWriteDeadline(writeTimeout, ws)
			if err != nil {
				h.logger.Error("set write timeout", zap.Error(err))
			}
		}
	}()

	return nil
}

func (h *HTTPHandler) listenShutdown(ws Websocket) {
	closer := newWsCloser(h.logger, ws)

	<-h.shutdownCh

	closer.Close(gorillaws.CloseNormalClosure)
}

func (h *HTTPHandler) setWriteDeadline(d time.Duration, ws Websocket) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return ws.SetWriteDeadline(time.Now().Add(d))
}

func (h *HTTPHandler) setReadDeadline(d time.Duration, ws Websocket) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return ws.SetReadDeadline(time.Now().Add(d))
}

func (h *HTTPHandler) closeWriter(w io.WriteCloser) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return w.Close()
}
