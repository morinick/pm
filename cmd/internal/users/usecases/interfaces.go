package usecases

import (
	"context"
	"passman/cmd/internal/users"

	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=mock/repository.go
type dbRepo interface {
	AddUser(context.Context, users.User) error
	FindUser(context.Context, string) (users.User, error)
	UpdateUser(context.Context, users.User) error
	DeleteUser(context.Context, uuid.UUID) error
	IsEmptyRows(error) bool
}
