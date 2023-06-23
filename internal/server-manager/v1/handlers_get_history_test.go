package managerv1_test

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	internalerrors "github.com/karasunokami/chat-service/internal/errors"
	managerv1 "github.com/karasunokami/chat-service/internal/server-manager/v1"
	"github.com/karasunokami/chat-service/internal/types"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/manager/get-history"
)

func (s *HandlersSuite) TestGetHistory_BindRequestError() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/getChatHistory", `{"page_size":`)

	// Action.
	err := s.handlers.PostGetChatHistory(eCtx, managerv1.PostGetChatHistoryParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestGetHistory_Usecase_InvalidRequest() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/getChatHistory", `{"pageSize":9}`)
	s.getHistoryUseCase.EXPECT().Handle(eCtx.Request().Context(), gethistory.Request{
		ID:        reqID,
		ManagerID: s.managerID,
		PageSize:  9,
		Cursor:    "",
	}).Return(gethistory.Response{}, gethistory.ErrInvalidRequest)

	// Action.
	err := s.handlers.PostGetChatHistory(eCtx, managerv1.PostGetChatHistoryParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestGetHistory_Usecase_InvalidCursor() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/getChatHistory", `{"cursor":"abracadabra"}`)
	s.getHistoryUseCase.EXPECT().Handle(eCtx.Request().Context(), gethistory.Request{
		ID:        reqID,
		ManagerID: s.managerID,
		PageSize:  0,
		Cursor:    "abracadabra",
	}).Return(gethistory.Response{}, gethistory.ErrInvalidCursor)

	// Action.
	err := s.handlers.PostGetChatHistory(eCtx, managerv1.PostGetChatHistoryParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestGetHistory_Usecase_UnknownError() {
	// Arrange.
	reqID := types.NewRequestID()
	chatID := types.NewChatID()

	resp, eCtx := s.newEchoCtx(reqID, "/v1/getChatHistory", fmt.Sprintf(`{"pageSize":10, "chatId": "%s"}`, chatID))
	s.getHistoryUseCase.EXPECT().Handle(eCtx.Request().Context(), gethistory.Request{
		ID:        reqID,
		ChatID:    chatID,
		ManagerID: s.managerID,
		PageSize:  10,
	}).Return(gethistory.Response{}, errors.New("something went wrong"))

	// Action.
	err := s.handlers.PostGetChatHistory(eCtx, managerv1.PostGetChatHistoryParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestGetHistory_Usecase_Success() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/getChatHistory", `{"pageSize":10}`)

	msgs := []gethistory.Message{
		{
			ID:        types.NewMessageID(),
			AuthorID:  types.NewUserID(),
			Body:      "hello!",
			CreatedAt: time.Unix(1, 1).UTC(),
		},
		{
			ID:        types.NewMessageID(),
			AuthorID:  types.NewUserID(),
			Body:      "hello 2!",
			CreatedAt: time.Unix(2, 2).UTC(),
		},
	}
	s.getHistoryUseCase.EXPECT().Handle(eCtx.Request().Context(), gethistory.Request{
		ID:        reqID,
		ManagerID: s.managerID,
		PageSize:  10,
	}).Return(gethistory.Response{
		Messages:   msgs,
		NextCursor: "",
	}, nil)

	// Action.
	err := s.handlers.PostGetChatHistory(eCtx, managerv1.PostGetChatHistoryParams{XRequestID: reqID})

	// Assert.
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.Code)
	s.JSONEq(fmt.Sprintf(`
{
    "data":
    {
        "messages":
        [
            {
                "authorId": %q,
                "body": "hello!",
                "createdAt": "1970-01-01T00:00:01.000000001Z",
                "id": %q
            },
            {
                "authorId": %q,
                "body": "hello 2!",
                "createdAt": "1970-01-01T00:00:02.000000002Z",
                "id": %q
            }
        ],
        "next": ""
    }
}`, msgs[0].AuthorID, msgs[0].ID, msgs[1].AuthorID, msgs[1].ID), resp.Body.String())
}
