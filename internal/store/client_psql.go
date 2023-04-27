package store

import (
	"database/sql"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:generate options-gen -out-filename=client_psql_options.gen.go -from-struct=PSQLOptions
type PSQLOptions struct {
	address  string `option:"mandatory" validate:"required,hostname_port"`
	username string `option:"mandatory" validate:"required"`
	password string `option:"mandatory" validate:"required"`
	database string `option:"mandatory" validate:"required"`
	debug    bool
}

func NewPSQLClient(opts PSQLOptions) (*Client, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %v", err)
	}

	db, err := NewPgxDB(NewPgxOptions(opts.address, opts.username, opts.password, opts.database))
	if err != nil {
		return nil, fmt.Errorf("init db driver: %v", err)
	}

	drv := entsql.OpenDB("postgres", db)

	clientOpts := []Option{
		Driver(drv),
	}

	if opts.debug {
		clientOpts = append(clientOpts, Debug())
	}

	return NewClient(clientOpts...), nil
}

//go:generate options-gen -out-filename=client_psql_pgx_options.gen.go -from-struct=PgxOptions
type PgxOptions struct {
	address  string `option:"mandatory" validate:"required,hostname_port"`
	username string `option:"mandatory" validate:"required"`
	password string `option:"mandatory" validate:"required"`
	database string `option:"mandatory" validate:"required"`
}

func (o *PgxOptions) connString() string {
	return fmt.Sprintf("postgres://%v:%v@%v/%v", o.username, o.password, o.address, o.database)
}

func NewPgxDB(opts PgxOptions) (*sql.DB, error) {
	err := opts.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	cfg, err := pgx.ParseConfig(opts.connString())
	if err != nil {
		return nil, fmt.Errorf("parse conn string to pgx config, err=%v", err)
	}

	pgxDB := stdlib.OpenDB(*cfg)

	return pgxDB, nil
}
