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
	query := "INSERT INTO t_user (login, password_hash) VALUES ($1, $2) ON CONFLICT (login) DO NOTHING RETURNING id"
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
	query := "SELECT id, login, password_hash FROM t_user WHERE login=$1"
	err := p.db.QueryRowContext(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (p *PostgresRepository) CreateOrder(ctx context.Context, userID int, orderNumber string) error {
	query := `
        INSERT INTO t_order (user_id, number, status, uploaded_at) 
        VALUES ($1, $2, 'NEW', NOW())
        ON CONFLICT (number) DO NOTHING
        RETURNING user_id`

	var insertedUserID int
	err := p.db.QueryRowContext(ctx, query, userID, orderNumber).Scan(&insertedUserID)

	if errors.Is(err, sql.ErrNoRows) {
		var existingUserID int
		err = p.db.QueryRowContext(ctx,
			"SELECT user_id FROM t_order WHERE number = $1",
			orderNumber).Scan(&existingUserID)

		if err != nil {
			return err
		}
		if existingUserID == userID {
			return ErrOrderExists
		}
		return ErrOrderExistsOther
	}

	return err
}

func (p *PostgresRepository) GetOrders(ctx context.Context, userID int) ([]model.Order, error) {
	rows, err := p.db.QueryContext(ctx,
		"SELECT number, status, accrual, uploaded_at FROM t_order WHERE user_id = $1 ORDER BY uploaded_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		var accrual sql.NullFloat64
		if err = rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		if accrual.Valid {
			order.Accrual = accrual.Float64
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresRepository) Close() error {
	return p.db.Close()
}
