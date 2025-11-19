package usecases

import (
	"context"

	"passman/internal/server/creds"
	"passman/pkg/cipher"

	"github.com/google/uuid"
)

type credsUsecase struct {
	repo    repository
	ciphers []cipher.AESCipher
}

func New(r repository, c []cipher.AESCipher) *credsUsecase {
	return &credsUsecase{repo: r, ciphers: c}
}

func (cu *credsUsecase) AddCredsRecord(ctx context.Context, crt creds.CredsRecordTransfer) error {
	serviceID, err := cu.repo.GetServiceID(ctx, crt.ServiceName)
	if err != nil {
		if cu.repo.IsEmptyRows(err) {
			return newClientError("invalid service name")
		}
		return newInternalError("AddCredsRecord", "failed getting service id", err)
	}

	dublicateID, err := cu.repo.GetCredsRecordID(ctx, crt.UserID, serviceID, crt.Name)
	if err != nil && !cu.repo.IsEmptyRows(err) {
		return newInternalError("AddCredsRecord", "failed checking dublicates", err)
	}
	if dublicateID != uuid.Nil {
		return newClientError("creds with this name already exist")
	}

	record := crt.ToCredsRecord(serviceID, cu.ciphers)

	if err := cu.repo.AddCredsRecord(ctx, record); err != nil {
		return newInternalError("AddCredsRecord", "failed adding creds record", err)
	}

	return nil
}

func (cu *credsUsecase) GetCredsRecordsInService(ctx context.Context, params creds.QueryParams) ([]creds.CredsRecordTransfer, error) {
	records, err := cu.repo.GetUserCredsInService(ctx, params)
	if err != nil {
		return nil, newInternalError("GetCredsRecordsInService", "failed getting records", err)
	}

	crts := make([]creds.CredsRecordTransfer, 0, len(records))
	for _, r := range records {
		crt, err := r.ToCredsRecordTransfer(cu.ciphers)
		if err != nil {
			return nil, newInternalError("GetCredsRecordsInService", "failed decrypting record", err)
		}
		crts = append(crts, crt)
	}

	return crts, nil
}

func (cu *credsUsecase) UpdateCredsRecord(ctx context.Context, oldCredRecordName string, updatedCredRecord creds.CredsRecordTransfer) error {
	serviceID, err := cu.repo.GetServiceID(ctx, updatedCredRecord.ServiceName)
	if err != nil {
		if cu.repo.IsEmptyRows(err) {
			return newClientError("invalid service name")
		}
		return newInternalError("UpdateCredsRecord", "failed getting service id", err)
	}

	if _, err := cu.repo.GetCredsRecordID(ctx, updatedCredRecord.UserID, serviceID, oldCredRecordName); err != nil {
		if cu.repo.IsEmptyRows(err) {
			return newClientError("invalid old cred name")
		}
		return newInternalError("UpdateCredsRecord", "failed checking old cred record name", err)
	}

	record := updatedCredRecord.ToCredsRecord(serviceID, cu.ciphers)

	if err := cu.repo.UpdateCredsRecord(ctx, oldCredRecordName, record); err != nil {
		return newInternalError("UpdateCredsRecord", "failed updating cred record", err)
	}

	return nil
}

func (cu *credsUsecase) RemoveCredsRecord(ctx context.Context, credRecordName string, params creds.QueryParams) error {
	if err := cu.repo.RemoveCredsRecord(ctx, params.UserID, credRecordName, params.ServiceName); err != nil {
		return newInternalError("RemoveCredsRecord", "failed removing creds record", err)
	}

	return nil
}

func (cu *credsUsecase) RemoveAllCredsInService(ctx context.Context, params creds.QueryParams) error {
	if err := cu.repo.RemoveAllCredsInService(ctx, params.UserID, params.ServiceName); err != nil {
		return newInternalError("RemoveAllCredsInService", "failed removing creds records in service", err)
	}

	return nil
}

func (cu *credsUsecase) ParseMyError(err error) (int, string, error) {
	return parseCredsError(err)
}
