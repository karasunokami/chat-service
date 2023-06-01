package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/karasunokami/chat-service/internal/types"
)

// jobMaxAttempts is some limit as protection from endless retries of outbox jobs.
const jobMaxAttempts = 30

type Job struct {
	ent.Schema
}

func (Job) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.JobID{}).Default(types.NewJobID).Unique().Immutable(),
		field.String("name").NotEmpty().Immutable(),
		field.Text("payload").NotEmpty().Immutable(),
		field.Int("attempts").Min(0).Max(jobMaxAttempts).Default(0),
		field.Time("available_at").Immutable(),
		field.Time("reserved_until").Default(defaultTime()),
		field.Time("created_at").Immutable().Default(defaultTime()),
	}
}

func (Job) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("available_at", "reserved_until"),
	}
}

type FailedJob struct {
	ent.Schema
}

func (FailedJob) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.FailedJobID{}).Default(types.NewFailedJobID).Unique().Immutable(),
		field.String("name").NotEmpty().Immutable(),
		field.Text("payload").NotEmpty().Immutable(),
		field.String("reason").NotEmpty().Immutable(),
		field.Time("created_at").Immutable().Default(defaultTime()),
	}
}
