package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// FIXME: В мидлваре, логирующей запрос, необходимо заиспользовать internal/errors.GetServerErrorCode,
// FIXME: чтобы при наличии ошибки менять status на соответствующий код.
// FIXME: Иначе в логах мы всегда будем видеть 200 OK и пропускать ошибки :)

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
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			lg.Info(
				"request",
				zap.Duration("latency", v.Latency),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("Host", v.Host),
				zap.String("method", v.Method),
				zap.String("uri_path", v.URIPath),
				zap.String("request_id", v.RequestID),
				zap.String("user_agent", v.UserAgent),
				zap.Int("status", v.Status),
			)

			if v.Error != nil {
				lg.Error("middleware error", zap.Error(v.Error))
			}

			return nil
		},
	})
}
