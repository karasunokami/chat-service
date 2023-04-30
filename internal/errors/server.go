package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

const defaultInternalError = "something went wrong"

// ServerError is used to return custom error codes to client.
type ServerError struct {
	Code    int
	Message string
	cause   error
}

func NewServerError(code int, msg string, err error) *ServerError {
	return &ServerError{
		Code:    code,
		Message: msg,
		cause:   err,
	}
}

func (s *ServerError) Error() string {
	return fmt.Sprintf("%s: %v", s.Message, s.cause)
}

func (s *ServerError) Unwrap() error {
	return s.cause
}

func GetServerErrorCode(err error) int {
	code, _, _ := ProcessServerError(err)
	return code
}

// ProcessServerError tries to retrieve from given error it's code, message and some details.
// For example, that fields can be used to build error response for client.
func ProcessServerError(err error) (code int, msg string, details string) {
	var srvError *ServerError
	if ok := errors.As(err, &srvError); ok {
		return srvError.Code, srvError.Message, srvError.Error()
	}

	var echoHTTPError *echo.HTTPError
	if ok := errors.As(err, &echoHTTPError); ok {
		mes, _ := (echoHTTPError.Message).(string)

		return echoHTTPError.Code, mes, echoHTTPError.Error()
	}

	if err != nil {
		details = err.Error()
	}

	return http.StatusInternalServerError, defaultInternalError, details
}
