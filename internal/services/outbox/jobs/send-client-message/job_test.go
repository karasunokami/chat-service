package sendclientmessagejob_test

import (
	"context"
	"testing"
	"time"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	msgproducer "github.com/karasunokami/chat-service/internal/services/msg-producer"
	sendclientmessagejob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/send-client-message"
	sendclientmessagejobmocks "github.com/karasunokami/chat-service/internal/services/outbox/jobs/send-client-message/mocks"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestJob_Handle(t *testing.T) {
	// Arrange.
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgProducer := sendclientmessagejobmocks.NewMockmessageProducer(ctrl)
	msgRepo := sendclientmessagejobmocks.NewMockmessageRepository(ctrl)
	eventStream := sendclientmessagejobmocks.NewMockeventStream(ctrl)
	job, err := sendclientmessagejob.New(sendclientmessagejob.NewOptions(msgProducer, msgRepo, eventStream))
	require.NoError(t, err)

	clientID := types.NewUserID()
	msgID := types.NewMessageID()
	chatID := types.NewChatID()
	requestID := types.NewRequestID()
	createdAt := time.Now()
	const isService = false
	const body = "Hello!"

	msg := messagesrepo.Message{
		ID:                  msgID,
		ChatID:              chatID,
		AuthorID:            clientID,
		InitialRequestID:    requestID,
		Body:                body,
		CreatedAt:           createdAt,
		IsVisibleForClient:  true,
		IsVisibleForManager: false,
		IsBlocked:           false,
		IsService:           isService,
	}
	msgRepo.EXPECT().GetMessageByID(gomock.Any(), msgID).Return(&msg, nil)

	msgProducer.EXPECT().ProduceMessage(gomock.Any(), msgproducer.Message{
		ID:         msgID,
		ChatID:     chatID,
		Body:       body,
		FromClient: true,
	}).Return(nil)

	eventStream.EXPECT().Publish(ctx, clientID, &eventstream.NewMessageEvent{
		RequestID:   requestID,
		ChatID:      chatID,
		MessageID:   msgID,
		UserID:      clientID,
		CreatedAt:   createdAt,
		MessageBody: body,
		IsService:   isService,
	}).Return(nil)

	// Action & assert.
	payload, err := sendclientmessagejob.MarshalPayload(msgID)
	require.NoError(t, err)

	err = job.Handle(ctx, payload)
	require.NoError(t, err)
}
