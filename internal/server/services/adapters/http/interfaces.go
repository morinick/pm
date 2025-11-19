package http

import (
	"context"
	"io"

	"passman/internal/server/services"

	"github.com/google/uuid"
)

type serviceUsecase interface {
	AddService(context.Context, string, io.Reader) error
	GetAllServices(context.Context) ([]services.ServiceDTO, error)
	GetAllUserServices(context.Context, uuid.UUID) ([]services.ServiceDTO, error)
	UpdateService(context.Context, string, string, io.Reader) error
	RemoveService(context.Context, uuid.UUID, string) error
	ParseUserError(error) (int, string, error)
}

type sessionManager interface {
	GetString(context.Context, string) string
	Keys(context.Context) []string
}
