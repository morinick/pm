package db

import (
	"context"
	"database/sql"
	"errors"

	"passman/internal/server/users"
	"passman/internal/server/users/adapters/db/queries"

	"github.com/google/uuid"
)

type Adapter struct {
	storage *queries.Queries
}

func New(st queries.DBTX) *Adapter {
	return &Adapter{storage: queries.New(st)}
}

func (a *Adapter) AddUser(ctx context.Context, userCreds users.User) error {
	return a.storage.AddUser(
		ctx,
		queries.AddUserParams{
			ID:       userCreds.ID,
			Username: userCreds.Username,
			Password: userCreds.Password,
		},
	)
}

func (a *Adapter) GetUser(ctx context.Context, username string) (users.User, error) {
	row, err := a.storage.GetUser(ctx, username)
	if err != nil {
		return users.User{}, err
	}
	return users.User{ID: row.ID, Username: username, Password: row.Password}, nil
}

func (a *Adapter) GetUserByID(ctx context.Context, userID uuid.UUID) (users.User, error) {
	row, err := a.storage.GetUserByID(ctx, userID)
	if err != nil {
		return users.User{}, err
	}
	return users.User{ID: userID, Username: row.Username, Password: row.Password}, nil
}

func (a *Adapter) UpdateUser(ctx context.Context, updatedUser users.User) error {
	return a.storage.UpdateUser(
		ctx,
		queries.UpdateUserParams{
			ID:       updatedUser.ID,
			Username: updatedUser.Username,
			Password: updatedUser.Password,
		},
	)
}

func (a *Adapter) RemoveUser(ctx context.Context, userID uuid.UUID) error {
	return a.storage.RemoveUser(ctx, userID)
}

func (a *Adapter) IsEmptyRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
