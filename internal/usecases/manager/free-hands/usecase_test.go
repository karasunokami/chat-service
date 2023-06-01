package freehands_test

import (
	"errors"
	"testing"

	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"
	freehands "github.com/karasunokami/chat-service/internal/usecases/manager/free-hands"
	freehandsmocks "github.com/karasunokami/chat-service/internal/usecases/manager/free-hands/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type UseCaseSuite struct {
	testingh.ContextSuite

	ctrl      *gomock.Controller
	mLoadMock *freehandsmocks.MockmanagerLoadService
	mPoolMock *freehandsmocks.MockmanagerPool
	uCase     freehands.UseCase
}

func TestUseCaseSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UseCaseSuite))
}

func (s *UseCaseSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mLoadMock = freehandsmocks.NewMockmanagerLoadService(s.ctrl)
	s.mPoolMock = freehandsmocks.NewMockmanagerPool(s.ctrl)

	var err error
	s.uCase, err = freehands.New(freehands.NewOptions(s.mLoadMock, s.mPoolMock))
	s.Require().NoError(err)

	s.ContextSuite.SetupTest()
}

func (s *UseCaseSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *UseCaseSuite) TestInvalidRequest() {
	// Arrange.
	req := freehands.Request{}

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, freehands.ErrInvalidRequest)
}

func (s *UseCaseSuite) TestCanManagerTakeProblemError() {
	// Arrange.
	managerID := types.NewUserID()
	req := freehands.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, managerID).Return(false, errors.New("error"))

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
}

func (s *UseCaseSuite) TestCanManagerTakeProblemReturnFalse() {
	// Arrange.
	managerID := types.NewUserID()
	req := freehands.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, managerID).Return(false, nil)

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.Require().ErrorIs(err, freehands.ErrManagerOverload)
}

func (s *UseCaseSuite) TestManagerPoolPutError() {
	// Arrange.
	managerID := types.NewUserID()
	req := freehands.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, managerID).Return(true, nil)
	s.mPoolMock.EXPECT().Put(s.Ctx, managerID).Return(errors.New("error"))

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
}

func (s *UseCaseSuite) TestSuccess() {
	// Arrange.
	managerID := types.NewUserID()
	req := freehands.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, managerID).Return(true, nil)
	s.mPoolMock.EXPECT().Put(s.Ctx, managerID).Return(nil)

	// Action.
	err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().NoError(err)
}
