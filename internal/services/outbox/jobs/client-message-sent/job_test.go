package clientmessagesentjob_test

import (
	"context"
	"testing"
	"time"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/internal/services/outbox"
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

	msgRepo := clientmessagesentjobmocks.NewMockmessageRepo(ctrl)
	problemsRepo := clientmessagesentjobmocks.NewMockproblemsRepo(ctrl)
	eventStream := clientmessagesentjobmocks.NewMockeventStream(ctrl)
	job, err := clientmessagesentjob.New(clientmessagesentjob.NewOptions(eventStream, msgRepo, problemsRepo))
	require.NoError(t, err)

	clientID := types.NewUserID()
	msgID := types.NewMessageID()
	requestID := types.NewRequestID()
	eventID := types.NewEventID()
	chatID := types.NewChatID()
	mngID := types.NewUserID()
	problemID := types.NewProblemID()
	expectedBody := `Hi!`
	msgCreatedAt := time.Now()

	msg := messagesrepo.Message{
		ID:               msgID,
		AuthorID:         clientID,
		InitialRequestID: requestID,
		CreatedAt:        msgCreatedAt,
		ChatID:           chatID,
		Body:             expectedBody,
		ProblemID:        problemID,
	}

	msgRepo.EXPECT().GetMessageByID(gomock.Any(), msgID).Return(&msg, nil)
	problemsRepo.EXPECT().GetManagerID(gomock.Any(), problemID).Return(mngID, nil)

	eventStream.EXPECT().Publish(ctx, clientID, &eventstream.MessageSentEvent{
		RequestID: requestID,
		MessageID: msgID,
	}).Return(nil)

	eventStream.EXPECT().Publish(ctx, mngID, &eventstream.NewManagerMessageEvent{
		EventID:     eventID,
		RequestID:   requestID,
		ChatID:      chatID,
		MessageID:   msgID,
		CreatedAt:   msgCreatedAt,
		MessageBody: expectedBody,
		AuthorID:    clientID,
	}).Return(nil)

	// Action & assert.
	payload, err := outbox.MarshalMessageIDPayload(msgID)
	require.NoError(t, err)

	err = job.Handle(ctx, payload)
	require.NoError(t, err)
}
