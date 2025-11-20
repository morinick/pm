package db

import (
	"context"
	"database/sql"
	"errors"

	"passman/internal/server/accounts"
	"passman/internal/server/accounts/adapters/db/queries"

	"github.com/google/uuid"
)

type Adapter struct {
	storage *queries.Queries
}

func New(st queries.DBTX) *Adapter {
	return &Adapter{storage: queries.New(st)}
}

func (a *Adapter) AddAccount(ctx context.Context, newAccount accounts.Account) error {
	params := queries.AddAccountParams{
		ID:        newAccount.ID,
		UserID:    newAccount.UserID,
		ServiceID: newAccount.ServiceID,
		Name:      newAccount.Name,
		Secret:    newAccount.Secret,
		Payload:   newAccount.Payload,
	}
	return a.storage.AddAccount(ctx, params)
}

func (a *Adapter) GetUserAccountsInService(ctx context.Context, queryParams accounts.QueryParams) ([]accounts.Account, error) {
	params := queries.GetUserAccountsInServiceParams{
		UserID: queryParams.UserID,
		Name:   queryParams.ServiceName,
	}

	rows, err := a.storage.GetUserAccountsInService(ctx, params)
	if err != nil {
		return nil, err
	}

	res := make([]accounts.Account, 0, len(rows))
	for _, row := range rows {
		res = append(res, accounts.Account{UserID: queryParams.UserID, Name: row.Name, Secret: row.Secret, Payload: row.Payload})
	}
	return res, nil
}

func (a *Adapter) GetServiceID(ctx context.Context, serviceName string) (uuid.UUID, error) {
	return a.storage.GetServiceID(ctx, serviceName)
}

func (a *Adapter) GetAccountID(ctx context.Context, userID, serviceID uuid.UUID, credName string) (uuid.UUID, error) {
	return a.storage.GetAccountID(ctx, queries.GetAccountIDParams{UserID: userID, ServiceID: serviceID, Name: credName})
}

func (a *Adapter) UpdateAccount(ctx context.Context, oldServiceName string, updatedAccount accounts.Account) error {
	params := queries.UpdateAccountParams{
		UserID:  updatedAccount.UserID,
		OldName: oldServiceName,
		Name:    updatedAccount.Name,
		Secret:  updatedAccount.Secret,
		Payload: updatedAccount.Payload,
	}
	return a.storage.UpdateAccount(ctx, params)
}

func (a *Adapter) RemoveAccount(ctx context.Context, userID uuid.UUID, accName, serviceName string) error {
	return a.storage.RemoveAccount(ctx, queries.RemoveAccountParams{UserID: userID, Name: accName, ServiceName: serviceName})
}

func (a *Adapter) RemoveAllAccountsInService(ctx context.Context, userID uuid.UUID, serviceName string) error {
	return a.storage.RemoveAllAccountsInService(ctx, queries.RemoveAllAccountsInServiceParams{UserID: userID, Name: serviceName})
}

func (a *Adapter) IsEmptyRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
