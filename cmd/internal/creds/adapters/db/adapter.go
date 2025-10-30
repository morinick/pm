package db

import (
	"context"
	"errors"
	"fmt"
	"passman/cmd/internal/creds"
	database "passman/pkg/database/sqlite"

	"github.com/google/uuid"
)

type Adapter struct {
	storage database.Storage
}

func New(st database.Storage) *Adapter {
	return &Adapter{storage: st}
}

func (a *Adapter) AddNewCreds(ctx context.Context, newCreds creds.ServiceDAO) error {
	sql := `insert into services_creds (id, user_id, name, secret, payload) values (?, ?, ?, ?, ?)`
	args := []any{
		newCreds.ID,
		newCreds.Owner,
		newCreds.Name,
		newCreds.Key,
		newCreds.Payload,
	}

	if _, err := a.storage.ExecContext(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *Adapter) FindCreds(ctx context.Context, userID uuid.UUID, serviceName string) (creds.ServiceDAO, error) {
	sql := `select name, secret, payload from services_creds where user_id = ? and name = ?`
	serviceCreds := creds.ServiceDAO{}

	if err := a.storage.QueryRowContext(ctx, sql, userID, serviceName).Scan(&serviceCreds.Name, &serviceCreds.Key, &serviceCreds.Payload); err != nil {
		return creds.ServiceDAO{}, err
	}

	return serviceCreds, nil
}

func (a *Adapter) GetCredsList(ctx context.Context, userID uuid.UUID) ([]string, error) {
	sql := `select name from services_creds where user_id = ?`

	rows, err := a.storage.QueryContext(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servicesNames := make([]string, 0)
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		servicesNames = append(servicesNames, name)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return servicesNames, nil
}

func (a *Adapter) UpdateCreds(ctx context.Context, oldServiceName string, updatedCreds creds.ServiceDAO) error {
	sql := `update services_creds set name = ?, secret = ?, payload = ? where user_id = ? and name = ?`
	args := []any{
		updatedCreds.Name,
		updatedCreds.Key,
		updatedCreds.Payload,
		updatedCreds.Owner,
		oldServiceName,
	}

	if tag, err := a.storage.ExecContext(ctx, sql, args...); err != nil {
		return err
	} else {
		count, err := tag.RowsAffected()
		if err != nil {
			return fmt.Errorf("RowsAffected not support")
		}
		if count == 0 {
			return database.ErrNoRows
		}
	}

	return nil
}

func (a *Adapter) RemoveCreds(ctx context.Context, userID uuid.UUID, serviceName string) error {
	sql := `delete from services_creds where user_id = ? and name = ?`

	if _, err := a.storage.ExecContext(ctx, sql, userID, serviceName); err != nil {
		return err
	}

	return nil
}

func (a *Adapter) IsEmptyRows(err error) bool {
	return errors.Is(err, database.ErrNoRows)
}
