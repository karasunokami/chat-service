package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/karasunokami/chat-service/internal/types"
)

// Problem holds the schema definition for the Problem entity.
type Problem struct {
	ent.Schema
}

// Fields of the Problem.
func (Problem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.ProblemID{}).Default(types.NewProblemID).Unique().Immutable(),
		field.UUID("chat_id", types.ChatID{}).Immutable(),
		field.UUID("manager_id", types.UserID{}).Optional(),
		field.Time("resolved_at").Optional(),
		field.Time("created_at").Default(defaultTime).Immutable(),
	}
}

// Edges of the Problem.
func (Problem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("chat", Chat.Type).
			Field("chat_id").
			Required().
			Immutable().
			Ref("problems").
			Unique(),

		edge.To("messages", Message.Type),
	}
}

func (Problem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("chat_id", "manager_id"),
	}
}
