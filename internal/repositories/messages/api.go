package messagesrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/karasunokami/chat-service/internal/store"
	"github.com/karasunokami/chat-service/internal/store/message"
	"github.com/karasunokami/chat-service/internal/store/problem"
	"github.com/karasunokami/chat-service/internal/types"
)

var ErrMsgNotFound = errors.New("message not found")

func (r *Repo) GetMessageByRequestID(ctx context.Context, reqID types.RequestID) (*Message, error) {
	mes, err := r.db.Message(ctx).Query().Where(message.InitialRequestID(reqID)).First(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return nil, ErrMsgNotFound
		}

		return nil, fmt.Errorf("db select message by request id, err=%v", err)
	}

	return storeMessageToRepoMessage(mes), nil
}

func (r *Repo) GetMessageByID(ctx context.Context, id types.MessageID) (*Message, error) {
	mes, err := r.db.Message(ctx).Query().
		Where(
			message.IsService(false),
			message.IDEQ(id),
		).
		First(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return nil, ErrMsgNotFound
		}

		return nil, fmt.Errorf("db select message by id, err=%v", err)
	}

	return storeMessageToRepoMessage(mes), nil
}

// CreateClientVisible creates a message that is visible only to the client.
func (r *Repo) CreateClientVisible(
	ctx context.Context,
	reqID types.RequestID,
	problemID types.ProblemID,
	chatID types.ChatID,
	authorID types.UserID,
	msgBody string,
) (*Message, error) {
	mes, err := r.db.Message(ctx).Create().
		SetInitialRequestID(reqID).
		SetProblemID(problemID).
		SetChatID(chatID).
		SetAuthorID(authorID).
		SetBody(msgBody).
		SetIsVisibleForClient(true).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("db create new message, err=%v", err)
	}

	return storeMessageToRepoMessage(mes), nil
}

// CreateService creates a service message.
func (r *Repo) CreateService(
	ctx context.Context,
	problemID types.ProblemID,
	chatID types.ChatID,
	msgBody string,
) (*Message, error) {
	mes, err := r.db.Message(ctx).Create().
		SetProblemID(problemID).
		SetChatID(chatID).
		SetBody(msgBody).
		SetIsVisibleForClient(true).
		SetIsService(true).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("db create new message, err=%v", err)
	}

	return storeMessageToRepoMessage(mes), nil
}

func (r *Repo) CreateFullVisible(
	ctx context.Context,
	reqID types.RequestID,
	problemID types.ProblemID,
	chatID types.ChatID,
	authorID types.UserID,
	msgBody string,
) (*Message, error) {
	mes, err := r.db.Message(ctx).Create().
		SetInitialRequestID(reqID).
		SetProblemID(problemID).
		SetChatID(chatID).
		SetAuthorID(authorID).
		SetBody(msgBody).
		SetIsVisibleForClient(true).
		SetIsVisibleForManager(true).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("db create new message, err=%v", err)
	}

	return storeMessageToRepoMessage(mes), nil
}

// GetFirstProblemMessage get first message of problem by problem id.
func (r *Repo) GetFirstProblemMessage(ctx context.Context, problemID types.ProblemID) (*Message, error) {
	mes, err := r.db.Message(ctx).Query().
		Where(message.HasProblemWith(problem.IDEQ(problemID))).
		Order(store.Asc(message.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return nil, ErrMsgNotFound
		}
		return nil, fmt.Errorf("select first problem message, err=%v", err)
	}

	return storeMessageToRepoMessage(mes), nil
}
