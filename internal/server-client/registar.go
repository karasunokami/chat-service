package serverclient

import (
	clientv1 "github.com/karasunokami/chat-service/internal/server-client/v1"

	oapimdlwr "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
)

func NewHandlersRegistrar(
	v1Swagger *openapi3.T,
	v1Handlers clientv1.ServerInterface,
	httpErrorHandler echo.HTTPErrorHandler,
) func(e *echo.Echo) {
	return func(e *echo.Echo) {
		v1 := e.Group("v1", oapimdlwr.OapiRequestValidatorWithOptions(v1Swagger, &oapimdlwr.Options{
			Options: openapi3filter.Options{
				ExcludeRequestBody:  false,
				ExcludeResponseBody: true,
				AuthenticationFunc:  openapi3filter.NoopAuthenticationFunc,
			},
		}))
		clientv1.RegisterHandlers(v1, v1Handlers)

		e.HTTPErrorHandler = httpErrorHandler
	}
}
