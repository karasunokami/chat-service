package middlewares

import (
	"net/http"

	"github.com/karasunokami/chat-service/internal/errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func NewLoggerMiddleware(lg *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().Method == http.MethodOptions
		},
		LogLatency:   true,
		LogRemoteIP:  true,
		LogHost:      true,
		LogMethod:    true,
		LogURIPath:   true,
		LogRequestID: true,
		LogUserAgent: true,
		LogStatus:    true,
		LogError:     true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			status := v.Status

			if v.Error != nil {
				lg.Error("middleware error", zap.Error(v.Error))
				status = errors.GetServerErrorCode(v.Error)
			}

			lg.Info(
				"request",
				zap.Duration("latency", v.Latency),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("host", v.Host),
				zap.String("method", v.Method),
				zap.String("uri_path", v.URIPath),
				zap.String("request_id", v.RequestID),
				zap.String("user_agent", v.UserAgent),
				zap.Int("status", status),
			)

			return nil
		},
	})
}
