package http

import (
	"context"

	"passman/internal/server/accounts"
)

type accountsUsecases interface {
	AddAccount(context.Context, accounts.AccountDTO) error
	GetAccountsInService(context.Context, accounts.QueryParams) ([]accounts.AccountDTO, error)
	UpdateAccount(context.Context, string, accounts.AccountDTO) error
	RemoveAccount(context.Context, string, accounts.QueryParams) error
	RemoveAllAccountsInService(context.Context, accounts.QueryParams) error
	ParseMyError(error) (int, string, error)
}

type sessionManager interface {
	GetString(context.Context, string) string
	Keys(context.Context) []string
}
