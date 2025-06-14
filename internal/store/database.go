// Code generated by ent, DO NOT EDIT.

package store

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"go.uber.org/zap"
)

// Database is the client that holds all ent builders.
type Database struct {
	client *Client
	lg     *zap.Logger
}

// NewDatabase creates a new database based on Client.
func NewDatabase(client *Client, lg *zap.Logger) *Database {
	return &Database{client: client, lg: lg}
}

// RunInTx runs the given function f within a transaction.
// Inspired by https://entgo.io/docs/transactions/#best-practices.
// If there is already a transaction in the context, then the method uses it.
func (db *Database) RunInTx(ctx context.Context, f func(context.Context) error) error {
	var err error

	tx := TxFromContext(ctx)

	if tx == nil {
		tx, err = db.loadClient(ctx).Tx(ctx)
		if err != nil {
			return err
		}

		ctx = NewTxContext(ctx, tx)
	}

	defer func() {
		if v := recover(); v != nil {
			db.rollback(tx)
			panic(v)
		}
	}()

	err = f(ctx)
	if err != nil {
		db.rollback(tx)

		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) rollback(tx *Tx) {
	err := tx.Rollback()
	if err != nil {
		db.lg.Error("rollback transaction", zap.Error(err))
	}
}

func (db *Database) loadClient(ctx context.Context) *Client {
	tx := TxFromContext(ctx)
	if tx != nil {
		return tx.Client()
	}
	return db.client
}

// Exec executes a query that doesn't return rows. For example, in SQL, INSERT or UPDATE.
func (db *Database) Exec(ctx context.Context, query string, args ...interface{}) (*sql.Result, error) {
	var res sql.Result
	err := db.loadClient(ctx).driver.Exec(ctx, query, args, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Query executes a query that returns rows, typically a SELECT in SQL.
func (db *Database) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	var rows sql.Rows
	err := db.loadClient(ctx).driver.Query(ctx, query, args, &rows)
	if err != nil {
		return nil, err
	}
	return &rows, nil
}

// Chat is the client for interacting with the Chat builders.
func (db *Database) Chat(ctx context.Context) *ChatClient {
	return db.loadClient(ctx).Chat
}

// FailedJob is the client for interacting with the FailedJob builders.
func (db *Database) FailedJob(ctx context.Context) *FailedJobClient {
	return db.loadClient(ctx).FailedJob
}

// Job is the client for interacting with the Job builders.
func (db *Database) Job(ctx context.Context) *JobClient {
	return db.loadClient(ctx).Job
}

// Message is the client for interacting with the Message builders.
func (db *Database) Message(ctx context.Context) *MessageClient {
	return db.loadClient(ctx).Message
}

// Problem is the client for interacting with the Problem builders.
func (db *Database) Problem(ctx context.Context) *ProblemClient {
	return db.loadClient(ctx).Problem
}
