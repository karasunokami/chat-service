package managerassignedtoproblemjob_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	messagesrepo "github.com/karasunokami/chat-service/internal/repositories/messages"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	msgproducer "github.com/karasunokami/chat-service/internal/services/msg-producer"
	managerassignedtoproblemjob "github.com/karasunokami/chat-service/internal/services/outbox/jobs/manager-assigned-to-problem"
	managerassignedtoproblemjobmocks "github.com/karasunokami/chat-service/internal/services/outbox/jobs/manager-assigned-to-problem/mocks"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestJob_Handle(t *testing.T) {
	// Arrange.
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgRepo := managerassignedtoproblemjobmocks.NewMockmessageRepository(ctrl)
	eventStream := managerassignedtoproblemjobmocks.NewMockeventStream(ctrl)
	msgProducer := managerassignedtoproblemjobmocks.NewMockmessageProducer(ctrl)
	job, err := managerassignedtoproblemjob.New(managerassignedtoproblemjob.NewOptions(msgProducer, eventStream, msgRepo))
	require.NoError(t, err)

	clientID := types.NewUserID()
	msgID := types.NewMessageID()
	serviceMsgID := types.NewMessageID()
	requestID := types.NewRequestID()
	eventID := types.NewEventID()
	chatID := types.NewChatID()
	mngID := types.NewUserID()
	problemID := types.NewProblemID()
	expectedBody := fmt.Sprintf(managerassignedtoproblemjob.ServiceMessageTpl, mngID)
	msgCreatedAt := time.Now()

	const (
		canTakeMoreProblem = true
		fromClient         = false
		isService          = true
	)

	msg := messagesrepo.Message{
		ID:               msgID,
		AuthorID:         clientID,
		InitialRequestID: requestID,
		CreatedAt:        msgCreatedAt,
		ChatID:           chatID,
		Body:             expectedBody,
		IsService:        isService,
	}
	serviceMsg := messagesrepo.Message{
		ID:               serviceMsgID,
		InitialRequestID: requestID,
		CreatedAt:        msgCreatedAt,
		ChatID:           chatID,
		Body:             expectedBody,
		IsService:        isService,
	}

	msgRepo.EXPECT().GetFirstProblemMessage(gomock.Any(), problemID).Return(&msg, nil)
	msgRepo.EXPECT().CreateService(gomock.Any(), problemID, chatID, expectedBody).Return(&serviceMsg, nil)

	msgProducer.EXPECT().ProduceMessage(gomock.Any(), msgproducer.Message{
		ID:         serviceMsgID,
		ChatID:     chatID,
		Body:       expectedBody,
		FromClient: fromClient,
	}).Return(nil)

	eventStream.EXPECT().Publish(ctx, clientID, &eventstream.NewMessageEvent{
		EventID:     eventID,
		RequestID:   requestID,
		ChatID:      chatID,
		MessageID:   serviceMsgID,
		AuthorID:    types.UserIDNil,
		MessageBody: expectedBody,
		IsService:   isService,
		CreatedAt:   msgCreatedAt,
	}).Return(nil)

	eventStream.EXPECT().Publish(ctx, mngID, &eventstream.NewChatEvent{
		EventID:             eventID,
		RequestID:           requestID,
		ChatID:              chatID,
		ClientID:            clientID,
		CanTakeMoreProblems: canTakeMoreProblem,
	}).Return(nil)

	// Action & assert.
	payload, err := managerassignedtoproblemjob.MarshalPayload(mngID, problemID, canTakeMoreProblem)
	require.NoError(t, err)

	err = job.Handle(ctx, payload)
	require.NoError(t, err)
}
