package gethistory_test

import (
	"errors"
	"testing"
	"time"

	"github.com/karasunokami/chat-service/internal/cursor"
	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	"github.com/karasunokami/chat-service/internal/testingh"
	"github.com/karasunokami/chat-service/internal/types"
	gethistory "github.com/karasunokami/chat-service/internal/usecases/manager/get-history"
	gethistorymocks "github.com/karasunokami/chat-service/internal/usecases/manager/get-history/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type UseCaseSuite struct {
	testingh.ContextSuite

	ctrl    *gomock.Controller
	msgRepo *gethistorymocks.MockmessagesRepository
	uCase   gethistory.UseCase
}

func TestUseCaseSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UseCaseSuite))
}

func (s *UseCaseSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.msgRepo = gethistorymocks.NewMockmessagesRepository(s.ctrl)

	var err error
	s.uCase, err = gethistory.New(gethistory.NewOptions(s.msgRepo))
	s.Require().NoError(err)

	s.ContextSuite.SetupTest()
}

func (s *UseCaseSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *UseCaseSuite) TestRequestValidationError() {
	// Arrange.
	req := gethistory.Request{}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, gethistory.ErrInvalidRequest)
	s.Empty(resp.Messages)
	s.Empty(resp.NextCursor)
}

func (s *UseCaseSuite) TestCursorDecodingError() {
	// Arrange.
	req := gethistory.Request{
		ID:        types.NewRequestID(),
		ManagerID: types.NewUserID(),
		ChatID:    types.NewChatID(),
		Cursor:    "eyJwYWdlX3NpemUiOjEwMA==", // {"page_size":100
	}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, gethistory.ErrInvalidCursor)
	s.Empty(resp.Messages)
	s.Empty(resp.NextCursor)
}

func (s *UseCaseSuite) TestGetManagerChatMessages_InvalidCursor() {
	// Arrange.
	managerID := types.NewUserID()
	chatID := types.NewChatID()

	c := messagesrepo.Cursor{PageSize: -1, LastCreatedAt: time.Now()}
	cursorWithNegativePageSize, err := cursor.Encode(c)
	s.Require().NoError(err)

	s.msgRepo.EXPECT().GetManagerChatMessages(s.Ctx, chatID, managerID, 0, messagesrepo.NewCursorMatcher(c)).
		Return(nil, nil, messagesrepo.ErrInvalidCursor)

	req := gethistory.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
		ChatID:    chatID,
		PageSize:  0,
		Cursor:    cursorWithNegativePageSize,
	}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.ErrorIs(err, gethistory.ErrInvalidCursor)
	s.Empty(resp.Messages)
	s.Empty(resp.NextCursor)
}

func (s *UseCaseSuite) TestGetManagerChatMessages_SomeError() {
	// Arrange.
	managerID := types.NewUserID()
	chatID := types.NewChatID()
	errExpected := errors.New("any error")

	s.msgRepo.EXPECT().GetManagerChatMessages(s.Ctx, chatID, managerID, 20, (*messagesrepo.Cursor)(nil)).
		Return(nil, nil, errExpected)

	req := gethistory.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
		ChatID:    chatID,
		PageSize:  20,
	}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)

	// Assert.
	s.Require().Error(err)
	s.Empty(resp.Messages)
	s.Empty(resp.NextCursor)
}

func (s *UseCaseSuite) TestGetManagerChatMessages_Success_SinglePage() {
	// Arrange.
	const messagesCount = 10
	const pageSize = messagesCount + 1

	chatID := types.NewChatID()
	clientID := types.NewUserID()
	managerID := types.NewUserID()
	expectedMsgs := s.createMessages(messagesCount, clientID, chatID)

	// Message.IsReceived logic:
	{
		// Processed by AFC and blocked.
		expectedMsgs[0].IsBlocked = true
		expectedMsgs[0].IsVisibleForManager = false

		// Processed by AFC and allowed.
		expectedMsgs[1].IsBlocked = false
		expectedMsgs[1].IsVisibleForManager = true

		// Not processed by AFC yet.
		expectedMsgs[2].IsBlocked = false
		expectedMsgs[2].IsVisibleForManager = false
	}

	s.msgRepo.EXPECT().GetManagerChatMessages(s.Ctx, chatID, managerID, pageSize, (*messagesrepo.Cursor)(nil)).
		Return(expectedMsgs, nil, nil)

	req := gethistory.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
		ChatID:    chatID,
		PageSize:  pageSize,
	}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)
	s.Require().NoError(err)

	// Assert.
	s.Empty(resp.NextCursor)

	s.Require().Len(resp.Messages, messagesCount)
	for i := 0; i < messagesCount; i++ {
		s.Equal(expectedMsgs[i].ID, resp.Messages[i].ID)
		s.Equal(expectedMsgs[i].AuthorID, resp.Messages[i].AuthorID)
		s.Equal(expectedMsgs[i].Body, resp.Messages[i].Body)
		s.Equal(expectedMsgs[i].CreatedAt.Unix(), resp.Messages[i].CreatedAt.Unix())
	}
}

func (s *UseCaseSuite) TestGetManagerChatMessages_Success_FirstPage() {
	// Arrange.
	const messagesCount = 10
	const pageSize = messagesCount + 1

	chatID := types.NewChatID()
	clientID := types.NewUserID()
	managerID := types.NewUserID()
	expectedMsgs := s.createMessages(messagesCount, clientID, chatID)
	lastMsg := expectedMsgs[len(expectedMsgs)-1]

	nextCursor := &messagesrepo.Cursor{PageSize: pageSize, LastCreatedAt: lastMsg.CreatedAt}
	s.msgRepo.EXPECT().GetManagerChatMessages(s.Ctx, chatID, managerID, pageSize, (*messagesrepo.Cursor)(nil)).
		Return(expectedMsgs, nextCursor, nil)

	req := gethistory.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
		ChatID:    chatID,
		PageSize:  pageSize,
	}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)
	s.Require().NoError(err)

	// Assert.
	s.NotEmpty(resp.NextCursor)
	s.Require().Len(resp.Messages, messagesCount)
}

func (s *UseCaseSuite) TestGetManagerChatMessages_Success_LastPage() {
	// Arrange.
	const messagesCount = 10
	const pageSize = messagesCount + 1

	chatID := types.NewChatID()
	clientID := types.NewUserID()
	managerID := types.NewUserID()
	expectedMsgs := s.createMessages(messagesCount, clientID, chatID)

	c := messagesrepo.Cursor{PageSize: pageSize, LastCreatedAt: time.Now()}
	s.msgRepo.EXPECT().GetManagerChatMessages(s.Ctx, chatID, managerID, 0, messagesrepo.NewCursorMatcher(c)).
		Return(expectedMsgs, nil, nil)

	cursorStr, err := cursor.Encode(c)
	s.Require().NoError(err)

	req := gethistory.Request{
		ID:        types.NewRequestID(),
		ManagerID: managerID,
		ChatID:    chatID,
		Cursor:    cursorStr,
	}

	// Action.
	resp, err := s.uCase.Handle(s.Ctx, req)
	s.Require().NoError(err)

	// Assert.
	s.Empty(resp.NextCursor)
	s.Require().Len(resp.Messages, messagesCount)
}

func (s *UseCaseSuite) createMessages(count int, authorID types.UserID, chatID types.ChatID) []messagesrepo.Message {
	s.T().Helper()

	result := make([]messagesrepo.Message, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, messagesrepo.Message{
			ID:                  types.NewMessageID(),
			ChatID:              chatID,
			AuthorID:            authorID,
			Body:                uuid.New().String(),
			CreatedAt:           time.Now(),
			IsVisibleForClient:  true,
			IsVisibleForManager: true,
			IsBlocked:           false,
			IsService:           false,
		})
	}
	return result
}
