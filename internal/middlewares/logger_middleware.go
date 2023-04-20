package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func NewLoggerMiddleware(lg *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
			if c.Request().Method == http.MethodOptions {
				return nil
			}

			lg.Sugar().Infof(
				"latency=%v, logr_emote_ip=%v, host=%v, method=%v, uri_path=%v, request_id=%v, user_agent=%v, status=%v",
				v.Latency,
				v.RemoteIP,
				v.Host,
				v.Method,
				v.URIPath,
				v.RequestID,
				v.UserAgent,
				v.Status,
			)

			if v.Error != nil {
				lg.Sugar().Errorf("middleware error, err=%v", v.Error)
			}

			return nil
		},
	})
}
