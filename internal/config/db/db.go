package db

import (
	"errors"
	"github.com/bezjen/gophermart/internal/config"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func InitDB(cfg config.Config) (repository.Repository, error) {
	if err := runMigrations(cfg.DatabaseDSN); err != nil {
		return nil, err
	}

	repoDB, err := repository.NewPostgresRepository(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	return repoDB, nil
}

func runMigrations(databaseDSN string) error {
	m, err := migrate.New("file://migrations", databaseDSN)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
