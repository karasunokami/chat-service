package managerevents

import (
	"fmt"

	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
)

type Adapter struct{}

func (Adapter) Adapt(ev eventstream.Event) (any, error) {
	var event Event
	var err error

	switch v := ev.(type) {
	case *eventstream.NewChatEvent:
		err = event.FromNewChatEvent(NewChatEvent{
			CanTakeMoreProblems: v.CanTakeMoreProblems,
			ChatId:              v.ChatID,
			ClientId:            v.ClientID,
			EventId:             v.EventID,
			RequestId:           v.RequestID,
		})

	case *eventstream.NewManagerMessageEvent:
		err = event.FromNewMessageEvent(NewMessageEvent{
			AuthorId:  v.AuthorID,
			Body:      v.MessageBody,
			ChatId:    v.ChatID,
			CreatedAt: v.CreatedAt,
			EventId:   v.EventID,
			MessageId: v.MessageID,
			RequestId: v.RequestID,
		})

	case *eventstream.ChatClosedEvent:
		err = event.FromChatClosedEvent(ChatClosedEvent{
			CanTakeMoreProblems: v.CanTakeMoreProblems,
			ChatId:              v.ChatID,
			EventId:             v.EventID,
			RequestId:           v.RequestID,
		})

	default:
		return nil, fmt.Errorf("unknown manager event: %v (%T)", v, v)
	}

	if err != nil {
		return nil, err
	}

	return event, nil
}
