package gethistory

import (
	"context"
	"errors"
	"fmt"

	"github.com/karasunokami/chat-service/internal/cursor"
	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	"github.com/karasunokami/chat-service/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=gethistorymocks

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidCursor  = errors.New("invalid cursor")
)

type messagesRepository interface {
	GetClientChatMessages(
		ctx context.Context,
		clientID types.UserID,
		pageSize int,
		cursor *messagesrepo.Cursor,
	) ([]messagesrepo.Message, *messagesrepo.Cursor, error)
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	msgRepo messagesRepository `option:"mandatory" validate:"required"`
}

type UseCase struct {
	Options
}

func New(opts Options) (UseCase, error) {
	err := opts.Validate()
	if err != nil {
		return UseCase{}, fmt.Errorf("validate options, err=%v", err)
	}

	return UseCase{opts}, nil
}

func (u UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	err := req.Validate()
	if err != nil {
		return Response{}, fmt.Errorf("request validate, err=%w", ErrInvalidRequest)
	}

	var crs *messagesrepo.Cursor
	if req.Cursor != "" {
		if err := cursor.Decode(req.Cursor, &crs); err != nil {
			return Response{}, fmt.Errorf("decode cursor: %w: %v", ErrInvalidCursor, err)
		}
	}

	msgs, nextCrs, err := u.msgRepo.GetClientChatMessages(ctx, req.ClientID, req.PageSize, crs)
	if err != nil {
		if errors.Is(err, messagesrepo.ErrInvalidCursor) {
			return Response{}, fmt.Errorf("get client chat messages, err=%w, err=%v", ErrInvalidCursor, err)
		}

		return Response{}, fmt.Errorf("get client chat messages, err=%w", err)
	}

	return formatResp(msgs, nextCrs)
}

func formatResp(msgs []messagesrepo.Message, nextCrs *messagesrepo.Cursor) (Response, error) {
	var err error

	resp := Response{
		Messages: adoptMessages(msgs),
	}

	if nextCrs != nil {
		resp.NextCursor, err = cursor.Encode(nextCrs)
		if err != nil {
			return Response{}, fmt.Errorf("encode cursor, err=%v", err)
		}
	}

	return resp, nil
}
