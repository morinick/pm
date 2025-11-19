package db

import (
	"context"
	"database/sql"
	"errors"

	"passman/internal/server/services"
	"passman/internal/server/services/adapters/db/queries"

	"github.com/google/uuid"
)

type Adapter struct {
	storage *queries.Queries
}

func New(st queries.DBTX) *Adapter {
	return &Adapter{storage: queries.New(st)}
}

func (a *Adapter) AddService(ctx context.Context, newService services.Service) error {
	return a.storage.AddService(ctx, queries.AddServiceParams{ID: newService.ID, Name: newService.Name, Logo: newService.Logo})
}

func (a *Adapter) GetService(ctx context.Context, serviceName string) (services.Service, error) {
	row, err := a.storage.GetService(ctx, serviceName)
	return services.Service{ID: row.ID, Name: serviceName, Logo: row.Logo}, err
}

func (a *Adapter) CheckExistingRecord(ctx context.Context, userID uuid.UUID, serviceName string) (uuid.UUID, error) {
	return a.storage.CheckExistingRecord(ctx, queries.CheckExistingRecordParams{UserID: userID, Name: serviceName})
}

func (a *Adapter) GetAllServices(ctx context.Context) ([]services.ServiceDTO, error) {
	rows, err := a.storage.GetServicesList(ctx)
	if err != nil {
		return nil, err
	}

	srvs := make([]services.ServiceDTO, 0, len(rows))
	for _, row := range rows {
		srvs = append(srvs, services.ServiceDTO{Name: row.Name, Logo: row.Logo})
	}

	return srvs, nil
}

func (a *Adapter) GetAllUserServices(ctx context.Context, userID uuid.UUID) ([]services.ServiceDTO, error) {
	rows, err := a.storage.GetUserServicesList(ctx, userID)
	if err != nil {
		return nil, err
	}

	srvs := make([]services.ServiceDTO, 0, len(rows))
	for _, row := range rows {
		srvs = append(srvs, services.ServiceDTO{Name: row.Name, Logo: row.Logo})
	}

	return srvs, nil
}

func (a *Adapter) UpdateService(ctx context.Context, oldName string, updatedService services.ServiceDTO) error {
	return a.storage.UpdateService(ctx, queries.UpdateServiceParams{Name: updatedService.Name, Logo: updatedService.Logo, OldName: oldName})
}

func (a *Adapter) RemoveService(ctx context.Context, serviceName string) error {
	return a.storage.RemoveService(ctx, serviceName)
}

func (a *Adapter) IsEmptyRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
