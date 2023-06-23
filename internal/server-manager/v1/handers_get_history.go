package managerv1

import (
	"net/http"

	"github.com/karasunokami/chat-service/internal/middlewares"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/manager/get-history"
	"github.com/karasunokami/chat-service/pkg/pointer"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostGetChatHistory(eCtx echo.Context, params PostGetChatHistoryParams) error {
	ctx := eCtx.Request().Context()
	managerID := middlewares.MustUserID(eCtx)

	req := GetHistoryRequest{}
	err := eCtx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	resp, err := h.getHistory.Handle(ctx, gethistory.Request{
		ID:        params.XRequestID,
		ManagerID: managerID,
		ChatID:    req.ChatId,
		PageSize:  pointer.Indirect(req.PageSize),
		Cursor:    pointer.Indirect(req.Cursor),
	})
	if err != nil {
		return newHandleError(err, getErrorCode(err))
	}

	page := make([]Message, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		mm := Message{
			AuthorId:  m.AuthorID,
			Body:      m.Body,
			CreatedAt: m.CreatedAt,
			Id:        m.ID,
		}
		page = append(page, mm)
	}

	return eCtx.JSON(http.StatusOK, GetHistoryResponse{Data: &MessagesPage{
		Messages: page,
		Next:     resp.NextCursor,
	}})
}
