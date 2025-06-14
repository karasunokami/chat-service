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
	"github.com/karasunokami/chat-service/internal/store/chat"
	"github.com/karasunokami/chat-service/internal/store/message"
	"github.com/karasunokami/chat-service/internal/store/problem"
	"github.com/karasunokami/chat-service/internal/types"
)

// ProblemCreate is the builder for creating a Problem entity.
type ProblemCreate struct {
	config
	mutation *ProblemMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetChatID sets the "chat_id" field.
func (pc *ProblemCreate) SetChatID(ti types.ChatID) *ProblemCreate {
	pc.mutation.SetChatID(ti)
	return pc
}

// SetManagerID sets the "manager_id" field.
func (pc *ProblemCreate) SetManagerID(ti types.UserID) *ProblemCreate {
	pc.mutation.SetManagerID(ti)
	return pc
}

// SetNillableManagerID sets the "manager_id" field if the given value is not nil.
func (pc *ProblemCreate) SetNillableManagerID(ti *types.UserID) *ProblemCreate {
	if ti != nil {
		pc.SetManagerID(*ti)
	}
	return pc
}

// SetResolvedAt sets the "resolved_at" field.
func (pc *ProblemCreate) SetResolvedAt(t time.Time) *ProblemCreate {
	pc.mutation.SetResolvedAt(t)
	return pc
}

// SetNillableResolvedAt sets the "resolved_at" field if the given value is not nil.
func (pc *ProblemCreate) SetNillableResolvedAt(t *time.Time) *ProblemCreate {
	if t != nil {
		pc.SetResolvedAt(*t)
	}
	return pc
}

// SetCreatedAt sets the "created_at" field.
func (pc *ProblemCreate) SetCreatedAt(t time.Time) *ProblemCreate {
	pc.mutation.SetCreatedAt(t)
	return pc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (pc *ProblemCreate) SetNillableCreatedAt(t *time.Time) *ProblemCreate {
	if t != nil {
		pc.SetCreatedAt(*t)
	}
	return pc
}

// SetID sets the "id" field.
func (pc *ProblemCreate) SetID(ti types.ProblemID) *ProblemCreate {
	pc.mutation.SetID(ti)
	return pc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (pc *ProblemCreate) SetNillableID(ti *types.ProblemID) *ProblemCreate {
	if ti != nil {
		pc.SetID(*ti)
	}
	return pc
}

// SetChat sets the "chat" edge to the Chat entity.
func (pc *ProblemCreate) SetChat(c *Chat) *ProblemCreate {
	return pc.SetChatID(c.ID)
}

// AddMessageIDs adds the "messages" edge to the Message entity by IDs.
func (pc *ProblemCreate) AddMessageIDs(ids ...types.MessageID) *ProblemCreate {
	pc.mutation.AddMessageIDs(ids...)
	return pc
}

// AddMessages adds the "messages" edges to the Message entity.
func (pc *ProblemCreate) AddMessages(m ...*Message) *ProblemCreate {
	ids := make([]types.MessageID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return pc.AddMessageIDs(ids...)
}

// Mutation returns the ProblemMutation object of the builder.
func (pc *ProblemCreate) Mutation() *ProblemMutation {
	return pc.mutation
}

// Save creates the Problem in the database.
func (pc *ProblemCreate) Save(ctx context.Context) (*Problem, error) {
	pc.defaults()
	return withHooks[*Problem, ProblemMutation](ctx, pc.sqlSave, pc.mutation, pc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (pc *ProblemCreate) SaveX(ctx context.Context) *Problem {
	v, err := pc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pc *ProblemCreate) Exec(ctx context.Context) error {
	_, err := pc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pc *ProblemCreate) ExecX(ctx context.Context) {
	if err := pc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (pc *ProblemCreate) defaults() {
	if _, ok := pc.mutation.CreatedAt(); !ok {
		v := problem.DefaultCreatedAt()
		pc.mutation.SetCreatedAt(v)
	}
	if _, ok := pc.mutation.ID(); !ok {
		v := problem.DefaultID()
		pc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pc *ProblemCreate) check() error {
	if _, ok := pc.mutation.ChatID(); !ok {
		return &ValidationError{Name: "chat_id", err: errors.New(`store: missing required field "Problem.chat_id"`)}
	}
	if v, ok := pc.mutation.ChatID(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "chat_id", err: fmt.Errorf(`store: validator failed for field "Problem.chat_id": %w`, err)}
		}
	}
	if v, ok := pc.mutation.ManagerID(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "manager_id", err: fmt.Errorf(`store: validator failed for field "Problem.manager_id": %w`, err)}
		}
	}
	if _, ok := pc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`store: missing required field "Problem.created_at"`)}
	}
	if v, ok := pc.mutation.ID(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`store: validator failed for field "Problem.id": %w`, err)}
		}
	}
	if _, ok := pc.mutation.ChatID(); !ok {
		return &ValidationError{Name: "chat", err: errors.New(`store: missing required edge "Problem.chat"`)}
	}
	return nil
}

func (pc *ProblemCreate) sqlSave(ctx context.Context) (*Problem, error) {
	if err := pc.check(); err != nil {
		return nil, err
	}
	_node, _spec := pc.createSpec()
	if err := sqlgraph.CreateNode(ctx, pc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*types.ProblemID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	pc.mutation.id = &_node.ID
	pc.mutation.done = true
	return _node, nil
}

func (pc *ProblemCreate) createSpec() (*Problem, *sqlgraph.CreateSpec) {
	var (
		_node = &Problem{config: pc.config}
		_spec = sqlgraph.NewCreateSpec(problem.Table, sqlgraph.NewFieldSpec(problem.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = pc.conflict
	if id, ok := pc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := pc.mutation.ManagerID(); ok {
		_spec.SetField(problem.FieldManagerID, field.TypeUUID, value)
		_node.ManagerID = value
	}
	if value, ok := pc.mutation.ResolvedAt(); ok {
		_spec.SetField(problem.FieldResolvedAt, field.TypeTime, value)
		_node.ResolvedAt = value
	}
	if value, ok := pc.mutation.CreatedAt(); ok {
		_spec.SetField(problem.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := pc.mutation.ChatIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   problem.ChatTable,
			Columns: []string{problem.ChatColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(chat.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.ChatID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := pc.mutation.MessagesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   problem.MessagesTable,
			Columns: []string{problem.MessagesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(message.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Problem.Create().
//		SetChatID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.ProblemUpsert) {
//			SetChatID(v+v).
//		}).
//		Exec(ctx)
func (pc *ProblemCreate) OnConflict(opts ...sql.ConflictOption) *ProblemUpsertOne {
	pc.conflict = opts
	return &ProblemUpsertOne{
		create: pc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Problem.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (pc *ProblemCreate) OnConflictColumns(columns ...string) *ProblemUpsertOne {
	pc.conflict = append(pc.conflict, sql.ConflictColumns(columns...))
	return &ProblemUpsertOne{
		create: pc,
	}
}

type (
	// ProblemUpsertOne is the builder for "upsert"-ing
	//  one Problem node.
	ProblemUpsertOne struct {
		create *ProblemCreate
	}

	// ProblemUpsert is the "OnConflict" setter.
	ProblemUpsert struct {
		*sql.UpdateSet
	}
)

// SetManagerID sets the "manager_id" field.
func (u *ProblemUpsert) SetManagerID(v types.UserID) *ProblemUpsert {
	u.Set(problem.FieldManagerID, v)
	return u
}

// UpdateManagerID sets the "manager_id" field to the value that was provided on create.
func (u *ProblemUpsert) UpdateManagerID() *ProblemUpsert {
	u.SetExcluded(problem.FieldManagerID)
	return u
}

// ClearManagerID clears the value of the "manager_id" field.
func (u *ProblemUpsert) ClearManagerID() *ProblemUpsert {
	u.SetNull(problem.FieldManagerID)
	return u
}

// SetResolvedAt sets the "resolved_at" field.
func (u *ProblemUpsert) SetResolvedAt(v time.Time) *ProblemUpsert {
	u.Set(problem.FieldResolvedAt, v)
	return u
}

// UpdateResolvedAt sets the "resolved_at" field to the value that was provided on create.
func (u *ProblemUpsert) UpdateResolvedAt() *ProblemUpsert {
	u.SetExcluded(problem.FieldResolvedAt)
	return u
}

// ClearResolvedAt clears the value of the "resolved_at" field.
func (u *ProblemUpsert) ClearResolvedAt() *ProblemUpsert {
	u.SetNull(problem.FieldResolvedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Problem.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(problem.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *ProblemUpsertOne) UpdateNewValues() *ProblemUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(problem.FieldID)
		}
		if _, exists := u.create.mutation.ChatID(); exists {
			s.SetIgnore(problem.FieldChatID)
		}
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(problem.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Problem.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *ProblemUpsertOne) Ignore() *ProblemUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *ProblemUpsertOne) DoNothing() *ProblemUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the ProblemCreate.OnConflict
// documentation for more info.
func (u *ProblemUpsertOne) Update(set func(*ProblemUpsert)) *ProblemUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&ProblemUpsert{UpdateSet: update})
	}))
	return u
}

// SetManagerID sets the "manager_id" field.
func (u *ProblemUpsertOne) SetManagerID(v types.UserID) *ProblemUpsertOne {
	return u.Update(func(s *ProblemUpsert) {
		s.SetManagerID(v)
	})
}

// UpdateManagerID sets the "manager_id" field to the value that was provided on create.
func (u *ProblemUpsertOne) UpdateManagerID() *ProblemUpsertOne {
	return u.Update(func(s *ProblemUpsert) {
		s.UpdateManagerID()
	})
}

// ClearManagerID clears the value of the "manager_id" field.
func (u *ProblemUpsertOne) ClearManagerID() *ProblemUpsertOne {
	return u.Update(func(s *ProblemUpsert) {
		s.ClearManagerID()
	})
}

// SetResolvedAt sets the "resolved_at" field.
func (u *ProblemUpsertOne) SetResolvedAt(v time.Time) *ProblemUpsertOne {
	return u.Update(func(s *ProblemUpsert) {
		s.SetResolvedAt(v)
	})
}

// UpdateResolvedAt sets the "resolved_at" field to the value that was provided on create.
func (u *ProblemUpsertOne) UpdateResolvedAt() *ProblemUpsertOne {
	return u.Update(func(s *ProblemUpsert) {
		s.UpdateResolvedAt()
	})
}

// ClearResolvedAt clears the value of the "resolved_at" field.
func (u *ProblemUpsertOne) ClearResolvedAt() *ProblemUpsertOne {
	return u.Update(func(s *ProblemUpsert) {
		s.ClearResolvedAt()
	})
}

// Exec executes the query.
func (u *ProblemUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("store: missing options for ProblemCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *ProblemUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *ProblemUpsertOne) ID(ctx context.Context) (id types.ProblemID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("store: ProblemUpsertOne.ID is not supported by MySQL driver. Use ProblemUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *ProblemUpsertOne) IDX(ctx context.Context) types.ProblemID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// ProblemCreateBulk is the builder for creating many Problem entities in bulk.
type ProblemCreateBulk struct {
	config
	builders []*ProblemCreate
	conflict []sql.ConflictOption
}

// Save creates the Problem entities in the database.
func (pcb *ProblemCreateBulk) Save(ctx context.Context) ([]*Problem, error) {
	specs := make([]*sqlgraph.CreateSpec, len(pcb.builders))
	nodes := make([]*Problem, len(pcb.builders))
	mutators := make([]Mutator, len(pcb.builders))
	for i := range pcb.builders {
		func(i int, root context.Context) {
			builder := pcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*ProblemMutation)
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
					_, err = mutators[i+1].Mutate(root, pcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = pcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, pcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, pcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (pcb *ProblemCreateBulk) SaveX(ctx context.Context) []*Problem {
	v, err := pcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pcb *ProblemCreateBulk) Exec(ctx context.Context) error {
	_, err := pcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pcb *ProblemCreateBulk) ExecX(ctx context.Context) {
	if err := pcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Problem.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.ProblemUpsert) {
//			SetChatID(v+v).
//		}).
//		Exec(ctx)
func (pcb *ProblemCreateBulk) OnConflict(opts ...sql.ConflictOption) *ProblemUpsertBulk {
	pcb.conflict = opts
	return &ProblemUpsertBulk{
		create: pcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Problem.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (pcb *ProblemCreateBulk) OnConflictColumns(columns ...string) *ProblemUpsertBulk {
	pcb.conflict = append(pcb.conflict, sql.ConflictColumns(columns...))
	return &ProblemUpsertBulk{
		create: pcb,
	}
}

// ProblemUpsertBulk is the builder for "upsert"-ing
// a bulk of Problem nodes.
type ProblemUpsertBulk struct {
	create *ProblemCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Problem.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(problem.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *ProblemUpsertBulk) UpdateNewValues() *ProblemUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(problem.FieldID)
			}
			if _, exists := b.mutation.ChatID(); exists {
				s.SetIgnore(problem.FieldChatID)
			}
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(problem.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Problem.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *ProblemUpsertBulk) Ignore() *ProblemUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *ProblemUpsertBulk) DoNothing() *ProblemUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the ProblemCreateBulk.OnConflict
// documentation for more info.
func (u *ProblemUpsertBulk) Update(set func(*ProblemUpsert)) *ProblemUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&ProblemUpsert{UpdateSet: update})
	}))
	return u
}

// SetManagerID sets the "manager_id" field.
func (u *ProblemUpsertBulk) SetManagerID(v types.UserID) *ProblemUpsertBulk {
	return u.Update(func(s *ProblemUpsert) {
		s.SetManagerID(v)
	})
}

// UpdateManagerID sets the "manager_id" field to the value that was provided on create.
func (u *ProblemUpsertBulk) UpdateManagerID() *ProblemUpsertBulk {
	return u.Update(func(s *ProblemUpsert) {
		s.UpdateManagerID()
	})
}

// ClearManagerID clears the value of the "manager_id" field.
func (u *ProblemUpsertBulk) ClearManagerID() *ProblemUpsertBulk {
	return u.Update(func(s *ProblemUpsert) {
		s.ClearManagerID()
	})
}

// SetResolvedAt sets the "resolved_at" field.
func (u *ProblemUpsertBulk) SetResolvedAt(v time.Time) *ProblemUpsertBulk {
	return u.Update(func(s *ProblemUpsert) {
		s.SetResolvedAt(v)
	})
}

// UpdateResolvedAt sets the "resolved_at" field to the value that was provided on create.
func (u *ProblemUpsertBulk) UpdateResolvedAt() *ProblemUpsertBulk {
	return u.Update(func(s *ProblemUpsert) {
		s.UpdateResolvedAt()
	})
}

// ClearResolvedAt clears the value of the "resolved_at" field.
func (u *ProblemUpsertBulk) ClearResolvedAt() *ProblemUpsertBulk {
	return u.Update(func(s *ProblemUpsert) {
		s.ClearResolvedAt()
	})
}

// Exec executes the query.
func (u *ProblemUpsertBulk) Exec(ctx context.Context) error {
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("store: OnConflict was set for builder %d. Set it on the ProblemCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("store: missing options for ProblemCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *ProblemUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
