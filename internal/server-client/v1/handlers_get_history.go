package clientv1

import (
	"errors"
	"net/http"

	"github.com/karasunokami/chat-service/internal/middlewares"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/client/get-history"
	"github.com/karasunokami/chat-service/pkg/pointer"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostGetHistory(eCtx echo.Context, params PostGetHistoryParams) error {
	ctx := eCtx.Request().Context()
	clientID := middlewares.MustUserID(eCtx)

	req := GetHistoryRequest{}
	err := eCtx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	resp, err := h.getHistory.Handle(ctx, gethistory.Request{
		ID:       params.XRequestID,
		ClientID: clientID,
		PageSize: pointer.Indirect(req.PageSize),
		Cursor:   pointer.Indirect(req.Cursor),
	})
	if err != nil {
		if errors.Is(err, gethistory.ErrInvalidRequest) || errors.Is(err, gethistory.ErrInvalidCursor) {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		return err
	}

	page := make([]Message, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		mm := Message{
			AuthorId:   m.AuthorID.AsPointer(),
			Body:       m.Body,
			CreatedAt:  m.CreatedAt,
			Id:         m.ID,
			IsBlocked:  m.IsBlocked,
			IsReceived: m.IsReceived,
			IsService:  m.IsService,
		}
		page = append(page, mm)
	}

	return eCtx.JSON(http.StatusOK, GetHistoryResponse{Data: &MessagesPage{
		Messages: page,
		Next:     resp.NextCursor,
	}})
}
