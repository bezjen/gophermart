package db

import (
	"github.com/bezjen/gophermart/internal/config"
	"github.com/bezjen/gophermart/internal/repository"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func InitDB(cfg config.Config) (repository.Repository, error) {
	repoDB, err := repository.NewPostgresRepository(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	return repoDB, nil
}
