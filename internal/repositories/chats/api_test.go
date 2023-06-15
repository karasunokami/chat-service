//go:build integration

package chatsrepo_test

import (
	"context"
	"testing"
	"time"

	chatsrepo "github.com/karasunokami/chat-service/internal/repositories/chats"
	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/stretchr/testify/suite"
)

type ChatsRepoSuite struct {
	testingh.DBSuite
	repo *chatsrepo.Repo
}

func TestChatsRepoSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &ChatsRepoSuite{DBSuite: testingh.NewDBSuite("TestChatsRepoSuite")})
}

func (s *ChatsRepoSuite) SetupSuite() {
	s.DBSuite.SetupSuite()

	var err error

	s.repo, err = chatsrepo.New(chatsrepo.NewOptions(s.Database))
	s.Require().NoError(err)
}

func (s *ChatsRepoSuite) Test_CreateIfNotExists() {
	s.Run("chat does not exist, should be created", func() {
		clientID := types.NewUserID()

		chatID, err := s.repo.CreateIfNotExists(s.Ctx, clientID)
		s.Require().NoError(err)
		s.NotEmpty(chatID)
	})

	s.Run("chat already exists", func() {
		clientID := types.NewUserID()

		// Create chat.
		chat, err := s.Database.Chat(s.Ctx).Create().SetClientID(clientID).Save(s.Ctx)
		s.Require().NoError(err)

		chatID, err := s.repo.CreateIfNotExists(s.Ctx, clientID)
		s.Require().NoError(err)
		s.Require().NotEmpty(chatID)
		s.Equal(chat.ID, chatID)
	})
}

func (s *ChatsRepoSuite) Test_GetManagerOpened() {
	s.Run("opened chats with manager exists", func() {
		managerID := types.NewUserID()

		s.createChatWithProblem(s.Ctx, types.NewUserID(), managerID, true)
		s.createChatWithProblem(s.Ctx, types.NewUserID(), managerID, true)

		chats, err := s.repo.GetManagerOpened(s.Ctx, managerID)
		s.Require().NoError(err)
		s.Len(chats, 2)
	})

	s.Run("closed chats with manager exists", func() {
		managerID := types.NewUserID()

		s.createChatWithProblem(s.Ctx, types.NewUserID(), managerID, false)
		s.createChatWithProblem(s.Ctx, types.NewUserID(), managerID, false)

		chats, err := s.repo.GetManagerOpened(s.Ctx, managerID)
		s.Require().NoError(err)
		s.Len(chats, 0)
	})

	s.Run("opened chats with another manager exists", func() {
		managerID := types.NewUserID()

		s.createChatWithProblem(s.Ctx, types.NewUserID(), types.NewUserID(), true)
		s.createChatWithProblem(s.Ctx, types.NewUserID(), types.NewUserID(), true)

		chats, err := s.repo.GetManagerOpened(s.Ctx, managerID)
		s.Require().NoError(err)
		s.Len(chats, 0)
	})

	s.Run("closed chats with another manager exists", func() {
		managerID := types.NewUserID()

		s.createChatWithProblem(s.Ctx, types.NewUserID(), types.NewUserID(), false)
		s.createChatWithProblem(s.Ctx, types.NewUserID(), types.NewUserID(), false)

		chats, err := s.repo.GetManagerOpened(s.Ctx, managerID)
		s.Require().NoError(err)
		s.Len(chats, 0)
	})

	s.Run("has chat with manager", func() {
		managerID := types.NewUserID()

		_, err := s.Database.Chat(s.Ctx).Create().SetClientID(managerID).Save(s.Ctx)
		s.Require().NoError(err)

		chats, err := s.repo.GetManagerOpened(s.Ctx, managerID)
		s.Require().NoError(err)
		s.Len(chats, 0)
	})
}

func (s *ChatsRepoSuite) Test_GetClientID() {
	s.Run("chat does not exist", func() {
		clientID, err := s.repo.GetClientID(s.Ctx, types.NewChatID())
		s.Require().Error(err)
		s.Require().ErrorIs(err, chatsrepo.ErrNotFound)
		s.Empty(clientID)
	})

	s.Run("chat already exists", func() {
		chatClientID := types.NewUserID()

		// Create chat.
		chat, err := s.Database.Chat(s.Ctx).Create().SetClientID(chatClientID).Save(s.Ctx)
		s.Require().NoError(err)

		clientID, err := s.repo.GetClientID(s.Ctx, chat.ID)
		s.Require().NoError(err)
		s.Require().NotEmpty(clientID)
		s.Equal(clientID, chatClientID)
	})
}

func (s *ChatsRepoSuite) createChatWithProblem(
	ctx context.Context,
	clientID, managerID types.UserID,
	opened bool,
) (types.ChatID, types.ProblemID) {
	c, err := s.Database.Chat(ctx).Create().SetClientID(clientID).Save(ctx)
	s.Require().NoError(err)

	pq := s.Database.Problem(ctx).Create().
		SetManagerID(managerID).
		SetChatID(c.ID)

	if !opened {
		pq.SetResolvedAt(time.Now())
	}

	p, err := pq.Save(ctx)
	s.Require().NoError(err)

	return c.ID, p.ID
}
