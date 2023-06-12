package websocketstream

import (
	"context"
	"errors"
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
	serviceName  = "websocket-stream"
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

	opts.logger = opts.logger.Named(serviceName)

	return &HTTPHandler{Options: opts}, nil
}

func (h *HTTPHandler) Serve(eCtx echo.Context) error {
	ws, err := h.upgrader.Upgrade(eCtx.Response().Writer, eCtx.Request(), eCtx.Response().Header())
	if err != nil {
		return fmt.Errorf("upgrade connection, err=%v", err)
	}

	userID := middlewares.MustUserID(eCtx)

	ctx, cancel := context.WithCancel(eCtx.Request().Context())
	defer cancel()

	eventsCh, err := h.eventStream.Subscribe(ctx, userID)
	if err != nil {
		return fmt.Errorf("event stream subscribe, err=%v", err)
	}

	eg := errgroup.Group{}

	eg.Go(func() error { return h.readLoop(ws) })
	eg.Go(func() error { return h.writeLoop(eventsCh, ws) })

	h.logger.Debug("WS handler started for user", zap.String("userID", userID.String()))

	err = eg.Wait()
	if err != nil {
		h.logger.Error("WS closed with error", zap.Error(err))
	}

	return nil
}

// readLoop listen PONGs.
func (h *HTTPHandler) readLoop(ws Websocket) error {
	h.logger.Debug("Starting reading loop...")
	defer func() {
		wsCloser := newWsCloser(h.logger, ws)
		wsCloser.Close(gorillaws.CloseNormalClosure)

		h.logger.Debug("Reading loop done")
	}()

	err := ws.SetReadDeadline(time.Now().Add(h.pingPeriod + time.Second))
	if err != nil {
		h.logger.Error("Set read deadline", zap.Error(err))

		return fmt.Errorf("set read deadline, err=%v", err)
	}

	ws.SetPongHandler(func(string) error {
		err := ws.SetReadDeadline(time.Now().Add(h.pingPeriod + time.Second))
		if err != nil {
			h.logger.Error("Set read deadline", zap.Error(err))
		}

		return nil
	})

	for {
		select {
		case <-h.shutdownCh:
			return nil
		default:
			err := h.safeReadFromWS(ws)
			if err != nil {
				if gorillaws.IsUnexpectedCloseError(
					err,
					gorillaws.CloseGoingAway,
					gorillaws.CloseAbnormalClosure,
					gorillaws.CloseNoStatusReceived,
				) {
					h.logger.Error("Unexpected error on get next reader", zap.Error(err))
				}

				return nil
			}
		}
	}
}

func (h *HTTPHandler) safeReadFromWS(ws Websocket) error {
	defer func() {
		if e := recover(); e != nil {
			h.logger.Error("Recovered from ws next reader call", zap.Any("recovered", e))
		}
	}()

	_, _, err := ws.NextReader()
	if err != nil {
		return err
	}

	return nil
}

// writeLoop listen events and writes them into Websocket.
func (h *HTTPHandler) writeLoop(events <-chan eventstream.Event, ws Websocket) error {
	h.logger.Debug("Starting write loop...")

	t := time.NewTicker(h.pingPeriod)
	defer func() {
		t.Stop()
		wsCloser := newWsCloser(h.logger, ws)
		wsCloser.Close(gorillaws.CloseNormalClosure)

		h.logger.Debug("Writing loop done")
	}()

	for {
		select {
		case <-h.shutdownCh:
			return nil

		case event, opened := <-events:
			h.logger.Debug("New event received from events channel", zap.Any("event", event))

			if !opened {
				return nil
			}

			err := h.write(ws, event)
			if err != nil {
				h.logger.Error("Write to WS", zap.Error(err))

				return fmt.Errorf("write to ws, err=%v", err)
			}

		case <-t.C:
			err := ws.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err != nil {
				h.logger.Error("SetWriteDeadline", zap.Error(err))

				return fmt.Errorf("set write deadline, err=%v", err)
			}

			if err := ws.WriteMessage(gorillaws.PingMessage, nil); err != nil {
				if !errors.Is(err, gorillaws.ErrCloseSent) {
					h.logger.Error("Write ping message", zap.Error(err))

					return err
				}
			}
		}
	}
}

func (h *HTTPHandler) write(ws Websocket, event eventstream.Event) error {
	err := ws.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return fmt.Errorf("set write deadline, err=%v", err)
	}

	w, err := ws.NextWriter(gorillaws.TextMessage)
	if err != nil {
		return fmt.Errorf("write loop, get next writer, err=%v", err)
	}
	defer func() {
		err := w.Close()
		if err != nil {
			h.logger.Error("Close writer", zap.Error(err))
		}
	}()

	adapted, err := h.eventAdapter.Adapt(event)
	if err != nil {
		return fmt.Errorf("adapt event with type %T, err=%v", event, err)
	}

	err = h.eventWriter.Write(adapted, w)
	if err != nil {
		return fmt.Errorf("write loop, write event, err=%v", err)
	}

	h.logger.Debug("Write event to ws done", zap.Any("adapted_event", adapted))

	return nil
}
