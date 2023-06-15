package chatclosed_test

import (
	"context"
	"testing"
	"time"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	msgproducer "github.com/karasunokami/chat-service/internal/services/msg-producer"
	chatclosedjob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/chat-closed"
	chatclosedjobmocks "github.com/karasunokami/chat-service/internal/services/outbox/jobs/chat-closed/mocks"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestJob_Handle(t *testing.T) {
	// Arrange.
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgProducer := chatclosedjobmocks.NewMockmessageProducer(ctrl)
	msgRepo := chatclosedjobmocks.NewMockmessagesRepository(ctrl)
	chatsRepo := chatclosedjobmocks.NewMockchatsRepository(ctrl)
	eventStream := chatclosedjobmocks.NewMockeventStream(ctrl)
	managerLoad := chatclosedjobmocks.NewMockmanagerLoadService(ctrl)
	job, err := chatclosedjob.New(chatclosedjob.NewOptions(
		msgProducer,
		msgRepo,
		chatsRepo,
		eventStream,
		managerLoad,
	))
	require.NoError(t, err)

	clientID := types.NewUserID()
	managerID := types.NewUserID()
	msgID := types.NewMessageID()
	chatID := types.NewChatID()
	requestID := types.NewRequestID()
	createdAt := time.Now()
	body := `body`

	const (
		canTakeMoreProblem = true
		isService          = true
		fromClient         = false
	)

	msg := messagesrepo.Message{
		ID:               msgID,
		AuthorID:         clientID,
		InitialRequestID: requestID,
		CreatedAt:        createdAt,
		Body:             body,
		ChatID:           chatID,
		IsService:        isService,
	}

	msgRepo.EXPECT().GetMessageByID(gomock.Any(), msgID).Return(&msg, nil)

	msgProducer.EXPECT().ProduceMessage(ctx, msgproducer.Message{
		ID:         msgID,
		ChatID:     chatID,
		Body:       body,
		FromClient: fromClient,
	}).Return(nil)

	chatsRepo.EXPECT().GetClientID(ctx, chatID).Return(clientID, nil)

	eventStream.EXPECT().Publish(ctx, clientID, eventstream.NewNewMessageEvent(
		types.NewEventID(),
		requestID,
		chatID,
		msgID,
		createdAt,
		body,
		types.UserIDNil,
		isService,
	)).Return(nil)

	managerLoad.EXPECT().CanManagerTakeProblem(ctx, managerID).Return(canTakeMoreProblem, nil)

	eventStream.EXPECT().Publish(ctx, managerID, eventstream.NewChatClosedEvent(
		canTakeMoreProblem,
		chatID,
		types.NewEventID(),
		requestID,
	))

	// Action & assert.
	payload, err := chatclosedjob.MarshalPayload(managerID, msgID, requestID)
	require.NoError(t, err)

	err = job.Handle(ctx, payload)
	require.NoError(t, err)
}
