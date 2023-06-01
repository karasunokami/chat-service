package clientmessagesentjob_test

import (
	"context"
	"testing"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	clientmessagesentjob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/client-message-sent"
	clientmessagesentjobmocks "github.com/karasunokami/chat-service/internal/services/outbox/jobs/client-message-sent/mocks"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestJob_Handle(t *testing.T) {
	// Arrange.
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgRepo := clientmessagesentjobmocks.NewMockmessageRepository(ctrl)
	eventStream := clientmessagesentjobmocks.NewMockeventStream(ctrl)
	job, err := clientmessagesentjob.New(clientmessagesentjob.NewOptions(eventStream, msgRepo))
	require.NoError(t, err)

	clientID := types.NewUserID()
	msgID := types.NewMessageID()
	requestID := types.NewRequestID()
	eventID := types.NewEventID()

	msg := messagesrepo.Message{
		ID:               msgID,
		AuthorID:         clientID,
		InitialRequestID: requestID,
	}

	msgRepo.EXPECT().GetMessageByID(gomock.Any(), msgID).Return(&msg, nil)

	eventStream.EXPECT().Publish(ctx, clientID, &eventstream.MessageSentEvent{
		RequestID: requestID,
		MessageID: msgID,
		EventID:   eventID,
	}).Return(nil)

	// Action & assert.
	payload, err := clientmessagesentjob.MarshalPayload(msgID)
	require.NoError(t, err)

	err = job.Handle(ctx, payload)
	require.NoError(t, err)
}
