//go:build integration

package problemsrepo_test

import (
	"testing"

	problemsrepo "github.com/karasunokami/chat-service/internal/repositories/problems"
	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/stretchr/testify/suite"
)

type ProblemsRepoManagerAPISuite struct {
	testingh.DBSuite
	repo *problemsrepo.Repo
}

func TestProblemsRepoManagerAPISuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &ProblemsRepoManagerAPISuite{DBSuite: testingh.NewDBSuite("TestProblemsRepoManagerAPISuite")})
}

func (s *ProblemsRepoManagerAPISuite) SetupSuite() {
	s.DBSuite.SetupSuite()

	var err error

	s.repo, err = problemsrepo.New(problemsrepo.NewOptions(s.Database))
	s.Require().NoError(err)
}

func (s *ProblemsRepoManagerAPISuite) Test_GetProblemsWithoutManagers() {
	s.Run("problem with manager", func() {
		clientID := types.NewUserID()

		// Create chat.
		chat, err := s.Database.Chat(s.Ctx).Create().SetClientID(clientID).Save(s.Ctx)
		s.Require().NoError(err)

		// Create problem.
		_, err = s.Database.Problem(s.Ctx).Create().
			SetChatID(chat.ID).
			SetManagerID(types.NewUserID()).
			Save(s.Ctx)
		s.Require().NoError(err)

		problems, err := s.repo.GetProblemsWithoutManagers(s.Ctx, 1)
		s.Require().NoError(err)
		s.Empty(problems)
	})

	s.Run("problem without manager and messages", func() {
		clientID := types.NewUserID()

		_, _ = s.createChatWithProblemAssignedTo(clientID)

		problems, err := s.repo.GetProblemsWithoutManagers(s.Ctx, 1)
		s.Require().NoError(err)
		s.Empty(problems)
	})

	s.Run("problem without manager, with messages", func() {
		clientID := types.NewUserID()

		// Create chat.
		chat, err := s.Database.Chat(s.Ctx).Create().SetClientID(clientID).Save(s.Ctx)
		s.Require().NoError(err)

		// Create problem.
		problem, err := s.Database.Problem(s.Ctx).Create().
			SetChatID(chat.ID).
			Save(s.Ctx)
		s.Require().NoError(err)

		// Create message.
		_, err = s.Database.Message(s.Ctx).Create().
			SetProblemID(problem.ID).
			SetChatID(chat.ID).
			SetIsVisibleForManager(true).
			SetBody("body").
			Save(s.Ctx)
		s.Require().NoError(err)

		problems, err := s.repo.GetProblemsWithoutManagers(s.Ctx, 1)
		s.Require().NoError(err)
		s.Require().Len(problems, 1)
		s.EqualValues(problem.ID, problems[0].ID)
	})
}

func (s *ProblemsRepoManagerAPISuite) Test_SetManagerToProblem() {
	s.Run("set manager to problem", func() {
		managerID := types.NewUserID()

		chat, err := s.Database.Chat(s.Ctx).Create().SetClientID(types.NewUserID()).Save(s.Ctx)
		s.Require().NoError(err)

		p, err := s.Database.Problem(s.Ctx).Create().SetChatID(chat.ID).Save(s.Ctx)
		s.Require().NoError(err)

		err = s.repo.SetManagerToProblem(s.Ctx, p.ID, managerID)
		s.Require().NoError(err)

		problem, err := s.Database.Problem(s.Ctx).Get(s.Ctx, p.ID)
		s.Require().NoError(err)
		s.EqualValues(managerID, problem.ManagerID)
	})
}

func (s *ProblemsRepoManagerAPISuite) Test_GetAssignedProblemID() {
	s.Run("set manager to problem", func() {
		managerID := types.NewUserID()

		chatID, problemID := s.createChatWithProblemAssignedTo(managerID)

		foundProblemID, err := s.repo.GetAssignedProblemID(s.Ctx, managerID, chatID)
		s.Require().NoError(err)
		s.EqualValues(problemID, foundProblemID)
	})
}

func (s *ProblemsRepoManagerAPISuite) Test_MarkProblemAsResolved() {
	s.Run("mark problem as resolved", func() {
		managerID := types.NewUserID()

		_, problemID := s.createChatWithProblemAssignedTo(managerID)

		err := s.repo.MarkProblemAsResolved(s.Ctx, problemID)
		s.Require().NoError(err)

		problem, err := s.Database.Problem(s.Ctx).Get(s.Ctx, problemID)
		s.Require().NoError(err)
		s.EqualValues(managerID, problem.ManagerID)
		s.NotEmpty(problem.ResolvedAt)
	})
}

func (s *ProblemsRepoManagerAPISuite) Test_CantMarkAnotherManagerProblemAsResolved() {
	s.Run("mark problem as resolved", func() {
		err := s.repo.MarkProblemAsResolved(s.Ctx, types.NewProblemID())
		s.Require().Error(err)
		s.ErrorIs(err, problemsrepo.ErrNotFound)
	})
}

func (s *ProblemsRepoManagerAPISuite) createChatWithProblemAssignedTo(managerID types.UserID) (types.ChatID, types.ProblemID) {
	s.T().Helper()

	// 1 chat can have only 1 open problem.

	chat, err := s.Database.Chat(s.Ctx).Create().SetClientID(types.NewUserID()).Save(s.Ctx)
	s.Require().NoError(err)

	p, err := s.Database.Problem(s.Ctx).Create().SetChatID(chat.ID).SetManagerID(managerID).Save(s.Ctx)
	s.Require().NoError(err)

	return chat.ID, p.ID
}
