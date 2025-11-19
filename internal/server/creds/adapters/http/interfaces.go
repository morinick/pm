package http

import (
	"context"

	"passman/internal/server/creds"
)

type credsUsecases interface {
	AddCredsRecord(ctx context.Context, crt creds.CredsRecordTransfer) error
	GetCredsRecordsInService(ctx context.Context, params creds.QueryParams) ([]creds.CredsRecordTransfer, error)
	UpdateCredsRecord(ctx context.Context, oldCredRecordName string, updatedCredRecord creds.CredsRecordTransfer) error
	RemoveCredsRecord(ctx context.Context, credRecordName string, params creds.QueryParams) error
	RemoveAllCredsInService(ctx context.Context, params creds.QueryParams) error
	ParseMyError(err error) (int, string, error)
}

type sessionManager interface {
	GetString(context.Context, string) string
	Keys(context.Context) []string
}
