package usecases

import (
	"context"
	"passman/cmd/internal/creds"
	"passman/pkg/cipher"

	"github.com/google/uuid"
)

var (
	errCredsNotFound     = newClientError("creds not found")
	errCredsAlreadyExist = newClientError("creds already exist")
)

type credsUsecase struct {
	repo    repository
	ciphers []cipher.AESCipher
}

func New(r repository, c []cipher.AESCipher) *credsUsecase {
	return &credsUsecase{repo: r, ciphers: c}
}

func (cu *credsUsecase) AddNewCreds(ctx context.Context, userID uuid.UUID, serviceCreds creds.Service) error {
	newServiceCreds := serviceCreds.ToServiceDAO(userID, cu.ciphers)

	// Check dublicates
	if existedCreds, err := cu.repo.FindCreds(ctx, newServiceCreds.Owner, newServiceCreds.Name); err != nil && !cu.repo.IsEmptyRows(err) {
		return newInternalError("AddNewCreds", "can't find dublicates", err)
	} else if len(existedCreds.Payload) > 0 {
		return errCredsAlreadyExist
	}

	if err := cu.repo.AddNewCreds(ctx, newServiceCreds); err != nil {
		return newInternalError("AddNewCreds", "can't add creds to db", err)
	}

	return nil
}

func (cu *credsUsecase) GetCreds(ctx context.Context, userID uuid.UUID, serviceName string) (creds.Service, error) {
	serviceCredsDAO, err := cu.repo.FindCreds(ctx, userID, serviceName)
	if err != nil {
		if cu.repo.IsEmptyRows(err) {
			return creds.Service{}, errCredsNotFound
		}
		return creds.Service{}, newInternalError("GetCreds", "can't get creds from db", err)
	}

	serviceCreds, err := serviceCredsDAO.ToService(cu.ciphers)
	if err != nil {
		return creds.Service{}, newInternalError("GetCreds", "can't convert creds from dao", err)
	}

	return serviceCreds, nil
}

func (cu *credsUsecase) GetCredsList(ctx context.Context, userID uuid.UUID) ([]string, error) {
	credsList, err := cu.repo.GetCredsList(ctx, userID)
	if err != nil {
		return []string{}, newInternalError("GetCredsList", "can't get list of creds from db", err)
	}

	return credsList, nil
}

func (cu *credsUsecase) UpdateCreds(ctx context.Context, userID uuid.UUID, oldServiceName string, updatedCreds creds.Service) error {
	credsDAO := updatedCreds.ToServiceDAO(userID, cu.ciphers)

	if err := cu.repo.UpdateCreds(ctx, oldServiceName, credsDAO); err != nil {
		if cu.repo.IsEmptyRows(err) {
			return errCredsNotFound
		}
		return newInternalError("UpdateCreds", "can't update creds", err)
	}

	return nil
}

func (cu *credsUsecase) RemoveCreds(ctx context.Context, userID uuid.UUID, serviceName string) error {
	if err := cu.repo.RemoveCreds(ctx, userID, serviceName); err != nil {
		return newInternalError("RemoveCreds", "can't remove creds from db", err)
	}

	return nil
}

func (cu *credsUsecase) ParseMyError(err error) (int, string, error) {
	return parseCredsError(err)
}
