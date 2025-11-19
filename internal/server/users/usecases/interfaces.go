package usecases

import (
	"context"

	"passman/internal/server/users"

	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=mock/repository.go
type dbRepo interface {
	AddUser(context.Context, users.User) error
	GetUser(context.Context, string) (users.User, error)
	GetUserByID(context.Context, uuid.UUID) (users.User, error)
	UpdateUser(context.Context, users.User) error
	RemoveUser(context.Context, uuid.UUID) error
	IsEmptyRows(error) bool
}
