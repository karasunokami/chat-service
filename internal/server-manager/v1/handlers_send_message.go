package managerv1

import (
	"fmt"
	"net/http"

	"github.com/karasunokami/chat-service/internal/middlewares"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/manager/send-message"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostSendMessage(eCtx echo.Context, params PostSendMessageParams) error {
	ctx := eCtx.Request().Context()
	managerID := middlewares.MustUserID(eCtx)

	req := SendMessageRequest{}
	err := eCtx.Bind(&req)
	if err != nil {
		return fmt.Errorf("bind request, err=%w", err)
	}

	resp, err := h.sendMessage.Handle(ctx, sendmessage.Request{
		ID:          params.XRequestID,
		ManagerID:   managerID,
		ChatID:      req.ChatId,
		MessageBody: req.MessageBody,
	})
	if err != nil {
		return newHandleError(err, getErrorCode(err))
	}

	return eCtx.JSON(http.StatusOK, SendMessageResponse{
		Data: &MessageWithoutBody{
			AuthorId:  managerID,
			CreatedAt: resp.CreatedAt,
			Id:        resp.MessageID,
		},
	})
}
