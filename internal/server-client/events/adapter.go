package clientevents

import (
	"fmt"

	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
)

type Adapter struct{}

func (Adapter) Adapt(ev eventstream.Event) (any, error) {
	var event Event
	var err error

	switch v := ev.(type) {
	case *eventstream.NewMessageEvent:
		err = event.FromNewMessageEvent(NewMessageEvent{
			AuthorId:  v.UserID.AsPointer(),
			Body:      v.MessageBody,
			CreatedAt: v.CreatedAt,
			IsService: v.IsService,
			MessageId: v.MessageID,
			EventId:   v.EventID,
			RequestId: v.RequestID,
		})

	case *eventstream.MessageSentEvent:
		err = event.FromMessageSentEvent(MessageSentEvent{
			MessageId: v.MessageID,
			EventId:   v.EventID,
			RequestId: v.RequestID,
		})

	case *eventstream.MessageBlockedEvent:
		err = event.FromMessageBlockedEvent(MessageBlockedEvent{
			MessageId: v.MessageID,
			EventId:   v.EventID,
			RequestId: v.RequestID,
		})

	default:
		return nil, fmt.Errorf("unknown client event: %v (%T)", v, v)
	}

	if err != nil {
		return nil, err
	}

	return event, nil
}
