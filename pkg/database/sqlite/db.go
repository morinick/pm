package sqlite

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(ctx context.Context, connString string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, err
	}
	return db, db.PingContext(ctx)
}
