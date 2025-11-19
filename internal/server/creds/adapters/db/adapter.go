package db

import (
	"context"
	"database/sql"
	"errors"

	"passman/internal/server/creds"
	"passman/internal/server/creds/adapters/db/queries"

	"github.com/google/uuid"
)

type Adapter struct {
	storage *queries.Queries
}

func New(st queries.DBTX) *Adapter {
	return &Adapter{storage: queries.New(st)}
}

func (a *Adapter) AddCredsRecord(ctx context.Context, newCreds creds.CredsRecord) error {
	params := queries.AddCredsRecordParams{
		ID:        newCreds.ID,
		UserID:    newCreds.UserID,
		ServiceID: newCreds.ServiceID,
		Name:      newCreds.Name,
		Secret:    newCreds.Secret,
		Payload:   newCreds.Payload,
	}
	return a.storage.AddCredsRecord(ctx, params)
}

func (a *Adapter) GetUserCredsInService(ctx context.Context, queryParams creds.QueryParams) ([]creds.CredsRecord, error) {
	params := queries.GetUserCredsInServiceParams{
		UserID: queryParams.UserID,
		Name:   queryParams.ServiceName,
	}

	rows, err := a.storage.GetUserCredsInService(ctx, params)
	if err != nil {
		return nil, err
	}

	res := make([]creds.CredsRecord, 0, len(rows))
	for _, row := range rows {
		res = append(res, creds.CredsRecord{UserID: queryParams.UserID, Name: row.Name, Secret: row.Secret, Payload: row.Payload})
	}
	return res, nil
}

func (a *Adapter) GetServiceID(ctx context.Context, serviceName string) (uuid.UUID, error) {
	return a.storage.GetServiceID(ctx, serviceName)
}

func (a *Adapter) GetCredsRecordID(ctx context.Context, userID, serviceID uuid.UUID, credName string) (uuid.UUID, error) {
	return a.storage.GetCredsRecordID(ctx, queries.GetCredsRecordIDParams{UserID: userID, ServiceID: serviceID, Name: credName})
}

func (a *Adapter) UpdateCredsRecord(ctx context.Context, oldServiceName string, updatedCreds creds.CredsRecord) error {
	params := queries.UpdateCredsRecordParams{
		UserID:  updatedCreds.UserID,
		OldName: oldServiceName,
		Name:    updatedCreds.Name,
		Secret:  updatedCreds.Secret,
		Payload: updatedCreds.Payload,
	}
	return a.storage.UpdateCredsRecord(ctx, params)
}

func (a *Adapter) RemoveCredsRecord(ctx context.Context, userID uuid.UUID, credName, serviceName string) error {
	return a.storage.RemoveCredsRecord(ctx, queries.RemoveCredsRecordParams{UserID: userID, Name: credName, ServiceName: serviceName})
}

func (a *Adapter) RemoveAllCredsInService(ctx context.Context, userID uuid.UUID, serviceName string) error {
	return a.storage.RemoveAllCredsInService(ctx, queries.RemoveAllCredsInServiceParams{UserID: userID, Name: serviceName})
}

func (a *Adapter) IsEmptyRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
