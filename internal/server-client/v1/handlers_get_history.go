package clientv1

import (
	"fmt"
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
	err := eCtx.JSON(200, GetHistoryResponse{Data: stub})
	if err != nil {
		return fmt.Errorf("ectx json, err=%v", err)
	}

	return nil
}
