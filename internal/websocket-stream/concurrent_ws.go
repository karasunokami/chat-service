package websocketstream

import (
	"io"
	"sync"
	"time"

	gorillaws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	closeDeadline = 5 * time.Second
	graceTimeout  = 1 * time.Second
)

type concurrentWS struct {
	mu     sync.Mutex
	once   sync.Once
	ws     Websocket
	logger *zap.Logger
}

func newConcurrentWS(logger *zap.Logger, ws Websocket) *concurrentWS {
	return &concurrentWS{
		logger: logger,
		ws:     ws,
	}
}

func (c *concurrentWS) SetWriteDeadline(deadline time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ws.SetWriteDeadline(deadline)
}

func (c *concurrentWS) SetReadDeadline(deadline time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ws.SetReadDeadline(deadline)
}

func (c *concurrentWS) CloseWriter(w io.WriteCloser) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return w.Close()
}

func (c *concurrentWS) NextWriter(messageType int) (io.WriteCloser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ws.NextWriter(messageType)
}

func (c *concurrentWS) NextReader() (messageType int, r io.Reader, err error) {
	return c.ws.NextReader()
}

func (c *concurrentWS) WriteControl(messageType int, data []byte, deadline time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ws.WriteControl(messageType, data, deadline)
}

func (c *concurrentWS) Close(code int) {
	c.once.Do(func() {
		c.logger.Debug("close connection")

		_ = c.ws.WriteControl(
			gorillaws.CloseMessage,
			gorillaws.FormatCloseMessage(code, ""),
			time.Now().Add(closeDeadline),
		)

		time.Sleep(graceTimeout)
		_ = c.ws.Close()
	})
}
