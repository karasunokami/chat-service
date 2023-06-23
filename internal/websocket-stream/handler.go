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

type tokenExpiration interface {
	NewExpireContext(ctx context.Context, userID string, deadline time.Time) (context.Context, error)
}

//go:generate options-gen -out-filename=handler_options.gen.go -from-struct=Options
type Options struct {
	pingPeriod time.Duration `default:"3s" validate:"omitempty,min=100ms,max=30s"`

	logger          *zap.Logger     `option:"mandatory" validate:"required"`
	eventStream     eventStream     `option:"mandatory" validate:"required"`
	eventAdapter    EventAdapter    `option:"mandatory" validate:"required"`
	eventWriter     EventWriter     `option:"mandatory" validate:"required"`
	upgrader        Upgrader        `option:"mandatory" validate:"required"`
	shutdownCh      <-chan struct{} `option:"mandatory" validate:"required"`
	tokenExpiration tokenExpiration `option:"mandatory" validate:"required"`
}

type HTTPHandler struct {
	Options
	pingPeriod time.Duration
	pongWait   time.Duration
}

func NewHTTPHandler(opts Options) (*HTTPHandler, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	opts.logger = opts.logger.Named(serviceName)

	return &HTTPHandler{
		Options:    opts,
		pingPeriod: opts.pingPeriod,
		pongWait:   pongWait(opts.pingPeriod),
	}, nil
}

func (h *HTTPHandler) Serve(eCtx echo.Context) error {
	ws, err := h.upgrader.Upgrade(eCtx.Response(), eCtx.Request(), nil)
	if err != nil {
		return fmt.Errorf("upgrade request, err=%v", err)
	}

	exp := middlewares.MustExpiresAt(eCtx)
	uid := middlewares.MustUserID(eCtx)

	ctxWithExpiration, err := h.tokenExpiration.NewExpireContext(eCtx.Request().Context(), uid.String(), exp)
	if err != nil {
		return fmt.Errorf("create context with token expiration timeout, err=%v", err)
	}

	wsCtx, cancel := context.WithCancel(ctxWithExpiration)
	defer cancel()

	wsCloser := newWsCloser(h.logger, ws)

	events, err := h.eventStream.Subscribe(wsCtx, uid)
	if err != nil {
		h.logger.Error("Cannot subscribe for events", zap.Error(err))
		wsCloser.Close(gorillaws.CloseInternalServerErr)

		return nil
	}

	eg, egCtx := errgroup.WithContext(wsCtx)

	eg.Go(func() error { return h.writeLoop(egCtx, ws, events) })
	eg.Go(func() error { return h.readLoop(egCtx, ws) })
	eg.Go(func() error {
		select {
		case <-egCtx.Done():
		case <-h.shutdownCh:
			wsCloser.Close(gorillaws.CloseNormalClosure)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		if gorillaws.IsUnexpectedCloseError(err, gorillaws.CloseNormalClosure, gorillaws.CloseNoStatusReceived) {
			h.logger.Error("unexpected error", zap.Error(err))
			wsCloser.Close(gorillaws.CloseInternalServerErr)

			return err
		}
	}

	wsCloser.Close(gorillaws.CloseNormalClosure)

	return nil
}

// readLoop listen PONGs.
func (h *HTTPHandler) readLoop(ctx context.Context, ws Websocket) error {
	ws.SetPongHandler(func(string) error {
		return ws.SetReadDeadline(time.Now().Add(h.pongWait))
	})

	if err := ws.SetReadDeadline(time.Now().Add(h.pongWait)); err != nil {
		return fmt.Errorf("set first read deadline, err=%v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, _, err := ws.NextReader()
			if gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure) {
				return nil
			}

			if err != nil {
				return fmt.Errorf("get next reader, err=%w", err)
			}
		}
	}
}

// writeLoop listen events and writes them into Websocket.
func (h *HTTPHandler) writeLoop(ctx context.Context, ws Websocket, events <-chan eventstream.Event) error {
	pingTicker := time.NewTicker(h.pingPeriod)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-pingTicker.C:
			if err := ws.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				return fmt.Errorf("set write deadline, err=%w", err)
			}

			if err := ws.WriteMessage(gorillaws.PingMessage, nil); err != nil {
				return fmt.Errorf("write ping message, err=%w", err)
			}
		case event, ok := <-events:
			if !ok {
				return errors.New("events stream was closed")
			}

			adapted, err := h.eventAdapter.Adapt(event)
			if err != nil {
				h.logger.With(zap.Error(err)).Error("cannot adapt event to out stream")

				continue
			}

			if err := ws.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				return fmt.Errorf("set write deadline, err=%w", err)
			}

			wr, err := ws.NextWriter(gorillaws.TextMessage)
			if err != nil {
				return fmt.Errorf("get next writer, err=%w", err)
			}

			if err := h.eventWriter.Write(adapted, wr); err != nil {
				return fmt.Errorf("write data to connection, err=%w", err)
			}

			if err := wr.Close(); err != nil {
				return fmt.Errorf("flush writer, err=%w", err)
			}
		}
	}
}

func pongWait(ping time.Duration) time.Duration {
	return ping * 3 / 2
}
