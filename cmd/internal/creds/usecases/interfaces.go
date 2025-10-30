package usecases

import (
	"context"
	"passman/cmd/internal/creds"

	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=mock/repository.go
type repository interface {
	AddNewCreds(ctx context.Context, newCreds creds.ServiceDAO) error
	FindCreds(ctx context.Context, userID uuid.UUID, serviceName string) (creds.ServiceDAO, error)
	GetCredsList(ctx context.Context, userID uuid.UUID) ([]string, error)
	UpdateCreds(ctx context.Context, oldServiceName string, updatedCreds creds.ServiceDAO) error
	RemoveCreds(ctx context.Context, userID uuid.UUID, serviceName string) error
	IsEmptyRows(err error) bool
}
