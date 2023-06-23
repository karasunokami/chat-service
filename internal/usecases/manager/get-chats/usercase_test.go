package getchats_test

import (
	"io"
	"testing"

	chatsrepo "github.com/karasunokami/chat-service/internal/repositories/chats"
	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"
	getchats "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats"
	getchatsmocks "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type UseCaseSuite struct {
	testingh.ContextSuite

	ctrl      *gomock.Controller
	chatsRepo *getchatsmocks.MockchatsRepo
	uCase     getchats.UseCase
}

func TestUseCaseSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UseCaseSuite))
}

func (s *UseCaseSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.chatsRepo = getchatsmocks.NewMockchatsRepo(s.ctrl)

	var err error
	s.uCase, err = getchats.New(getchats.NewOptions(s.chatsRepo))
	s.Require().NoError(err)

	s.ContextSuite.SetupTest()
}

func (s *UseCaseSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *UseCaseSuite) TestRequestValidationError() {
	// Arrange.
	req := getchats.Request{}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, getchats.ErrInvalidRequest)
	s.Empty(resp.Chats)
}

func (s *UseCaseSuite) TestGetChatsWithOpenProblemsError() {
	// Arrange.

	managerID := types.NewUserID()
	expectedError := io.EOF

	s.chatsRepo.EXPECT().GetManagerOpened(s.Ctx, managerID).
		Return(nil, expectedError)

	req := getchats.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	// Action.

	resp, err := s.uCase.Handle(s.Ctx, req)
	s.Require().Error(err)

	// Assert.

	s.Require().ErrorIs(err, expectedError)
	s.Require().Len(resp.Chats, 0)
	s.Require().Nil(resp.Chats)
}

func (s *UseCaseSuite) TestSuccessStory() {
	// Arrange.

	managerID := types.NewUserID()
	clientID := types.NewUserID()
	chatID1 := types.NewChatID()
	chatID2 := types.NewChatID()

	expectedChats := []chatsrepo.Chat{
		{
			ID:       chatID1,
			ClientID: clientID,
		},
		{
			ID:       chatID2,
			ClientID: clientID,
		},
	}

	s.chatsRepo.EXPECT().GetManagerOpened(s.Ctx, managerID).
		Return(expectedChats, nil)

	req := getchats.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
	}

	// Action.

	resp, err := s.uCase.Handle(s.Ctx, req)
	s.Require().NoError(err)

	// Assert.

	s.Require().Len(resp.Chats, len(expectedChats))

	for i := 0; i < len(expectedChats); i++ {
		s.Equal(expectedChats[i].ID, resp.Chats[i].ID)
		s.Equal(expectedChats[i].ClientID, resp.Chats[i].ClientID)
	}
}
