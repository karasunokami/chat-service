package clientv1

import (
	"errors"
	"net/http"

	internalerrors "github.com/karasunokami/chat-service/internal/errors"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/client/get-history"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/client/send-message"
)

const defaultHandleErrorMessage = "cannot handle something"

func newHandleError(err error, code int) error {
	return internalerrors.NewServerError(code, defaultHandleErrorMessage, err)
}

func getErrorCode(err error) int {
	switch {
	case errors.Is(err, gethistory.ErrInvalidRequest),
		errors.Is(err, gethistory.ErrInvalidCursor),
		errors.Is(err, sendmessage.ErrInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, sendmessage.ErrChatNotCreated):
		return int(ErrorCodeCreateChatError)
	case errors.Is(err, sendmessage.ErrProblemNotCreated):
		return int(ErrorCodeCreateProblemError)
	}

	return http.StatusInternalServerError
}
