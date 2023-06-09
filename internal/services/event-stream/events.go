package eventstream

import (
	"fmt"
	"time"

	"github.com/karasunokami/chat-service/internal/types"
	"github.com/karasunokami/chat-service/internal/validator"
)

type Event interface {
	eventMarker()
	Validate() error
}

type event struct{}         //
func (*event) eventMarker() {}

// MessageSentEvent indicates that the message was checked by AFC
// and was sent to the manager. Two gray ticks.
type MessageSentEvent struct {
	event

	EventID   types.EventID   `validate:"required"`
	RequestID types.RequestID `validate:"required"`
	MessageID types.MessageID `validate:"required"`
}

func NewMessageSentEvent(
	eventID types.EventID,
	requestID types.RequestID,
	messageID types.MessageID,
) *MessageSentEvent {
	return &MessageSentEvent{
		EventID:   eventID,
		RequestID: requestID,
		MessageID: messageID,
	}
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
	event
	EventID     types.EventID   `validate:"required"`
	RequestID   types.RequestID `validate:"required"`
	ChatID      types.ChatID    `validate:"required"`
	MessageID   types.MessageID `validate:"required"`
	CreatedAt   time.Time       `validate:"required"`
	MessageBody string          `validate:"required"`
	AuthorID    types.UserID
	IsService   bool
}

func NewNewMessageEvent(
	eventID types.EventID,
	requestID types.RequestID,
	chatID types.ChatID,
	messageID types.MessageID,
	authorID types.UserID,
	createdAt time.Time,
	messageBody string,
	isService bool,
) *NewMessageEvent {
	return &NewMessageEvent{
		EventID:     eventID,
		RequestID:   requestID,
		ChatID:      chatID,
		MessageID:   messageID,
		AuthorID:    authorID,
		CreatedAt:   createdAt,
		MessageBody: messageBody,
		IsService:   isService,
	}
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
	event

	EventID   types.EventID   `validate:"required"`
	RequestID types.RequestID `validate:"required"`
	MessageID types.MessageID `validate:"required"`
}

func NewMessageBlockedEvent(
	eventID types.EventID,
	requestID types.RequestID,
	messageID types.MessageID,
) *MessageBlockedEvent {
	return &MessageBlockedEvent{
		EventID:   eventID,
		RequestID: requestID,
		MessageID: messageID,
	}
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
	event

	CanTakeMoreProblems bool            `validate:"required"`
	EventID             types.EventID   `validate:"required"`
	RequestID           types.RequestID `validate:"required"`
	ChatID              types.ChatID    `validate:"required"`
	ClientID            types.UserID    `validate:"required"`
}

func NewNewChatEvent(
	eventID types.EventID,
	requestID types.RequestID,
	chatID types.ChatID,
	clientID types.UserID,
	canTakeMoreProblems bool,
) *NewChatEvent {
	return &NewChatEvent{
		CanTakeMoreProblems: canTakeMoreProblems,
		EventID:             eventID,
		RequestID:           requestID,
		ChatID:              chatID,
		ClientID:            clientID,
	}
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
