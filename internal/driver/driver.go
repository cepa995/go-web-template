package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// DB holds the database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

// Named constants which define the nature of Connection Pool
const maxOpenDBConn = 10              // Maximum number of open DB connections at one time
const maxIdleDBConn = 5               // How many connections can be in the pool, but remain idle
const maxDBLifetime = 5 * time.Minute // Maximum lifetime for DB connection

// ConnectSQL creates database pool for PostgreSQL database
func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenDBConn)
	d.SetMaxIdleConns(maxIdleDBConn)
	d.SetConnMaxLifetime(maxDBLifetime)

	dbConn.SQL = d

	err = TestDB(d)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// TestDB tries to PING the database
func TestDB(d *sql.DB) error {
	if err := d.Ping(); err != nil {
		return err
	}
	return nil
}

// NewDatabase opens new connection to the dtabase based on dsn string
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
