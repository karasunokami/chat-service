package managerv1

import (
	"errors"
	"net/http"

	internalerrors "github.com/karasunokami/chat-service/internal/errors"
	canreceiveproblems "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"
	closechat "github.com/karasunokami/chat-service/internal/usecases/manager/close-chat"
	freehands "github.com/karasunokami/chat-service/internal/usecases/manager/free-hands"
	getchats "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/manager/get-history"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/manager/send-message"
)

const defaultHandleErrorMessage = "cannot handle something"

func newHandleError(err error, code int) error {
	return internalerrors.NewServerError(code, defaultHandleErrorMessage, err)
}

func getErrorCode(err error) int {
	switch {
	case errors.Is(err, freehands.ErrInvalidRequest),
		errors.Is(err, getchats.ErrInvalidRequest),
		errors.Is(err, canreceiveproblems.ErrInvalidRequest),
		errors.Is(err, sendmessage.ErrInvalidRequest),
		errors.Is(err, gethistory.ErrInvalidRequest),
		errors.Is(err, gethistory.ErrInvalidCursor),
		errors.Is(err, closechat.ErrInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, freehands.ErrManagerOverload):
		return int(ErrorCodeFreeHandsManagerOverloadError)
	case errors.Is(err, closechat.ErrProblemNotFound):
		return int(ErrorCodeProblemNotFoundError)
	}

	return http.StatusInternalServerError
}
