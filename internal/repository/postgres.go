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

func (p *PostgresRepository) GetPendingOrders(ctx context.Context) ([]model.Order, error) {
	rows, err := p.db.QueryContext(ctx,
		"SELECT id, number, status, accrual, uploaded_at FROM t_order WHERE status in ('NEW', 'PROCESSING', 'REGISTERED') ORDER BY uploaded_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		var accrual sql.NullFloat64
		if err = rows.Scan(&order.ID, &order.Number, &order.Status, &accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		if accrual.Valid {
			order.Accrual = accrual.Float64
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (p *PostgresRepository) UpdateOrderWithBalance(ctx context.Context, order model.Order) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var dbOrder model.Order
	getOrderQuery := "SELECT id, status, user_id FROM t_order WHERE number = $1"
	err = p.db.QueryRowContext(ctx, getOrderQuery, order.Number).Scan(&dbOrder.ID, &dbOrder.Status, &dbOrder.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if order.Status == dbOrder.Status {
		return nil
	}

	updateOrderQuery := "UPDATE t_order SET status = $1, accrual = $2 WHERE id = $3"
	_, err = tx.ExecContext(ctx, updateOrderQuery, order.Status, order.Accrual, dbOrder.ID)
	if err != nil {
		return err
	}

	if order.Status == "PROCESSED" && order.Accrual > 0 {
		updateBalanceQuery := "UPDATE t_user SET current_balance = current_balance + $1 WHERE id = $2"
		_, err = tx.ExecContext(ctx, updateBalanceQuery, order.Accrual, dbOrder.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *PostgresRepository) GetBalance(ctx context.Context, userID int) (*model.Balance, error) {
	var balance model.Balance
	query := `SELECT current_balance, withdrawn_balance FROM t_user WHERE id = $1`
	err := p.db.QueryRowContext(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (p *PostgresRepository) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance model.Balance
	queryBalance := "SELECT current_balance, withdrawn_balance FROM t_user WHERE id = $1 FOR UPDATE"
	err = tx.QueryRowContext(ctx, queryBalance, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return err
	}

	if balance.Current < sum {
		return ErrNotEnoughBalance
	}

	queryUpdateBalance := `
		UPDATE t_user 
		SET current_balance = current_balance - $1, withdrawn_balance = withdrawn_balance + $2 
		WHERE id = $3`
	_, err = tx.ExecContext(ctx, queryUpdateBalance, sum, sum, userID)
	if err != nil {
		return err
	}

	queryInsert := "INSERT INTO t_withdrawal (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, NOW())"
	_, err = tx.ExecContext(ctx, queryInsert, userID, orderNumber, sum)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (p *PostgresRepository) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	queryWithdrawals := "SELECT order_number, sum, processed_at FROM t_withdrawal WHERE user_id = $1 ORDER BY processed_at ASC"
	rows, err := p.db.QueryContext(ctx, queryWithdrawals, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		if err = rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, rows.Err()
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresRepository) Close() error {
	return p.db.Close()
}
