package clientmessageblockedjob_test

import (
	"context"
	"testing"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/internal/services/outbox"
	clientmessageblockedjob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/client-message-blocked"
	clientmessageblockedjobmocks "github.com/karasunokami/chat-service/internal/services/outbox/jobs/client-message-blocked/mocks"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestJob_Handle(t *testing.T) {
	// Arrange.
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgRepo := clientmessageblockedjobmocks.NewMockmessageRepository(ctrl)
	eventStream := clientmessageblockedjobmocks.NewMockeventStream(ctrl)
	job, err := clientmessageblockedjob.New(clientmessageblockedjob.NewOptions(eventStream, msgRepo))
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

	eventStream.EXPECT().Publish(ctx, clientID, &eventstream.MessageBlockedEvent{
		RequestID: requestID,
		MessageID: msgID,
		EventID:   eventID,
	}).Return(nil)

	// Action & assert.
	payload, err := outbox.MarshalMessageIDPayload(msgID)
	require.NoError(t, err)

	err = job.Handle(ctx, payload)
	require.NoError(t, err)
}
