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
	return db, db.PingContext(ctx)
}
