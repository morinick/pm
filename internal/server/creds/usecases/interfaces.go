package usecases

import (
	"context"

	"passman/internal/server/creds"

	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=mock/repository.go
type repository interface {
	AddCredsRecord(ctx context.Context, newCreds creds.CredsRecord) error
	GetUserCredsInService(ctx context.Context, params creds.QueryParams) ([]creds.CredsRecord, error)
	GetServiceID(ctx context.Context, serviceName string) (uuid.UUID, error)
	GetCredsRecordID(ctx context.Context, userID, serviceID uuid.UUID, credName string) (uuid.UUID, error)
	UpdateCredsRecord(ctx context.Context, oldServiceName string, updatedCreds creds.CredsRecord) error
	RemoveCredsRecord(ctx context.Context, userID uuid.UUID, credName, serviceName string) error
	RemoveAllCredsInService(ctx context.Context, userID uuid.UUID, serviceName string) error
	IsEmptyRows(err error) bool
}
