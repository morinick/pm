package sqlite

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNoRows = sql.ErrNoRows

type Storage interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func NewDB(ctx context.Context, connString string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, initTables(ctx, db)
}

func initTables(ctx context.Context, db *sql.DB) error {
	sqlCreateUsers := `create table users (
  id uuid primary key,
  username text unique not null,
  password text not null
)`
	sqlCreateCreds := `create table services_creds (
  id uuid primary key,
  user_id uuid not null,
  name text not null,
  secret integer not null,
  payload text not null,
  foreign key (user_id) references users(id) on delete cascade
)`

	if _, err := db.ExecContext(ctx, sqlCreateUsers); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, sqlCreateCreds); err != nil {
		return err
	}
	return nil
}
