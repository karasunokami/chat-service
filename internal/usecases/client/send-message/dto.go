package sendmessage

import (
	"time"

	"github.com/karasunokami/chat-service/internal/types"
	"github.com/karasunokami/chat-service/internal/validator"
)

type Request struct {
	ID          types.RequestID `validate:"required"`
	ClientID    types.UserID    `validate:"required"`
	MessageBody string          `validate:"required,gte=1,lte=3000"`
}

func (r Request) Validate() error {
	return validator.Validator.Struct(r)
}

type Response struct {
	AuthorID  types.UserID    `json:"authorId"`
	MessageID types.MessageID `json:"id"`
	CreatedAt time.Time       `json:"createdAt"`
}
