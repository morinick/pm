package http

import (
	"context"
	"passman/cmd/internal/users"

	"github.com/google/uuid"
)

type sessionManager interface {
	GetString(context.Context, string) string
	Put(context.Context, string, any)
	Keys(context.Context) []string
	Destroy(context.Context) error
}

type userUsecase interface {
	Registration(context.Context, users.UserDTO) (uuid.UUID, error)
	Login(context.Context, users.UserDTO) (uuid.UUID, error)
	UpdateUser(context.Context, uuid.UUID, users.UserDTO) error
	DeleteUser(context.Context, uuid.UUID) error
	ParseUserError(error) (int, string, error)
}
