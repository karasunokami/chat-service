// Code generated by ent, DO NOT EDIT.

package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/karasunokami/chat-service/internal/store/job"
	"github.com/karasunokami/chat-service/internal/types"
)

// JobCreate is the builder for creating a Job entity.
type JobCreate struct {
	config
	mutation *JobMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetName sets the "name" field.
func (jc *JobCreate) SetName(s string) *JobCreate {
	jc.mutation.SetName(s)
	return jc
}

// SetPayload sets the "payload" field.
func (jc *JobCreate) SetPayload(s string) *JobCreate {
	jc.mutation.SetPayload(s)
	return jc
}

// SetAttempts sets the "attempts" field.
func (jc *JobCreate) SetAttempts(i int) *JobCreate {
	jc.mutation.SetAttempts(i)
	return jc
}

// SetNillableAttempts sets the "attempts" field if the given value is not nil.
func (jc *JobCreate) SetNillableAttempts(i *int) *JobCreate {
	if i != nil {
		jc.SetAttempts(*i)
	}
	return jc
}

// SetAvailableAt sets the "available_at" field.
func (jc *JobCreate) SetAvailableAt(t time.Time) *JobCreate {
	jc.mutation.SetAvailableAt(t)
	return jc
}

// SetReservedUntil sets the "reserved_until" field.
func (jc *JobCreate) SetReservedUntil(t time.Time) *JobCreate {
	jc.mutation.SetReservedUntil(t)
	return jc
}

// SetNillableReservedUntil sets the "reserved_until" field if the given value is not nil.
func (jc *JobCreate) SetNillableReservedUntil(t *time.Time) *JobCreate {
	if t != nil {
		jc.SetReservedUntil(*t)
	}
	return jc
}

// SetCreatedAt sets the "created_at" field.
func (jc *JobCreate) SetCreatedAt(t time.Time) *JobCreate {
	jc.mutation.SetCreatedAt(t)
	return jc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (jc *JobCreate) SetNillableCreatedAt(t *time.Time) *JobCreate {
	if t != nil {
		jc.SetCreatedAt(*t)
	}
	return jc
}

// SetID sets the "id" field.
func (jc *JobCreate) SetID(ti types.JobID) *JobCreate {
	jc.mutation.SetID(ti)
	return jc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (jc *JobCreate) SetNillableID(ti *types.JobID) *JobCreate {
	if ti != nil {
		jc.SetID(*ti)
	}
	return jc
}

// Mutation returns the JobMutation object of the builder.
func (jc *JobCreate) Mutation() *JobMutation {
	return jc.mutation
}

// Save creates the Job in the database.
func (jc *JobCreate) Save(ctx context.Context) (*Job, error) {
	jc.defaults()
	return withHooks[*Job, JobMutation](ctx, jc.sqlSave, jc.mutation, jc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (jc *JobCreate) SaveX(ctx context.Context) *Job {
	v, err := jc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (jc *JobCreate) Exec(ctx context.Context) error {
	_, err := jc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (jc *JobCreate) ExecX(ctx context.Context) {
	if err := jc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (jc *JobCreate) defaults() {
	if _, ok := jc.mutation.Attempts(); !ok {
		v := job.DefaultAttempts
		jc.mutation.SetAttempts(v)
	}
	if _, ok := jc.mutation.ReservedUntil(); !ok {
		v := job.DefaultReservedUntil
		jc.mutation.SetReservedUntil(v)
	}
	if _, ok := jc.mutation.CreatedAt(); !ok {
		v := job.DefaultCreatedAt
		jc.mutation.SetCreatedAt(v)
	}
	if _, ok := jc.mutation.ID(); !ok {
		v := job.DefaultID()
		jc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (jc *JobCreate) check() error {
	if _, ok := jc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`store: missing required field "Job.name"`)}
	}
	if v, ok := jc.mutation.Name(); ok {
		if err := job.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`store: validator failed for field "Job.name": %w`, err)}
		}
	}
	if _, ok := jc.mutation.Payload(); !ok {
		return &ValidationError{Name: "payload", err: errors.New(`store: missing required field "Job.payload"`)}
	}
	if v, ok := jc.mutation.Payload(); ok {
		if err := job.PayloadValidator(v); err != nil {
			return &ValidationError{Name: "payload", err: fmt.Errorf(`store: validator failed for field "Job.payload": %w`, err)}
		}
	}
	if _, ok := jc.mutation.Attempts(); !ok {
		return &ValidationError{Name: "attempts", err: errors.New(`store: missing required field "Job.attempts"`)}
	}
	if v, ok := jc.mutation.Attempts(); ok {
		if err := job.AttemptsValidator(v); err != nil {
			return &ValidationError{Name: "attempts", err: fmt.Errorf(`store: validator failed for field "Job.attempts": %w`, err)}
		}
	}
	if _, ok := jc.mutation.AvailableAt(); !ok {
		return &ValidationError{Name: "available_at", err: errors.New(`store: missing required field "Job.available_at"`)}
	}
	if _, ok := jc.mutation.ReservedUntil(); !ok {
		return &ValidationError{Name: "reserved_until", err: errors.New(`store: missing required field "Job.reserved_until"`)}
	}
	if _, ok := jc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`store: missing required field "Job.created_at"`)}
	}
	if v, ok := jc.mutation.ID(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`store: validator failed for field "Job.id": %w`, err)}
		}
	}
	return nil
}

func (jc *JobCreate) sqlSave(ctx context.Context) (*Job, error) {
	if err := jc.check(); err != nil {
		return nil, err
	}
	_node, _spec := jc.createSpec()
	if err := sqlgraph.CreateNode(ctx, jc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*types.JobID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	jc.mutation.id = &_node.ID
	jc.mutation.done = true
	return _node, nil
}

func (jc *JobCreate) createSpec() (*Job, *sqlgraph.CreateSpec) {
	var (
		_node = &Job{config: jc.config}
		_spec = sqlgraph.NewCreateSpec(job.Table, sqlgraph.NewFieldSpec(job.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = jc.conflict
	if id, ok := jc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := jc.mutation.Name(); ok {
		_spec.SetField(job.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := jc.mutation.Payload(); ok {
		_spec.SetField(job.FieldPayload, field.TypeString, value)
		_node.Payload = value
	}
	if value, ok := jc.mutation.Attempts(); ok {
		_spec.SetField(job.FieldAttempts, field.TypeInt, value)
		_node.Attempts = value
	}
	if value, ok := jc.mutation.AvailableAt(); ok {
		_spec.SetField(job.FieldAvailableAt, field.TypeTime, value)
		_node.AvailableAt = value
	}
	if value, ok := jc.mutation.ReservedUntil(); ok {
		_spec.SetField(job.FieldReservedUntil, field.TypeTime, value)
		_node.ReservedUntil = value
	}
	if value, ok := jc.mutation.CreatedAt(); ok {
		_spec.SetField(job.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Job.Create().
//		SetName(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.JobUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (jc *JobCreate) OnConflict(opts ...sql.ConflictOption) *JobUpsertOne {
	jc.conflict = opts
	return &JobUpsertOne{
		create: jc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Job.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (jc *JobCreate) OnConflictColumns(columns ...string) *JobUpsertOne {
	jc.conflict = append(jc.conflict, sql.ConflictColumns(columns...))
	return &JobUpsertOne{
		create: jc,
	}
}

type (
	// JobUpsertOne is the builder for "upsert"-ing
	//  one Job node.
	JobUpsertOne struct {
		create *JobCreate
	}

	// JobUpsert is the "OnConflict" setter.
	JobUpsert struct {
		*sql.UpdateSet
	}
)

// SetAttempts sets the "attempts" field.
func (u *JobUpsert) SetAttempts(v int) *JobUpsert {
	u.Set(job.FieldAttempts, v)
	return u
}

// UpdateAttempts sets the "attempts" field to the value that was provided on create.
func (u *JobUpsert) UpdateAttempts() *JobUpsert {
	u.SetExcluded(job.FieldAttempts)
	return u
}

// AddAttempts adds v to the "attempts" field.
func (u *JobUpsert) AddAttempts(v int) *JobUpsert {
	u.Add(job.FieldAttempts, v)
	return u
}

// SetReservedUntil sets the "reserved_until" field.
func (u *JobUpsert) SetReservedUntil(v time.Time) *JobUpsert {
	u.Set(job.FieldReservedUntil, v)
	return u
}

// UpdateReservedUntil sets the "reserved_until" field to the value that was provided on create.
func (u *JobUpsert) UpdateReservedUntil() *JobUpsert {
	u.SetExcluded(job.FieldReservedUntil)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Job.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(job.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *JobUpsertOne) UpdateNewValues() *JobUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(job.FieldID)
		}
		if _, exists := u.create.mutation.Name(); exists {
			s.SetIgnore(job.FieldName)
		}
		if _, exists := u.create.mutation.Payload(); exists {
			s.SetIgnore(job.FieldPayload)
		}
		if _, exists := u.create.mutation.AvailableAt(); exists {
			s.SetIgnore(job.FieldAvailableAt)
		}
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(job.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Job.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *JobUpsertOne) Ignore() *JobUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *JobUpsertOne) DoNothing() *JobUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the JobCreate.OnConflict
// documentation for more info.
func (u *JobUpsertOne) Update(set func(*JobUpsert)) *JobUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&JobUpsert{UpdateSet: update})
	}))
	return u
}

// SetAttempts sets the "attempts" field.
func (u *JobUpsertOne) SetAttempts(v int) *JobUpsertOne {
	return u.Update(func(s *JobUpsert) {
		s.SetAttempts(v)
	})
}

// AddAttempts adds v to the "attempts" field.
func (u *JobUpsertOne) AddAttempts(v int) *JobUpsertOne {
	return u.Update(func(s *JobUpsert) {
		s.AddAttempts(v)
	})
}

// UpdateAttempts sets the "attempts" field to the value that was provided on create.
func (u *JobUpsertOne) UpdateAttempts() *JobUpsertOne {
	return u.Update(func(s *JobUpsert) {
		s.UpdateAttempts()
	})
}

// SetReservedUntil sets the "reserved_until" field.
func (u *JobUpsertOne) SetReservedUntil(v time.Time) *JobUpsertOne {
	return u.Update(func(s *JobUpsert) {
		s.SetReservedUntil(v)
	})
}

// UpdateReservedUntil sets the "reserved_until" field to the value that was provided on create.
func (u *JobUpsertOne) UpdateReservedUntil() *JobUpsertOne {
	return u.Update(func(s *JobUpsert) {
		s.UpdateReservedUntil()
	})
}

// Exec executes the query.
func (u *JobUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("store: missing options for JobCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *JobUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *JobUpsertOne) ID(ctx context.Context) (id types.JobID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("store: JobUpsertOne.ID is not supported by MySQL driver. Use JobUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *JobUpsertOne) IDX(ctx context.Context) types.JobID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// JobCreateBulk is the builder for creating many Job entities in bulk.
type JobCreateBulk struct {
	config
	builders []*JobCreate
	conflict []sql.ConflictOption
}

// Save creates the Job entities in the database.
func (jcb *JobCreateBulk) Save(ctx context.Context) ([]*Job, error) {
	specs := make([]*sqlgraph.CreateSpec, len(jcb.builders))
	nodes := make([]*Job, len(jcb.builders))
	mutators := make([]Mutator, len(jcb.builders))
	for i := range jcb.builders {
		func(i int, root context.Context) {
			builder := jcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*JobMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, jcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = jcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, jcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, jcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (jcb *JobCreateBulk) SaveX(ctx context.Context) []*Job {
	v, err := jcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (jcb *JobCreateBulk) Exec(ctx context.Context) error {
	_, err := jcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (jcb *JobCreateBulk) ExecX(ctx context.Context) {
	if err := jcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Job.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.JobUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (jcb *JobCreateBulk) OnConflict(opts ...sql.ConflictOption) *JobUpsertBulk {
	jcb.conflict = opts
	return &JobUpsertBulk{
		create: jcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Job.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (jcb *JobCreateBulk) OnConflictColumns(columns ...string) *JobUpsertBulk {
	jcb.conflict = append(jcb.conflict, sql.ConflictColumns(columns...))
	return &JobUpsertBulk{
		create: jcb,
	}
}

// JobUpsertBulk is the builder for "upsert"-ing
// a bulk of Job nodes.
type JobUpsertBulk struct {
	create *JobCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Job.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(job.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *JobUpsertBulk) UpdateNewValues() *JobUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(job.FieldID)
			}
			if _, exists := b.mutation.Name(); exists {
				s.SetIgnore(job.FieldName)
			}
			if _, exists := b.mutation.Payload(); exists {
				s.SetIgnore(job.FieldPayload)
			}
			if _, exists := b.mutation.AvailableAt(); exists {
				s.SetIgnore(job.FieldAvailableAt)
			}
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(job.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Job.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *JobUpsertBulk) Ignore() *JobUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *JobUpsertBulk) DoNothing() *JobUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the JobCreateBulk.OnConflict
// documentation for more info.
func (u *JobUpsertBulk) Update(set func(*JobUpsert)) *JobUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&JobUpsert{UpdateSet: update})
	}))
	return u
}

// SetAttempts sets the "attempts" field.
func (u *JobUpsertBulk) SetAttempts(v int) *JobUpsertBulk {
	return u.Update(func(s *JobUpsert) {
		s.SetAttempts(v)
	})
}

// AddAttempts adds v to the "attempts" field.
func (u *JobUpsertBulk) AddAttempts(v int) *JobUpsertBulk {
	return u.Update(func(s *JobUpsert) {
		s.AddAttempts(v)
	})
}

// UpdateAttempts sets the "attempts" field to the value that was provided on create.
func (u *JobUpsertBulk) UpdateAttempts() *JobUpsertBulk {
	return u.Update(func(s *JobUpsert) {
		s.UpdateAttempts()
	})
}

// SetReservedUntil sets the "reserved_until" field.
func (u *JobUpsertBulk) SetReservedUntil(v time.Time) *JobUpsertBulk {
	return u.Update(func(s *JobUpsert) {
		s.SetReservedUntil(v)
	})
}

// UpdateReservedUntil sets the "reserved_until" field to the value that was provided on create.
func (u *JobUpsertBulk) UpdateReservedUntil() *JobUpsertBulk {
	return u.Update(func(s *JobUpsert) {
		s.UpdateReservedUntil()
	})
}

// Exec executes the query.
func (u *JobUpsertBulk) Exec(ctx context.Context) error {
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("store: OnConflict was set for builder %d. Set it on the JobCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("store: missing options for JobCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *JobUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
