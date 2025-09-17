package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/bezjen/gophermart/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(databaseDSN string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{
		db: db,
	}, nil
}

func (p *PostgresRepository) CreateUser(ctx context.Context, login, passwordHash string) (int, error) {
	var userID int
	query := "INSERT INTO users (login, password_hash) VALUES ($1, $2) ON CONFLICT (login) DO NOTHING RETURNING id"
	err := p.db.QueryRowContext(ctx, query, login, passwordHash).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserExists
		}
		return 0, err
	}
	return userID, nil
}

func (p *PostgresRepository) GetUserByLogin(ctx context.Context, login string) (*model.DbUser, error) {
	var user model.DbUser
	var passwordHash string
	query := "SELECT id, login, password_hash FROM users WHERE login=$1"
	err := p.db.QueryRowContext(ctx, query, login).Scan(&user.ID, &user.Login, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresRepository) Close() error {
	return p.db.Close()
}
