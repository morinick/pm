package usecases

import (
	"context"

	"passman/internal/server/services"

	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=mock/repository.go
type repository interface {
	AddService(ctx context.Context, newService services.Service) error
	GetService(ctx context.Context, serviceName string) (services.Service, error)
	CheckExistingRecord(ctx context.Context, userID uuid.UUID, serviceName string) (uuid.UUID, error)
	GetAllServices(ctx context.Context) ([]services.ServiceDTO, error)
	GetAllUserServices(ctx context.Context, userID uuid.UUID) ([]services.ServiceDTO, error)
	UpdateService(ctx context.Context, oldName string, updatedService services.ServiceDTO) error
	RemoveService(ctx context.Context, serviceName string) error
	IsEmptyRows(err error) bool
}
