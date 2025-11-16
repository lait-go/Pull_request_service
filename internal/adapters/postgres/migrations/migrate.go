package migrations

import (
	"fmt"
	"revivers/internal/adapters/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrate(path string, cfg postgres.Config) error {
	if cfg.Source == "" {
		return fmt.Errorf("migration database connection string is empty")
	}
	
	mig, err := migrate.New(fmt.Sprintf("file://%s", path), cfg.Source)
	if err != nil {
		return fmt.Errorf("Error to run migration: %w", err)
	}

	if err = mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Error to up migration: %w", err)
	}

	return nil
}
