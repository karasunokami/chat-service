package managerevents

import (
	"fmt"

	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
)

type Adapter struct{}

func (Adapter) Adapt(ev eventstream.Event) (any, error) {
	if e, ok := ev.(*eventstream.NewChatEvent); ok {
		return adaptNewChatEvent(e), nil
	}

	return nil, fmt.Errorf("unknown event type: %T", ev)
}

func adaptNewChatEvent(ev *eventstream.NewChatEvent) NewChatEvent {
	return NewChatEvent{
		CanTakeMoreProblems: ev.CanTakeMoreProblems,
		ChatId:              ev.ChatID,
		ClientId:            ev.ClientID,
		EventId:             ev.EventID,
		EventType:           "NewChatEvent",
		RequestId:           ev.RequestID,
	}
}
