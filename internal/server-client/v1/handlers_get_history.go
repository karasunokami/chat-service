package clientv1

import (
	"net/http"
	"time"

	"github.com/karasunokami/chat-service/internal/types"

	"github.com/labstack/echo/v4"
)

var stub = MessagesPage{Messages: []Message{
	{
		AuthorId:  types.NewUserID(),
		Body:      "Здравствуйте! Разберёмся.",
		CreatedAt: time.Now(),
		Id:        types.NewMessageID(),
	},
	{
		AuthorId:  types.MustParse[types.UserID]("187f5e1b-69cd-423d-9f06-b653cfcba290"),
		Body:      "Привет! Не могу снять денег с карты,\nпишет 'карта заблокирована'",
		CreatedAt: time.Now().Add(-time.Minute),
		Id:        types.NewMessageID(),
	},
}}

func (h Handlers) PostGetHistory(eCtx echo.Context, _ PostGetHistoryParams) error {
	return eCtx.JSON(http.StatusOK, GetHistoryResponse{Data: stub})
}
