package usecases

import (
	"context"

	"passman/internal/server/accounts"

	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=mock/repository.go
type repository interface {
	AddAccount(ctx context.Context, newAccount accounts.Account) error
	GetUserAccountsInService(ctx context.Context, params accounts.QueryParams) ([]accounts.Account, error)
	GetServiceID(ctx context.Context, serviceName string) (uuid.UUID, error)
	GetAccountID(ctx context.Context, userID, serviceID uuid.UUID, credName string) (uuid.UUID, error)
	UpdateAccount(ctx context.Context, oldServiceName string, updatedAccount accounts.Account) error
	RemoveAccount(ctx context.Context, userID uuid.UUID, accountName, serviceName string) error
	RemoveAllAccountsInService(ctx context.Context, userID uuid.UUID, serviceName string) error
	IsEmptyRows(err error) bool
}
