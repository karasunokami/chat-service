package canreceiveproblems_test

import (
	"errors"
	"testing"

	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"
	canreceiveproblems "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"
	canreceiveproblemsmocks "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type UseCaseSuite struct {
	testingh.ContextSuite

	ctrl      *gomock.Controller
	mLoadMock *canreceiveproblemsmocks.MockmanagerLoadService
	mPoolMock *canreceiveproblemsmocks.MockmanagerPool
	uCase     canreceiveproblems.UseCase
}

func TestUseCaseSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UseCaseSuite))
}

func (s *UseCaseSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mLoadMock = canreceiveproblemsmocks.NewMockmanagerLoadService(s.ctrl)
	s.mPoolMock = canreceiveproblemsmocks.NewMockmanagerPool(s.ctrl)

	var err error
	s.uCase, err = canreceiveproblems.New(canreceiveproblems.NewOptions(s.mLoadMock, s.mPoolMock))
	s.Require().NoError(err)

	s.ContextSuite.SetupTest()
}

func (s *UseCaseSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *UseCaseSuite) TestInvalidRequest() {
	// Arrange.
	req := canreceiveproblems.Request{}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, canreceiveproblems.ErrInvalidRequest)
	s.Empty(resp.Result)
}

func (s *UseCaseSuite) TestManagerPoolContainsError() {
	// Arrange.
	managerID := types.NewUserID()
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mPoolMock.EXPECT().Contains(s.Ctx, managerID).Return(false, errors.New("error"))

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.Empty(resp.Result)
}

func (s *UseCaseSuite) TestManagerIsInPoolError() {
	// Arrange.
	managerID := types.NewUserID()
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mPoolMock.EXPECT().Contains(s.Ctx, managerID).Return(true, nil)

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().NoError(err)
	s.False(resp.Result)
}

func (s *UseCaseSuite) TestLoadServiceError() {
	// Arrange.
	managerID := types.NewUserID()
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mPoolMock.EXPECT().Contains(s.Ctx, managerID).Return(false, nil)
	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, managerID).Return(false, errors.New("error"))

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.Empty(resp.Result)
}

func (s *UseCaseSuite) TestLoadServiceReturnsTrue() {
	// Arrange.
	managerID := types.NewUserID()
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mPoolMock.EXPECT().Contains(s.Ctx, managerID).Return(false, nil)
	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, managerID).Return(true, nil)

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().NoError(err)
	s.True(resp.Result)
}

func (s *UseCaseSuite) TestLoadServiceReturnsFalse() {
	// Arrange.
	managerID := types.NewUserID()
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mPoolMock.EXPECT().Contains(s.Ctx, managerID).Return(false, nil)
	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, managerID).Return(false, nil)

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().NoError(err)
	s.False(resp.Result)
}
