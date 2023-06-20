package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/karasunokami/chat-service/internal/types"
)

// Chat holds the schema definition for the Chat entity.
type Chat struct {
	ent.Schema
}

// Fields of the Chat.
func (Chat) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.ChatID{}).Default(types.NewChatID).Unique().Immutable(),
		field.UUID("client_id", types.UserID{}).Unique().Immutable(),
		field.Time("created_at").Immutable().Default(defaultTime),
	}
}

// Edges of the Chat.
func (Chat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("messages", Message.Type),
		edge.To("problems", Problem.Type),
	}
}

func (Chat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("client_id"),
	}
}
