package managerv1_test

import (
	"errors"
	"fmt"
	"net/http"

	managerv1 "github.com/karasunokami/chat-service/internal/server-manager/v1"
	"github.com/karasunokami/chat-service/internal/types"
	getchats "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats"
)

func (s *HandlersSuite) TestGetChats_Usecase_UnknownError() {
	// Arrange.
	reqID := types.NewRequestID()

	resp, eCtx := s.newEchoCtx(reqID, "/v1/getChats", fmt.Sprintf(`{"managerId":"%s"}`, s.managerID.String()))
	s.getChatsUseCase.EXPECT().Handle(eCtx.Request().Context(), getchats.Request{
		ID:        reqID,
		ManagerID: s.managerID,
	}).Return(getchats.Response{}, errors.New("something went wrong"))

	// Action.
	err := s.handlers.PostGetChats(eCtx, managerv1.PostGetChatsParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestGetChats_Usecase_Success() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/getChats", `{"pageSize":10}`)

	chats := []getchats.Chat{
		{
			ID:       types.NewChatID(),
			ClientID: types.NewUserID(),
		},
		{
			ID:       types.NewChatID(),
			ClientID: types.NewUserID(),
		},
	}
	s.getChatsUseCase.EXPECT().Handle(eCtx.Request().Context(), getchats.Request{
		ID:        reqID,
		ManagerID: s.managerID,
	}).Return(getchats.Response{
		Chats: chats,
	}, nil)

	// Action.
	err := s.handlers.PostGetChats(eCtx, managerv1.PostGetChatsParams{XRequestID: reqID})

	// Assert.
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.Code)
	s.JSONEq(fmt.Sprintf(`
{
    "data":
    {
        "chats":
        [
            {
                "chatId": %q,
                "clientId": %q
            },
            {
                "chatId": %q,
                "clientId": %q
            }
        ]
    }
}`, chats[0].ID, chats[0].ClientID, chats[1].ID, chats[1].ClientID), resp.Body.String())
}
