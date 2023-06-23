package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/karasunokami/chat-service/internal/types"
)

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.MessageID{}).Default(types.NewMessageID).Unique().Immutable(),
		field.UUID("chat_id", types.ChatID{}).Immutable(),
		field.UUID("problem_id", types.ProblemID{}),
		field.UUID("author_id", types.UserID{}).Optional(),
		field.UUID("initial_request_id", types.RequestID{}).Optional().Unique(),
		field.Bool("is_visible_for_client").Default(false),
		field.Bool("is_visible_for_manager").Default(false),
		field.Text("body").Immutable().MaxLen(3000).MinLen(1),
		field.Time("checked_at").Optional(),
		field.Bool("is_blocked").Default(false),
		field.Bool("is_service").Default(false),
		field.Time("created_at").Default(defaultTime).Immutable(),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("chat", Chat.Type).
			Field("chat_id").
			Required().
			Immutable().
			Ref("messages").
			Unique(),

		edge.From("problem", Problem.Type).
			Field("problem_id").
			Required().
			Ref("messages").
			Unique(),
	}
}

func (Message) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("initial_request_id").
			Unique().
			Annotations(entsql.IndexWhere("not is_service")),
		index.Fields("created_at", "chat_id"),
		index.Fields("created_at", "problem_id"),
	}
}
