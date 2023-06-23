package sendmessage_test

import (
	"context"
	"io"
	"testing"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	problemsrepo "github.com/karasunokami/chat-service/internal/repositories/problems"
	sendmanagermessagejob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/send-manager-message"
	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"
	sendmessage "github.com/karasunokami/chat-service/internal/usecases/manager/send-message"
	sendmessagemocks "github.com/karasunokami/chat-service/internal/usecases/manager/send-message/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type UseCaseSuite struct {
	testingh.ContextSuite

	ctrl        *gomock.Controller
	msgRepo     *sendmessagemocks.MockmessagesRepository
	problemRepo *sendmessagemocks.MockproblemsRepository
	txtor       *sendmessagemocks.Mocktransactor
	outBoxSvc   *sendmessagemocks.MockoutboxService
	uCase       sendmessage.UseCase
}

func TestUseCaseSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UseCaseSuite))
}

func (s *UseCaseSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.msgRepo = sendmessagemocks.NewMockmessagesRepository(s.ctrl)
	s.outBoxSvc = sendmessagemocks.NewMockoutboxService(s.ctrl)
	s.problemRepo = sendmessagemocks.NewMockproblemsRepository(s.ctrl)
	s.txtor = sendmessagemocks.NewMocktransactor(s.ctrl)

	var err error
	s.uCase, err = sendmessage.New(sendmessage.NewOptions(s.msgRepo, s.outBoxSvc, s.problemRepo, s.txtor))
	s.Require().NoError(err)

	s.ContextSuite.SetupTest()
}

func (s *UseCaseSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *UseCaseSuite) TestRequestValidationError() {
	// Arrange.
	req := sendmessage.Request{}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, sendmessage.ErrInvalidRequest)
	s.Empty(resp.MessageID)
}

func (s *UseCaseSuite) TestGetAssignedProblemError() {
	// Arrange.
	req := sendmessage.Request{
		ID:          types.NewRequestID(),
		ManagerID:   types.NewUserID(),
		ChatID:      types.NewChatID(),
		MessageBody: `Hi`,
	}

	s.problemRepo.EXPECT().GetAssignedProblemID(s.Ctx, req.ManagerID, req.ChatID).
		Return(types.ProblemIDNil, problemsrepo.ErrNotFound)

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, sendmessage.ErrProblemNotFound)
	s.Empty(resp.MessageID)
}

func (s *UseCaseSuite) TestCreateMessageError() {
	// Arrange.
	req := sendmessage.Request{
		ID:          types.NewRequestID(),
		ManagerID:   types.NewUserID(),
		ChatID:      types.NewChatID(),
		MessageBody: `Hi`,
	}

	problemID := types.NewProblemID()
	expectedError := io.EOF

	s.problemRepo.EXPECT().GetAssignedProblemID(s.Ctx, req.ManagerID, req.ChatID).Return(problemID, nil)
	s.txtor.EXPECT().RunInTx(s.Ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(ctx context.Context) error) error {
			return f(ctx)
		})
	s.msgRepo.EXPECT().CreateFullVisible(s.Ctx, req.ID, problemID, req.ChatID, req.ManagerID, req.MessageBody).
		Return(nil, expectedError)

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, expectedError)
	s.Empty(resp.MessageID)
}

func (s *UseCaseSuite) TestPutJobToOutboxServiceMessageError() {
	// Arrange.
	req := sendmessage.Request{
		ID:          types.NewRequestID(),
		ManagerID:   types.NewUserID(),
		ChatID:      types.NewChatID(),
		MessageBody: `Hi`,
	}

	problemID := types.NewProblemID()
	expectedError := io.EOF
	expectedMessage := &messagesrepo.Message{
		ID: types.NewMessageID(),
	}

	s.problemRepo.EXPECT().GetAssignedProblemID(s.Ctx, req.ManagerID, req.ChatID).Return(problemID, nil)
	s.txtor.EXPECT().RunInTx(s.Ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(ctx context.Context) error) error {
			err := f(ctx)
			s.Require().NoError(err)

			return expectedError
		})
	s.msgRepo.EXPECT().CreateFullVisible(
		s.Ctx,
		req.ID,
		problemID,
		req.ChatID,
		req.ManagerID,
		req.MessageBody,
	).Return(expectedMessage, nil)
	s.outBoxSvc.EXPECT().Put(s.Ctx, sendmanagermessagejob.Name, gomock.Any(), gomock.Any()).Return(types.NewJobID(), nil)

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, expectedError)
	s.Empty(resp.MessageID)
}

func (s *UseCaseSuite) TestSuccess() {
	// Arrange.
	req := sendmessage.Request{
		ID:          types.NewRequestID(),
		ManagerID:   types.NewUserID(),
		ChatID:      types.NewChatID(),
		MessageBody: `Hi`,
	}

	problemID := types.NewProblemID()
	expectedMessage := &messagesrepo.Message{
		ID: types.NewMessageID(),
	}

	s.problemRepo.EXPECT().GetAssignedProblemID(s.Ctx, req.ManagerID, req.ChatID).Return(problemID, nil)
	s.txtor.EXPECT().RunInTx(s.Ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(ctx context.Context) error) error {
			return f(ctx)
		})
	s.msgRepo.EXPECT().CreateFullVisible(
		s.Ctx,
		req.ID,
		problemID,
		req.ChatID,
		req.ManagerID,
		req.MessageBody,
	).Return(expectedMessage, nil)
	s.outBoxSvc.EXPECT().Put(s.Ctx, sendmanagermessagejob.Name, gomock.Any(), gomock.Any()).Return(types.NewJobID(), nil)

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().NoError(err)
	s.NotEmpty(resp.MessageID)
	s.EqualValues(expectedMessage.ID, resp.MessageID)
}
