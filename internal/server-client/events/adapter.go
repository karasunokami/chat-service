package clientevents

import (
	"fmt"

	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/pkg/pointer"
)

type Adapter struct{}

func (Adapter) Adapt(ev eventstream.Event) (any, error) {
	switch e := ev.(type) {
	case *eventstream.MessageSentEvent:
		return adaptMessageSendEvent(e), nil

	case *eventstream.NewMessageEvent:
		return adaptNewMessageEvent(e), nil
	}

	return nil, fmt.Errorf("unknown event type: %T", ev)
}

func adaptMessageSendEvent(ev *eventstream.MessageSentEvent) MessageSentEvent {
	return MessageSentEvent{
		EventId:   ev.EventID,
		EventType: "MessageSentEvent",
		MessageId: ev.MessageID,
		RequestId: ev.RequestID,
	}
}

func adaptNewMessageEvent(ev *eventstream.NewMessageEvent) NewMessageEvent {
	return NewMessageEvent{
		AuthorId:  ev.UserID.AsPointer(),
		Body:      pointer.Ptr(ev.MessageBody),
		CreatedAt: pointer.Ptr(ev.CreatedAt),
		EventId:   ev.EventID,
		EventType: "NewMessageEvent",
		IsService: pointer.Ptr(ev.IsService),
		MessageId: ev.MessageID,
		RequestId: ev.RequestID,
	}
}
