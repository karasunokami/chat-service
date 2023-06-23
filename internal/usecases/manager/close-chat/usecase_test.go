package closechat_test

import (
	"context"
	"io"
	"testing"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	chatclosed "github.com/karasunokami/chat-service/internal/services/outbox/jobs/chat-closed"
	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"
	closechat "github.com/karasunokami/chat-service/internal/usecases/manager/close-chat"
	closechatmocks "github.com/karasunokami/chat-service/internal/usecases/manager/close-chat/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type UseCaseSuite struct {
	testingh.ContextSuite

	ctrl              *gomock.Controller
	problemsRepoMock  *closechatmocks.MockproblemsRepo
	messagesRepoMock  *closechatmocks.MockmessagesRepo
	outboxServiceMock *closechatmocks.MockoutboxService
	transactorMock    *closechatmocks.Mocktransactor
	uCase             closechat.UseCase
}

func TestUseCaseSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UseCaseSuite))
}

func (s *UseCaseSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.problemsRepoMock = closechatmocks.NewMockproblemsRepo(s.ctrl)
	s.messagesRepoMock = closechatmocks.NewMockmessagesRepo(s.ctrl)
	s.outboxServiceMock = closechatmocks.NewMockoutboxService(s.ctrl)
	s.transactorMock = closechatmocks.NewMocktransactor(s.ctrl)

	var err error
	s.uCase, err = closechat.New(closechat.NewOptions(
		s.outboxServiceMock,
		s.problemsRepoMock,
		s.messagesRepoMock,
		s.transactorMock,
	))
	s.Require().NoError(err)

	s.ContextSuite.SetupTest()
}

func (s *UseCaseSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *UseCaseSuite) TestInvalidRequest() {
	// Arrange.
	req := closechat.Request{}

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, closechat.ErrInvalidRequest)
}

func (s *UseCaseSuite) TestGetAssignedProblemIDError() {
	// Arrange.
	reqID := types.NewRequestID()
	managerID := types.NewUserID()
	chatID := types.NewChatID()

	req := closechat.Request{
		ID:        reqID,
		ManagerID: managerID,
		ChatID:    chatID,
	}

	expectedError := io.EOF

	s.problemsRepoMock.EXPECT().GetAssignedProblemID(s.Ctx, managerID, chatID).Return(
		types.ProblemIDNil,
		expectedError,
	)

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, expectedError)
}

func (s *UseCaseSuite) TestTransactorError() {
	// Arrange.
	reqID := types.NewRequestID()
	managerID := types.NewUserID()
	chatID := types.NewChatID()
	problemID := types.ProblemID{}

	req := closechat.Request{
		ID:        reqID,
		ManagerID: managerID,
		ChatID:    chatID,
	}

	expectedError := io.EOF

	s.problemsRepoMock.EXPECT().GetAssignedProblemID(s.Ctx, managerID, chatID).Return(
		problemID,
		nil,
	)
	s.transactorMock.EXPECT().RunInTx(s.Ctx, gomock.Any()).Return(expectedError)

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, expectedError)
}

func (s *UseCaseSuite) TestMarkProblemAsResolvedError() {
	// Arrange.
	reqID := types.NewRequestID()
	managerID := types.NewUserID()
	chatID := types.NewChatID()
	problemID := types.ProblemID{}

	req := closechat.Request{
		ID:        reqID,
		ManagerID: managerID,
		ChatID:    chatID,
	}

	expectedError := io.EOF

	s.problemsRepoMock.EXPECT().GetAssignedProblemID(s.Ctx, managerID, chatID).Return(
		problemID,
		nil,
	)
	s.transactorMock.EXPECT().RunInTx(s.Ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(ctx context.Context) error) error {
			return f(ctx)
		})
	s.problemsRepoMock.EXPECT().MarkProblemAsResolved(s.Ctx, problemID).Return(expectedError)

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, expectedError)
}

func (s *UseCaseSuite) TestCreateClientServiceMessageError() {
	// Arrange.
	reqID := types.NewRequestID()
	managerID := types.NewUserID()
	chatID := types.NewChatID()
	problemID := types.ProblemID{}

	req := closechat.Request{
		ID:        reqID,
		ManagerID: managerID,
		ChatID:    chatID,
	}

	expectedError := io.EOF

	s.problemsRepoMock.EXPECT().GetAssignedProblemID(s.Ctx, managerID, chatID).Return(
		problemID,
		nil,
	)
	s.transactorMock.EXPECT().RunInTx(s.Ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(ctx context.Context) error) error {
			return f(ctx)
		})
	s.problemsRepoMock.EXPECT().MarkProblemAsResolved(s.Ctx, problemID).Return(nil)
	s.messagesRepoMock.EXPECT().CreateClientService(s.Ctx, problemID, chatID, gomock.Any()).
		Return(nil, expectedError)

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, expectedError)
}

func (s *UseCaseSuite) TestOutboxPutError() {
	// Arrange.
	reqID := types.NewRequestID()
	managerID := types.NewUserID()
	chatID := types.NewChatID()
	problemID := types.ProblemID{}
	msg := messagesrepo.Message{ID: types.NewMessageID()}

	req := closechat.Request{
		ID:        reqID,
		ManagerID: managerID,
		ChatID:    chatID,
	}

	expectedError := io.EOF

	s.problemsRepoMock.EXPECT().GetAssignedProblemID(s.Ctx, managerID, chatID).Return(
		problemID,
		nil,
	)
	s.transactorMock.EXPECT().RunInTx(s.Ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(ctx context.Context) error) error {
			return f(ctx)
		})
	s.problemsRepoMock.EXPECT().MarkProblemAsResolved(s.Ctx, problemID).Return(nil)
	s.messagesRepoMock.EXPECT().CreateClientService(s.Ctx, problemID, chatID, gomock.Any()).
		Return(&msg, nil)

	payload, err := chatclosed.MarshalPayload(managerID, msg.ID, reqID)
	s.Require().NoError(err)

	s.outboxServiceMock.EXPECT().Put(s.Ctx, chatclosed.Name, payload, gomock.Any()).
		Return(types.JobIDNil, expectedError)

	// Action.
	err = s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, expectedError)
}

func (s *UseCaseSuite) TestSuccessStory() {
	// Arrange.
	reqID := types.NewRequestID()
	managerID := types.NewUserID()
	chatID := types.NewChatID()
	problemID := types.ProblemID{}
	msg := messagesrepo.Message{ID: types.NewMessageID()}

	req := closechat.Request{
		ID:        reqID,
		ManagerID: managerID,
		ChatID:    chatID,
	}

	s.problemsRepoMock.EXPECT().GetAssignedProblemID(s.Ctx, managerID, chatID).Return(
		problemID,
		nil,
	)
	s.transactorMock.EXPECT().RunInTx(s.Ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(ctx context.Context) error) error {
			return f(ctx)
		})
	s.problemsRepoMock.EXPECT().MarkProblemAsResolved(s.Ctx, problemID).Return(nil)
	s.messagesRepoMock.EXPECT().CreateClientService(s.Ctx, problemID, chatID, gomock.Any()).
		Return(&msg, nil)

	payload, err := chatclosed.MarshalPayload(managerID, msg.ID, reqID)
	s.Require().NoError(err)

	s.outboxServiceMock.EXPECT().Put(s.Ctx, chatclosed.Name, payload, gomock.Any()).
		Return(types.NewJobID(), nil)

	// Action.
	err = s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().NoError(err)
}
