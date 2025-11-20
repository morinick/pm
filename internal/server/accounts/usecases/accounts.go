package usecases

import (
	"context"

	"passman/internal/server/accounts"
	"passman/pkg/cipher"

	"github.com/google/uuid"
)

type AccountsUsecase struct {
	repo    repository
	ciphers []cipher.AESCipher
}

func New(r repository, c []cipher.AESCipher) *AccountsUsecase {
	return &AccountsUsecase{repo: r, ciphers: c}
}

func (cu *AccountsUsecase) AddAccount(ctx context.Context, dto accounts.AccountDTO) error {
	serviceID, err := cu.repo.GetServiceID(ctx, dto.ServiceName)
	if err != nil {
		if cu.repo.IsEmptyRows(err) {
			return newClientError("invalid service name")
		}
		return newInternalError("AddAccount", "failed getting service id", err)
	}

	dublicateID, err := cu.repo.GetAccountID(ctx, dto.UserID, serviceID, dto.Name)
	if err != nil && !cu.repo.IsEmptyRows(err) {
		return newInternalError("AddAccount", "failed checking dublicates", err)
	}
	if dublicateID != uuid.Nil {
		return newClientError("account with this name already exist")
	}

	account := dto.ToAccount(serviceID, cu.ciphers)

	if err := cu.repo.AddAccount(ctx, account); err != nil {
		return newInternalError("AddAccount", "failed adding account", err)
	}

	return nil
}

func (cu *AccountsUsecase) GetAccountsInService(ctx context.Context, params accounts.QueryParams) ([]accounts.AccountDTO, error) {
	records, err := cu.repo.GetUserAccountsInService(ctx, params)
	if err != nil {
		return nil, newInternalError("GetAccountsInService", "failed getting accounts", err)
	}

	dtos := make([]accounts.AccountDTO, 0, len(records))
	for _, r := range records {
		dto, err := r.ToAccountDTO(cu.ciphers)
		if err != nil {
			return nil, newInternalError("GetAccountsInService", "failed decrypting account", err)
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func (cu *AccountsUsecase) UpdateAccount(ctx context.Context, oldAccountName string, updatedAccountDTO accounts.AccountDTO) error {
	serviceID, err := cu.repo.GetServiceID(ctx, updatedAccountDTO.ServiceName)
	if err != nil {
		if cu.repo.IsEmptyRows(err) {
			return newClientError("invalid service name")
		}
		return newInternalError("UpdateAccount", "failed getting service id", err)
	}

	if _, err := cu.repo.GetAccountID(ctx, updatedAccountDTO.UserID, serviceID, oldAccountName); err != nil {
		if cu.repo.IsEmptyRows(err) {
			return newClientError("invalid old account name")
		}
		return newInternalError("UpdateAccount", "failed checking old account name", err)
	}

	record := updatedAccountDTO.ToAccount(serviceID, cu.ciphers)

	if err := cu.repo.UpdateAccount(ctx, oldAccountName, record); err != nil {
		return newInternalError("UpdateAccount", "failed updating account", err)
	}

	return nil
}

func (cu *AccountsUsecase) RemoveAccount(ctx context.Context, accountName string, params accounts.QueryParams) error {
	if err := cu.repo.RemoveAccount(ctx, params.UserID, accountName, params.ServiceName); err != nil {
		return newInternalError("RemoveAccount", "failed removing account", err)
	}

	return nil
}

func (cu *AccountsUsecase) RemoveAllAccountsInService(ctx context.Context, params accounts.QueryParams) error {
	if err := cu.repo.RemoveAllAccountsInService(ctx, params.UserID, params.ServiceName); err != nil {
		return newInternalError("RemoveAllAccountsInService", "failed removing creds records in service", err)
	}

	return nil
}

func (cu *AccountsUsecase) ParseMyError(err error) (int, string, error) {
	return parseAccountsError(err)
}
