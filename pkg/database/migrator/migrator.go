package migrator

import (
	"context"
	"errors"
	"fmt"

	sqlDatabase "passman/pkg/database/sqlite"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type migrator struct {
	m *migrate.Migrate
}

func NewMigrator(path, dbURL string) (*migrator, error) {
	db, err := sqlDatabase.NewDB(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed creating db instance: %w", err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed creating driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(path, "sqlite3", driver)
	if err != nil {
		return nil, fmt.Errorf("failed creating migrator: %w", err)
	}

	return &migrator{m: m}, nil
}

func (m *migrator) Up() error {
	if err := m.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed up migrations: %w", err)
	}
	return nil
}

func (m *migrator) Down() error {
	if err := m.m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed down migrations: %w", err)
	}
	return nil
}

func (m *migrator) MigrateToVersion(version uint) error {
	if err := m.m.Migrate(version); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed migrating to %d version: %w", version, err)
	}
	return nil
}

func (m *migrator) Close() (source error, database error) {
	return m.m.Close()
}
