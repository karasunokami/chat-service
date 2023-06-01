package websocketstream

import (
	"context"
	"fmt"
	"time"

	"github.com/karasunokami/chat-service/internal/middlewares"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/internal/types"

	gorillaws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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

	userID := middlewares.MustUserID(eCtx)

	ctx := context.Background()
	eventsCh, err := h.eventStream.Subscribe(ctx, userID)
	if err != nil {
		return fmt.Errorf("event stream subscribe, err=%v", err)
	}

	cws := newConcurrentWS(h.logger, ws)

	eg, _ := errgroup.WithContext(ctx)

	eg.Go(func() error { return h.listenShutdown(cws) })
	eg.Go(func() error { return h.readLoop(cws) })
	eg.Go(func() error { return h.writeLoop(eventsCh, cws) })
	eg.Go(func() error { return h.writePingsLoop(cws) })

	h.logger.Debug("WS handler started for user", zap.String("userID", userID.String()))

	return eg.Wait()
}

// readLoop listen PONGs.
func (h *HTTPHandler) readLoop(cws *concurrentWS) error {
	for {
		msgType, _, err := cws.NextReader()
		if err != nil {
			return fmt.Errorf("ws next reader, err=%v", err)
		}

		if msgType == gorillaws.PongMessage {
			err = cws.SetReadDeadline(time.Now().Add(writeTimeout))
			if err != nil {
				h.logger.Error("ws set read deadline", zap.Error(err))
			}
		}
	}
}

// writeLoop listen events and writes them into Websocket.
func (h *HTTPHandler) writeLoop(events <-chan eventstream.Event, cws *concurrentWS) error {
	for ev := range events {
		err := cws.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err != nil {
			return fmt.Errorf("set write deadline, err=%v", err)
		}

		w, err := cws.NextWriter(gorillaws.TextMessage)
		if err != nil {
			return fmt.Errorf("write loop, get next writer, err=%v", err)
		}

		adapted, err := h.eventAdapter.Adapt(ev)
		if err != nil {
			return fmt.Errorf("adapt event with type %T, err=%v", ev, err)
		}

		err = h.eventWriter.Write(adapted, w)
		if err != nil {
			h.logger.Error("write loop, write event", zap.Error(err))
		}

		err = cws.CloseWriter(w)
		if err != nil {
			h.logger.Error("write loop, close writer", zap.Error(err))
		}
	}

	return nil
}

func (h *HTTPHandler) writePingsLoop(cws *concurrentWS) error {
	t := time.NewTicker(h.pingPeriod)
	defer t.Stop()

	for range t.C {
		err := cws.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err != nil {
			return fmt.Errorf("set write deadline, err=%v", err)
		}

		err = cws.WriteControl(gorillaws.PingMessage, []byte{}, time.Now().Add(writeTimeout))
		if err != nil {
			return fmt.Errorf("ws get next writer, err=%v", err)
		}
	}

	return nil
}

func (h *HTTPHandler) listenShutdown(cws *concurrentWS) error {
	<-h.shutdownCh

	cws.Close(gorillaws.CloseNormalClosure)

	return nil
}
