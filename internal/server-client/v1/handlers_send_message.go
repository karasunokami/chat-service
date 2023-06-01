package clientv1

import (
	"fmt"
	"net/http"

	"github.com/karasunokami/chat-service/internal/middlewares"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/client/send-message"
	"github.com/karasunokami/chat-service/pkg/pointer"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostSendMessage(eCtx echo.Context, params PostSendMessageParams) error {
	ctx := eCtx.Request().Context()
	clientID := middlewares.MustUserID(eCtx)

	req := SendMessageRequest{}
	err := eCtx.Bind(&req)
	if err != nil {
		return fmt.Errorf("bind request, err=%w", err)
	}

	resp, err := h.sendMessage.Handle(ctx, sendmessage.Request{
		ID:          params.XRequestID,
		ClientID:    clientID,
		MessageBody: req.MessageBody,
	})
	if err != nil {
		return newHandleError(err, getErrorCode(err))
	}

	return eCtx.JSON(http.StatusOK, SendMessageResponse{
		Data: &MessageHeader{
			AuthorId:  pointer.PtrWithZeroAsNil(resp.AuthorID),
			CreatedAt: resp.CreatedAt,
			Id:        resp.MessageID,
		},
	})
}
