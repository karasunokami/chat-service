package managerv1

import (
	"errors"
	"net/http"

	internalerrors "github.com/karasunokami/chat-service/internal/errors"
	freehands "github.com/karasunokami/chat-service/internal/usecases/manager/free-hands"
	getchats "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/manager/get-history"
)

const defaultHandleErrorMessage = "cannot handle something"

func newHandleError(err error, code int) error {
	return internalerrors.NewServerError(code, defaultHandleErrorMessage, err)
}

func getErrorCode(err error) int {
	switch {
	case errors.Is(err, freehands.ErrInvalidRequest),
		errors.Is(err, getchats.ErrInvalidRequest),
		errors.Is(err, gethistory.ErrInvalidRequest),
		errors.Is(err, gethistory.ErrInvalidCursor):
		return http.StatusBadRequest
	case errors.Is(err, freehands.ErrManagerOverload):
		return int(ErrorCodeFreeHandsManagerOverloadError)
	}

	return http.StatusInternalServerError
}
