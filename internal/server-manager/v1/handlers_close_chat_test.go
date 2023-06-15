package managerv1_test

import (
	"errors"
	"fmt"
	"net/http"

	internalerrors "github.com/karasunokami/chat-service/internal/errors"
	managerv1 "github.com/karasunokami/chat-service/internal/server-manager/v1"
	"github.com/karasunokami/chat-service/internal/types"
	closechat "github.com/karasunokami/chat-service/internal/usecases/manager/close-chat"
)

func (s *HandlersSuite) TestCloseChat_UseCase_Error() {
	// Arrange.
	reqID := types.NewRequestID()
	chatID := types.NewChatID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/closeChat", fmt.Sprintf(`{"chatId": "%s"}`, chatID))
	s.closeChatUseCase.EXPECT().Handle(eCtx.Request().Context(), closechat.Request{
		ManagerID: s.managerID,
		ChatID:    chatID,
		ID:        reqID,
	}).Return(errors.New("something went wrong"))

	// Action.
	err := s.handlers.PostCloseChat(eCtx, managerv1.PostCloseChatParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestCloseChat_UseCase_NoProblemError() {
	// Arrange.
	reqID := types.NewRequestID()
	chatID := types.NewChatID()

	resp, eCtx := s.newEchoCtx(reqID, "/v1/closeChat", fmt.Sprintf(`{"chatId": "%s"}`, chatID))
	s.closeChatUseCase.EXPECT().Handle(eCtx.Request().Context(), closechat.Request{
		ManagerID: s.managerID,
		ChatID:    chatID,
		ID:        reqID,
	}).Return(closechat.ErrProblemNotFound)

	// Action.
	err := s.handlers.PostCloseChat(eCtx, managerv1.PostCloseChatParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.EqualValues(managerv1.ErrorCodeProblemNotFoundError, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestCloseChat_Usecase_Success() {
	// Arrange.
	reqID := types.NewRequestID()
	chatID := types.NewChatID()

	resp, eCtx := s.newEchoCtx(reqID, "/v1/CloseChat", fmt.Sprintf(`{"chatId": "%s"}`, chatID))
	s.closeChatUseCase.EXPECT().Handle(eCtx.Request().Context(), closechat.Request{
		ID:        reqID,
		ManagerID: s.managerID,
		ChatID:    chatID,
	}).Return(nil)

	// Action.
	err := s.handlers.PostCloseChat(eCtx, managerv1.PostCloseChatParams{XRequestID: reqID})

	// Assert.
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.Code)
	s.JSONEq(`
{
   "data": null
}`, resp.Body.String())
}
