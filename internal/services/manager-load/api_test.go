package managerload_test

import (
	"errors"
	"testing"

	managerload "github.com/karasunokami/chat-service/internal/services/manager-load"
	managerloadmocks "github.com/karasunokami/chat-service/internal/services/manager-load/mocks"
	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

const maxProblemAtTime = 5

type ServiceSuite struct {
	testingh.ContextSuite

	ctrl *gomock.Controller

	problemsRepo *managerloadmocks.MockproblemsRepository
	managerLoad  *managerload.Service
}

func TestServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.problemsRepo = managerloadmocks.NewMockproblemsRepository(s.ctrl)

	managerLoad, err := managerload.New(managerload.NewOptions(5, s.problemsRepo))
	if err != nil {
		s.Fail("create manager load", err)
	}

	s.managerLoad = managerLoad

	s.ContextSuite.SetupTest()
}

func (s *ServiceSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *ServiceSuite) TestCanManagerTakeProblem() {
	cases := []struct {
		name           string
		activeProblems int
		canTake        bool
		repoError      error
	}{
		{
			name:           "with zero active problems",
			activeProblems: 0,
			canTake:        true,
		},
		{
			name:           "with one active problem",
			activeProblems: 1,
			canTake:        true,
		},
		{
			name:           "with minus one active problem",
			activeProblems: -1,
			canTake:        true,
		},
		{
			name:           "with maxProblemAtTime active problems",
			activeProblems: maxProblemAtTime,
			canTake:        false,
		},
		{
			name:           "with maxProblemAtTime +1 active problems",
			activeProblems: maxProblemAtTime + 1,
			canTake:        false,
		},
		{
			name:      "with repo error",
			canTake:   false,
			repoError: errors.New("error"),
		},
	}

	for _, tt := range cases {
		s.Run(tt.name, func() {
			// Arrange.

			managerID := types.NewUserID()
			s.problemsRepo.EXPECT().GetManagerOpenProblemsCount(gomock.Any(), managerID).Return(tt.activeProblems, tt.repoError)

			// Action.

			can, err := s.managerLoad.CanManagerTakeProblem(s.Ctx, managerID)

			// Assert.

			if tt.repoError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tt.repoError)
			} else {
				s.Require().NoError(err)
			}

			s.Require().EqualValues(tt.canTake, can)
		})
	}
}
