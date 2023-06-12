package eventstream

import (
	"fmt"
	"time"

	"github.com/karasunokami/chat-service/internal/types"
	"github.com/karasunokami/chat-service/internal/validator"
)

//go:generate gonstructor --output=events.gen.go --type=NewMessageEvent --type=MessageSentEvent --type=MessageBlockedEvent --type=NewChatEvent --type=NewManagerMessageEvent

type Event interface {
	eventMarker()
	Validate() error
	Matches(x interface{}) bool
}
type event struct{}         //
func (*event) eventMarker() {}

// MessageSentEvent indicates that the message was checked by AFC
// and was sent to the manager. Two gray ticks.
type MessageSentEvent struct {
	event     `gonstructor:"-"`
	EventID   types.EventID   `validate:"required"`
	RequestID types.RequestID `validate:"required"`
	MessageID types.MessageID `validate:"required"`
}

func (e *MessageSentEvent) Validate() error {
	return validator.Validator.Struct(e)
}

func (e *MessageSentEvent) Matches(x interface{}) bool {
	ev, ok := x.(*MessageSentEvent)
	if !ok {
		return false
	}

	return ev.RequestID == e.RequestID && ev.MessageID == e.MessageID
}

func (e *MessageSentEvent) String() string {
	return fmt.Sprintf("{RequestID: %v, MessageID: %v}", e.RequestID, e.MessageID)
}

// NewMessageEvent is a signal about the appearance of a new message in the chat.
type NewMessageEvent struct {
	event       `gonstructor:"-"`
	EventID     types.EventID   `validate:"required"`
	RequestID   types.RequestID `validate:"required"`
	ChatID      types.ChatID    `validate:"required"`
	MessageID   types.MessageID `validate:"required"`
	CreatedAt   time.Time       `validate:"required"`
	MessageBody string          `validate:"required"`
	AuthorID    types.UserID
	IsService   bool
}

func (e *NewMessageEvent) Validate() error {
	return validator.Validator.Struct(e)
}

func (e *NewMessageEvent) Matches(x interface{}) bool {
	ev, ok := x.(*NewMessageEvent)
	if !ok {
		return false
	}

	return ev.RequestID == e.RequestID &&
		ev.ChatID == e.ChatID &&
		ev.MessageID == e.MessageID &&
		ev.AuthorID == e.AuthorID &&
		ev.CreatedAt == e.CreatedAt &&
		ev.MessageBody == e.MessageBody &&
		ev.IsService == e.IsService
}

func (e *NewMessageEvent) String() string {
	return fmt.Sprintf("%v", *e)
}

// MessageBlockedEvent indicates that the message was checked by AFC
// and marked as blocked.
type MessageBlockedEvent struct {
	event     `gonstructor:"-"`
	EventID   types.EventID   `validate:"required"`
	RequestID types.RequestID `validate:"required"`
	MessageID types.MessageID `validate:"required"`
}

func (e *MessageBlockedEvent) Validate() error {
	return validator.Validator.Struct(e)
}

func (e *MessageBlockedEvent) Matches(x interface{}) bool {
	ev, ok := x.(*MessageBlockedEvent)
	if !ok {
		return false
	}

	return ev.RequestID == e.RequestID && ev.MessageID == e.MessageID
}

func (e *MessageBlockedEvent) String() string {
	return fmt.Sprintf("{RequestID: %v, MessageID: %v}", e.RequestID, e.MessageID)
}

// Manager Events

// NewChatEvent is a signal about the appearance of a new chat for manager.
type NewChatEvent struct {
	event               `gonstructor:"-"`
	CanTakeMoreProblems bool            `validate:"boolean"`
	EventID             types.EventID   `validate:"required"`
	RequestID           types.RequestID `validate:"required"`
	ChatID              types.ChatID    `validate:"required"`
	ClientID            types.UserID    `validate:"required"`
}

func (e *NewChatEvent) Validate() error {
	return validator.Validator.Struct(e)
}

func (e *NewChatEvent) Matches(x interface{}) bool {
	ev, ok := x.(*NewChatEvent)
	if !ok {
		return false
	}

	return ev.RequestID == e.RequestID &&
		ev.ChatID == e.ChatID &&
		ev.ClientID == e.ClientID &&
		ev.CanTakeMoreProblems == e.CanTakeMoreProblems
}

func (e *NewChatEvent) String() string {
	return fmt.Sprintf("%v", *e)
}

// NewManagerMessageEvent is a signal about the appearance of a new manager message in the chat.
type NewManagerMessageEvent struct {
	event       `gonstructor:"-"`
	EventID     types.EventID   `validate:"required"`
	RequestID   types.RequestID `validate:"required"`
	ChatID      types.ChatID    `validate:"required"`
	MessageID   types.MessageID `validate:"required"`
	CreatedAt   time.Time       `validate:"required"`
	MessageBody string          `validate:"required"`
	AuthorID    types.UserID    `validate:"required"`
}

func (e *NewManagerMessageEvent) Validate() error {
	return validator.Validator.Struct(e)
}

func (e *NewManagerMessageEvent) Matches(x interface{}) bool {
	ev, ok := x.(*NewManagerMessageEvent)
	if !ok {
		return false
	}

	return ev.RequestID == e.RequestID &&
		ev.ChatID == e.ChatID &&
		ev.MessageID == e.MessageID &&
		ev.CreatedAt == e.CreatedAt &&
		ev.MessageBody == e.MessageBody &&
		ev.AuthorID == e.AuthorID
}

func (e *NewManagerMessageEvent) String() string {
	return fmt.Sprintf("%v", *e)
}
