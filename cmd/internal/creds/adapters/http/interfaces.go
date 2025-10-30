package http

import (
	"context"
	"passman/cmd/internal/creds"

	"github.com/google/uuid"
)

type credsUsecases interface {
	AddNewCreds(ctx context.Context, userID uuid.UUID, serviceCreds creds.Service) error
	GetCreds(ctx context.Context, userID uuid.UUID, serviceName string) (creds.Service, error)
	GetCredsList(ctx context.Context, userID uuid.UUID) ([]string, error)
	UpdateCreds(ctx context.Context, userID uuid.UUID, oldServiceName string, epdatedCreds creds.Service) error
	RemoveCreds(ctx context.Context, userID uuid.UUID, serviceName string) error
	ParseMyError(err error) (int, string, error)
}

type sessionManager interface {
	GetString(context.Context, string) string
	Keys(context.Context) []string
}
