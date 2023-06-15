package managerv1

import (
	"fmt"
	"net/http"

	"github.com/karasunokami/chat-service/internal/middlewares"
	closechat "github.com/karasunokami/chat-service/internal/usecases/manager/close-chat"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostCloseChat(eCtx echo.Context, params PostCloseChatParams) error {
	ctx := eCtx.Request().Context()
	managerID := middlewares.MustUserID(eCtx)

	req := CloseChatRequest{}
	err := eCtx.Bind(&req)
	if err != nil {
		return fmt.Errorf("bind request, err=%w", err)
	}

	err = h.closeChat.Handle(ctx, closechat.Request{
		ManagerID: managerID,
		ChatID:    req.ChatId,
		ID:        params.XRequestID,
	})
	if err != nil {
		return newHandleError(err, getErrorCode(err))
	}

	return eCtx.JSON(http.StatusOK, CloseChatResponse{})
}
