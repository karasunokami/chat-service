package managerv1_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karasunokami/chat-service/internal/middlewares"
	managerv1 "github.com/karasunokami/chat-service/internal/server-manager/v1"
	managerv1mocks "github.com/karasunokami/chat-service/internal/server-manager/v1/mocks"
	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type HandlersSuite struct {
	testingh.ContextSuite

	ctrl                      *gomock.Controller
	canReceiveProblemsUseCase *managerv1mocks.MockcanReceiveProblemsUseCase
	handlers                  managerv1.Handlers

	managerID          types.UserID
	freeHandsUseCase   *managerv1mocks.MockfreeHandsUseCase
	getChatsUseCase    *managerv1mocks.MockgetChatsUseCase
	getHistoryUseCase  *managerv1mocks.MockgetHistoryUseCase
	sendMessageUseCase *managerv1mocks.MocksendMessageUseCase
	closeChatUseCase   *managerv1mocks.MockcloseChatUseCase
}

func TestHandlersSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(HandlersSuite))
}

func (s *HandlersSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.canReceiveProblemsUseCase = managerv1mocks.NewMockcanReceiveProblemsUseCase(s.ctrl)
	s.freeHandsUseCase = managerv1mocks.NewMockfreeHandsUseCase(s.ctrl)
	s.getChatsUseCase = managerv1mocks.NewMockgetChatsUseCase(s.ctrl)
	s.getHistoryUseCase = managerv1mocks.NewMockgetHistoryUseCase(s.ctrl)
	s.sendMessageUseCase = managerv1mocks.NewMocksendMessageUseCase(s.ctrl)
	s.closeChatUseCase = managerv1mocks.NewMockcloseChatUseCase(s.ctrl)
	{
		var err error
		s.handlers, err = managerv1.NewHandlers(managerv1.NewOptions(
			s.canReceiveProblemsUseCase,
			s.freeHandsUseCase,
			s.getChatsUseCase,
			s.getHistoryUseCase,
			s.sendMessageUseCase,
			s.closeChatUseCase,
		))
		s.Require().NoError(err)
	}
	s.managerID = types.NewUserID()

	s.ContextSuite.SetupTest()
}

func (s *HandlersSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *HandlersSuite) newEchoCtx(
	requestID types.RequestID,
	path string,
	body string, //nolint:unparam
) (*httptest.ResponseRecorder, echo.Context) {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderXRequestID, requestID.String())

	resp := httptest.NewRecorder()

	ctx := echo.New().NewContext(req, resp)
	middlewares.SetToken(ctx, s.managerID)

	return resp, ctx
}
