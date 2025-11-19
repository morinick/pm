package usecases

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"passman/internal/server/services"

	"github.com/google/uuid"
)

type serviceUsecase struct {
	repo    repository
	saveDir string
}

func New(repo repository, saveDir string) *serviceUsecase {
	return &serviceUsecase{repo: repo, saveDir: saveDir}
}

func (su *serviceUsecase) AddService(ctx context.Context, serviceName string, logoFile io.Reader) error {
	saveDir := filepath.Join(su.saveDir, serviceName)

	// Check dublicates
	if srv, err := su.repo.GetService(ctx, serviceName); err != nil && !su.repo.IsEmptyRows(err) {
		return newInternalError("AddService", "failed checking services dublicates", err)
	} else if len(srv.Logo) > 0 {
		return newClientError("service with this name already exist")
	}

	if err := su.repo.AddService(ctx, services.Service{ID: uuid.New(), Name: serviceName, Logo: saveDir}); err != nil {
		return newInternalError("AddService", "failed inserting new service", err)
	}

	saveFile, err := os.Create(saveDir)
	if err != nil {
		return newInternalError("AddService", "failed creating logo file", err)
	}
	defer saveFile.Close()

	if _, err := io.Copy(saveFile, logoFile); err != nil {
		return newInternalError("AddService", "failed saving logo", err)
	}

	return nil
}

func (su *serviceUsecase) GetAllServices(ctx context.Context) ([]services.ServiceDTO, error) {
	srvs, err := su.repo.GetAllServices(ctx)
	if err != nil && !su.repo.IsEmptyRows(err) {
		return nil, newInternalError("GetAllServices", "failed getting all services from database", err)
	}
	return srvs, nil
}

func (su *serviceUsecase) GetAllUserServices(ctx context.Context, userID uuid.UUID) ([]services.ServiceDTO, error) {
	srvs, err := su.repo.GetAllUserServices(ctx, userID)
	if err != nil && !su.repo.IsEmptyRows(err) {
		return nil, newInternalError("GetAllUserServices", "failed getting all user services from database", err)
	}
	return srvs, nil
}

func (su *serviceUsecase) UpdateService(ctx context.Context, oldName, newName string, logoFile io.Reader) error {
	oldFilename := filepath.Join(su.saveDir, oldName)
	newFilename := filepath.Join(su.saveDir, newName)

	if oldName != newName {
		// Check dublicates
		if srv, err := su.repo.GetService(ctx, newName); err != nil && !su.repo.IsEmptyRows(err) {
			return newInternalError("UpdateService", "failed checking services dublicates", err)
		} else if len(srv.Logo) > 0 {
			return newClientError("service with this name already exist")
		}

		if err := su.repo.UpdateService(ctx, oldName, services.ServiceDTO{Name: newName, Logo: newFilename}); err != nil {
			return newInternalError("UpdateService", "failed updating service in database", err)
		}

		if err := os.Rename(oldFilename, newFilename); err != nil {
			return newInternalError("UpdateService", "failed renaming logo file", err)
		}
	}

	saveFile, err := os.OpenFile(newFilename, os.O_RDWR, 0o664)
	if err != nil {
		return newInternalError("UpdateService", "failed opening logo file", err)
	}
	defer saveFile.Close()

	if _, err := io.Copy(saveFile, logoFile); err != nil {
		return newInternalError("UpdateService", "failed saving logo file", err)
	}

	return nil
}

func (su *serviceUsecase) RemoveService(ctx context.Context, userID uuid.UUID, serviceName string) error {
	credID, err := su.repo.CheckExistingRecord(ctx, userID, serviceName)
	if err != nil && !su.repo.IsEmptyRows(err) {
		return newInternalError("RemoveService", "failed checking another records", err)
	}
	if credID != uuid.Nil {
		return newClientError("another users has records in this service")
	}

	if err := su.repo.RemoveService(ctx, serviceName); err != nil {
		return newInternalError("RemoveService", "failed removing service from database", err)
	}

	if err := os.Remove(filepath.Join(su.saveDir, serviceName)); err != nil {
		return newInternalError("RemoveService", "failed removing logo file", err)
	}

	return nil
}

func (su *serviceUsecase) ParseUserError(err error) (int, string, error) {
	return parseServiceError(err)
}
