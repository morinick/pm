package db

import (
	"context"
	"errors"
	"passman/cmd/internal/users"
	database "passman/pkg/database/sqlite"

	"github.com/google/uuid"
)

type Adapter struct {
	storage database.Storage
}

func New(st database.Storage) *Adapter {
	return &Adapter{storage: st}
}

func (a *Adapter) AddUser(ctx context.Context, userCreds users.User) error {
	sql := `insert into users (id, username, password) values (?, ?, ?)`
	args := []any{
		userCreds.ID,
		userCreds.Username,
		userCreds.Password,
	}

	if _, err := a.storage.ExecContext(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *Adapter) FindUser(ctx context.Context, username string) (users.User, error) {
	sql := `select id, username, password from users where username = ?`
	var user users.User

	if err := a.storage.QueryRowContext(ctx, sql, username).
		Scan(&user.ID, &user.Username, &user.Password); err != nil {
		return users.User{}, err
	}

	return user, nil
}

func (a *Adapter) UpdateUser(ctx context.Context, updatedUser users.User) error {
	sql := `update users set username = ?, password = ? where id = ?`
	args := []any{
		updatedUser.Username,
		updatedUser.Password,
		updatedUser.ID,
	}

	if _, err := a.storage.ExecContext(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *Adapter) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	sql := `delete from users where id = ?`

	if _, err := a.storage.ExecContext(ctx, sql, userID); err != nil {
		return err
	}

	return nil
}

func (a *Adapter) IsEmptyRows(err error) bool {
	return errors.Is(err, database.ErrNoRows)
}
