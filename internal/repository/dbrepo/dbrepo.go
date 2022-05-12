package dbrepo

import (
	"database/sql"

	"github.com/cepa995/go-web-template/internal/config"
	"github.com/cepa995/go-web-template/internal/repository"
)

// postgresDBRepo holds information about application config and DB connection.
type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// testDBRepo is struct used for unit testing and it holds information about application
// config and DB connection
type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresRepo instantiates new postgresDBRepo object based on specified DB connection
// and application config
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

// NewTestingRepo initializes new testDBRepo object based on application config only. We do
// not need a DB itself for the purpose of unit testing.
func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}
