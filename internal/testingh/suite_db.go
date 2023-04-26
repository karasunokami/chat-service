//go:build integration

package testingh

import (
	"context"
	"strings"

	"github.com/karasunokami/chat-service/internal/store"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DBSuite struct {
	ContextSuite

	DBPrefix string
	Store    *store.Client
	Database *store.Database
	cleanUp  func(ctx context.Context)
}

func NewDBSuite(dbPrefix string) DBSuite {
	return DBSuite{DBPrefix: dbPrefix}
}

func (ds *DBSuite) SetupSuite() {
	ds.ContextSuite.SetupSuite()

	db := ds.DBPrefix + strings.ReplaceAll(uuid.New().String(), "-", "")
	ds.T().Logf("database: %s", db)

	ds.Store, ds.cleanUp = PrepareDB(ds.SuiteCtx, ds.T(), db)
	lg, _ := zap.NewProduction()
	ds.Database = store.NewDatabase(ds.Store, lg)
}

func (ds *DBSuite) TearDownSuite() {
	if f := ds.cleanUp; f != nil {
		f(ds.SuiteCtx)
	}
	ds.ContextSuite.TearDownSuite()
}
