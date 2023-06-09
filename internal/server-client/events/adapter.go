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

	case *eventstream.MessageBlockedEvent:
		return adaptMessageBlockedEvent(e), nil
	}

	return nil, fmt.Errorf("unknown event type: %T", ev)
}

func adaptMessageSendEvent(ev *eventstream.MessageSentEvent) DefaultEvent {
	return DefaultEvent{
		EventId:   ev.EventID,
		EventType: "MessageSentEvent",
		MessageId: ev.MessageID,
		RequestId: ev.RequestID,
	}
}

func adaptMessageBlockedEvent(ev *eventstream.MessageBlockedEvent) DefaultEvent {
	return DefaultEvent{
		EventId:   ev.EventID,
		EventType: "MessageBlockedEvent",
		MessageId: ev.MessageID,
		RequestId: ev.RequestID,
	}
}

func adaptNewMessageEvent(ev *eventstream.NewMessageEvent) NewMessageEvent {
	return NewMessageEvent{
		AuthorId:  pointer.PtrWithZeroAsNil(ev.AuthorID),
		Body:      ev.MessageBody,
		CreatedAt: ev.CreatedAt,
		EventId:   ev.EventID,
		EventType: "NewMessageEvent",
		IsService: ev.IsService,
		MessageId: ev.MessageID,
		RequestId: ev.RequestID,
	}
}
