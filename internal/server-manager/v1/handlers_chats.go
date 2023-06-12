package managerv1

import (
	"net/http"

	"github.com/karasunokami/chat-service/internal/middlewares"
	getchats "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostGetChats(eCtx echo.Context, params PostGetChatsParams) error {
	ctx := eCtx.Request().Context()
	managerID := middlewares.MustUserID(eCtx)

	r := getchats.Request{
		ID:        params.XRequestID,
		ManagerID: managerID,
	}

	resp, err := h.getChats.Handle(ctx, r)
	if err != nil {
		return newHandleError(err, getErrorCode(err))
	}

	chs := make([]Chat, 0, len(resp.Chats))
	for _, c := range resp.Chats {
		chs = append(chs, Chat{
			ChatId:   c.ID,
			ClientId: c.ClientID,
		})
	}

	return eCtx.JSON(http.StatusOK, GetChatsResponse{
		Data: &ChatList{Chats: chs},
	})
}
