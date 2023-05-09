package clientv1

import (
	"errors"
	"net/http"

	internalerrors "github.com/karasunokami/chat-service/internal/errors"
	"github.com/karasunokami/chat-service/internal/middlewares"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/client/send-message"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostSendMessage(eCtx echo.Context, params PostSendMessageParams) error {
	ctx := eCtx.Request().Context()
	clientID := middlewares.MustUserID(eCtx)

	req := SendMessageRequest{}
	err := eCtx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	resp, err := h.sendMessage.Handle(ctx, sendmessage.Request{
		ID:          params.XRequestID,
		ClientID:    clientID,
		MessageBody: req.MessageBody,
	})
	if err != nil {
		return internalerrors.NewServerError(getErrorCode(err), err.Error(), err)
	}

	return eCtx.JSON(http.StatusOK, Response{Data: resp})
}

func getErrorCode(err error) int {
	switch {
	case errors.Is(err, sendmessage.ErrInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, sendmessage.ErrChatNotCreated):
		return int(ErrorCodeCreateChatError)
	case errors.Is(err, sendmessage.ErrProblemNotCreated):
		return int(ErrorCodeCreateProblemError)
	}

	return http.StatusInternalServerError
}
