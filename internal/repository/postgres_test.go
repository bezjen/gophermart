package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	tests := []struct {
		name          string
		mockBehavior  func()
		login         string
		passwordHash  string
		expectedID    int
		expectedError error
	}{
		{
			name: "Success",
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO t_user").
					WithArgs("test", "hash123").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			login:         "test",
			passwordHash:  "hash123",
			expectedID:    1,
			expectedError: nil,
		},
		{
			name: "User already exists",
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO t_user").
					WithArgs("existing", "hash123").
					WillReturnError(sql.ErrNoRows)
			},
			login:         "existing",
			passwordHash:  "hash123",
			expectedID:    0,
			expectedError: ErrUserExists,
		},
		{
			name: "Database error",
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO t_user").
					WithArgs("test", "hash123").
					WillReturnError(sql.ErrConnDone)
			},
			login:         "test",
			passwordHash:  "hash123",
			expectedID:    0,
			expectedError: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			id, err := repo.CreateUser(ctx, tt.login, tt.passwordHash)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_GetUserByLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	tests := []struct {
		name          string
		mockBehavior  func()
		login         string
		expectedUser  *model.DBUser
		expectedError error
	}{
		{
			name: "Success",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT id, login, password_hash FROM t_user WHERE login=\\$1").
					WithArgs("test").
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash"}).AddRow(1, "test", "hash123"))
			},
			login: "test",
			expectedUser: &model.DBUser{
				ID:           1,
				Login:        "test",
				PasswordHash: "hash123",
			},
			expectedError: nil,
		},
		{
			name: "User not found",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT id, login, password_hash FROM t_user WHERE login=\\$1").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			login:         "nonexistent",
			expectedUser:  nil,
			expectedError: ErrNotFound,
		},
		{
			name: "Database error",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT id, login, password_hash FROM t_user WHERE login=\\$1").
					WithArgs("test").
					WillReturnError(sql.ErrConnDone)
			},
			login:         "test",
			expectedUser:  nil,
			expectedError: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			user, err := repo.GetUserByLogin(ctx, tt.login)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedUser, user)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_CreateOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	tests := []struct {
		name          string
		mockBehavior  func()
		userID        int
		orderNumber   string
		expectedError error
	}{
		{
			name: "Success",
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO t_order").
					WithArgs(1, "12345").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))
			},
			userID:        1,
			orderNumber:   "12345",
			expectedError: nil,
		},
		{
			name: "Order exists for same user",
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO t_order").
					WithArgs(1, "12345").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("SELECT user_id FROM t_order WHERE number = \\$1").
					WithArgs("12345").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))
			},
			userID:        1,
			orderNumber:   "12345",
			expectedError: ErrOrderExists,
		},
		{
			name: "Order exists for different user",
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO t_order").
					WithArgs(1, "12345").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("SELECT user_id FROM t_order WHERE number = \\$1").
					WithArgs("12345").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(2))
			},
			userID:        1,
			orderNumber:   "12345",
			expectedError: ErrOrderExistsOther,
		},
		{
			name: "Database error on select",
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO t_order").
					WithArgs(1, "12345").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("SELECT user_id FROM t_order WHERE number = \\$1").
					WithArgs("12345").
					WillReturnError(sql.ErrConnDone)
			},
			userID:        1,
			orderNumber:   "12345",
			expectedError: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			err := repo.CreateOrder(ctx, tt.userID, tt.orderNumber)

			assert.Equal(t, tt.expectedError, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_GetOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	now := time.Now()

	tests := []struct {
		name           string
		mockBehavior   func()
		userID         int
		expectedOrders []model.Order
		expectedError  error
	}{
		{
			name: "Success with orders",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
					AddRow("12345", "PROCESSED", 100.5, now).
					AddRow("67890", "NEW", nil, now.Add(-time.Hour))
				mock.ExpectQuery("SELECT number, status, accrual, uploaded_at FROM t_order WHERE user_id = \\$1 ORDER BY uploaded_at DESC").
					WithArgs(1).
					WillReturnRows(rows)
			},
			userID: 1,
			expectedOrders: []model.Order{
				{
					Number:     "12345",
					Status:     "PROCESSED",
					Accrual:    100.5,
					UploadedAt: now,
				},
				{
					Number:     "67890",
					Status:     "NEW",
					Accrual:    0,
					UploadedAt: now.Add(-time.Hour),
				},
			},
			expectedError: nil,
		},
		{
			name: "No orders",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"})
				mock.ExpectQuery("SELECT number, status, accrual, uploaded_at FROM t_order WHERE user_id = \\$1 ORDER BY uploaded_at DESC").
					WithArgs(1).
					WillReturnRows(rows)
			},
			userID:         1,
			expectedOrders: []model.Order{},
			expectedError:  nil,
		},
		{
			name: "Database error",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT number, status, accrual, uploaded_at FROM t_order WHERE user_id = \\$1 ORDER BY uploaded_at DESC").
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			userID:         1,
			expectedOrders: nil,
			expectedError:  sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			orders, err := repo.GetOrders(ctx, tt.userID)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedOrders, orders)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_GetPendingOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	now := time.Now()

	tests := []struct {
		name           string
		mockBehavior   func()
		limit          int
		expectedOrders []model.Order
		expectedError  error
	}{
		{
			name: "Success with pending orders",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"id", "number", "status", "accrual", "uploaded_at"}).
					AddRow(1, "12345", "NEW", nil, now).
					AddRow(2, "67890", "PROCESSING", 50.0, now.Add(-time.Hour))
				mock.ExpectQuery("SELECT id, number, status, accrual, uploaded_at FROM t_order WHERE status in .* ORDER BY uploaded_at limit \\$1").
					WithArgs(10).
					WillReturnRows(rows)
			},
			limit: 10,
			expectedOrders: []model.Order{
				{
					ID:         1,
					Number:     "12345",
					Status:     "NEW",
					Accrual:    0,
					UploadedAt: now,
				},
				{
					ID:         2,
					Number:     "67890",
					Status:     "PROCESSING",
					Accrual:    50.0,
					UploadedAt: now.Add(-time.Hour),
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			orders, err := repo.GetPendingOrders(ctx, tt.limit)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedOrders, orders)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_UpdateOrderWithBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	tests := []struct {
		name         string
		mockBehavior func()
		order        model.Order
		expectedErr  error
	}{
		{
			name: "Success with balance update",
			mockBehavior: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, status, user_id FROM t_order WHERE number = \\$1").
					WithArgs("12345").
					WillReturnRows(sqlmock.NewRows([]string{"id", "status", "user_id"}).AddRow(1, "NEW", 1))
				mock.ExpectExec("UPDATE t_order SET status = \\$1, accrual = \\$2 WHERE id = \\$3").
					WithArgs("PROCESSED", 100.5, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE t_user SET current_balance = current_balance \\+ \\$1 WHERE id = \\$2").
					WithArgs(100.5, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			order: model.Order{
				Number:  "12345",
				Status:  "PROCESSED",
				Accrual: 100.5,
			},
			expectedErr: nil,
		},
		{
			name: "Order not found",
			mockBehavior: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, status, user_id FROM t_order WHERE number = \\$1").
					WithArgs("12345").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			order: model.Order{
				Number:  "12345",
				Status:  "PROCESSED",
				Accrual: 100.5,
			},
			expectedErr: ErrNotFound,
		},
		{
			name: "Status unchanged - no update needed",
			mockBehavior: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, status, user_id FROM t_order WHERE number = \\$1").
					WithArgs("12345").
					WillReturnRows(sqlmock.NewRows([]string{"id", "status", "user_id"}).AddRow(1, "PROCESSED", 1))
				mock.ExpectRollback() // Транзакция откатывается, т.к. нет изменений
			},
			order: model.Order{
				Number:  "12345",
				Status:  "PROCESSED",
				Accrual: 100.5,
			},
			expectedErr: nil,
		},
		{
			name: "Status changed but no accrual for PROCESSED",
			mockBehavior: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, status, user_id FROM t_order WHERE number = \\$1").
					WithArgs("12345").
					WillReturnRows(sqlmock.NewRows([]string{"id", "status", "user_id"}).AddRow(1, "NEW", 1))
				mock.ExpectExec("UPDATE t_order SET status = \\$1, accrual = \\$2 WHERE id = \\$3").
					WithArgs("PROCESSED", 0.0, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Не ожидаем обновления баланса, т.к. accrual = 0
				mock.ExpectCommit()
			},
			order: model.Order{
				Number:  "12345",
				Status:  "PROCESSED",
				Accrual: 0.0,
			},
			expectedErr: nil,
		},
		{
			name: "Status changed to non-PROCESSED",
			mockBehavior: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, status, user_id FROM t_order WHERE number = \\$1").
					WithArgs("12345").
					WillReturnRows(sqlmock.NewRows([]string{"id", "status", "user_id"}).AddRow(1, "NEW", 1))
				mock.ExpectExec("UPDATE t_order SET status = \\$1, accrual = \\$2 WHERE id = \\$3").
					WithArgs("INVALID", 100.5, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Не ожидаем обновления баланса, т.к. статус не PROCESSED
				mock.ExpectCommit()
			},
			order: model.Order{
				Number:  "12345",
				Status:  "INVALID",
				Accrual: 100.5,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			err := repo.UpdateOrderWithBalance(ctx, tt.order)

			assert.Equal(t, tt.expectedErr, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_GetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	tests := []struct {
		name            string
		mockBehavior    func()
		userID          int
		expectedBalance *model.Balance
		expectedError   error
	}{
		{
			name: "Success",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT current_balance, withdrawn_balance FROM t_user WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"current_balance", "withdrawn_balance"}).AddRow(100.5, 50.0))
			},
			userID: 1,
			expectedBalance: &model.Balance{
				Current:   100.5,
				Withdrawn: 50.0,
			},
			expectedError: nil,
		},
		{
			name: "User not found",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT current_balance, withdrawn_balance FROM t_user WHERE id = \\$1").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			userID:          999,
			expectedBalance: nil,
			expectedError:   sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			balance, err := repo.GetBalance(ctx, tt.userID)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedBalance, balance)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_Withdraw(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	tests := []struct {
		name         string
		mockBehavior func()
		userID       int
		orderNumber  string
		sum          float64
		expectedErr  error
	}{
		{
			name: "Success",
			mockBehavior: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT current_balance, withdrawn_balance FROM t_user WHERE id = \\$1 FOR UPDATE").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"current_balance", "withdrawn_balance"}).AddRow(200.0, 50.0))
				mock.ExpectExec("UPDATE t_user SET current_balance = current_balance - \\$1, withdrawn_balance = withdrawn_balance \\+ \\$2 WHERE id = \\$3").
					WithArgs(100.0, 100.0, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO t_withdrawal").
					WithArgs(1, "12345", 100.0).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userID:      1,
			orderNumber: "12345",
			sum:         100.0,
			expectedErr: nil,
		},
		{
			name: "Not enough balance",
			mockBehavior: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT current_balance, withdrawn_balance FROM t_user WHERE id = \\$1 FOR UPDATE").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"current_balance", "withdrawn_balance"}).AddRow(50.0, 0.0))
				mock.ExpectRollback()
			},
			userID:      1,
			orderNumber: "12345",
			sum:         100.0,
			expectedErr: ErrNotEnoughBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			err := repo.Withdraw(ctx, tt.userID, tt.orderNumber, tt.sum)

			assert.Equal(t, tt.expectedErr, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_GetWithdrawals(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	now := time.Now()

	tests := []struct {
		name                string
		mockBehavior        func()
		userID              int
		expectedWithdrawals []model.Withdrawal
		expectedError       error
	}{
		{
			name: "Success with withdrawals",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
					AddRow("12345", 50.0, now).
					AddRow("67890", 30.0, now.Add(-time.Hour))
				mock.ExpectQuery("SELECT order_number, sum, processed_at FROM t_withdrawal WHERE user_id = \\$1 ORDER BY processed_at ASC").
					WithArgs(1).
					WillReturnRows(rows)
			},
			userID: 1,
			expectedWithdrawals: []model.Withdrawal{
				{
					Order:       "12345",
					Sum:         50.0,
					ProcessedAt: now,
				},
				{
					Order:       "67890",
					Sum:         30.0,
					ProcessedAt: now.Add(-time.Hour),
				},
			},
			expectedError: nil,
		},
		{
			name: "No withdrawals",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"})
				mock.ExpectQuery("SELECT order_number, sum, processed_at FROM t_withdrawal WHERE user_id = \\$1 ORDER BY processed_at ASC").
					WithArgs(1).
					WillReturnRows(rows)
			},
			userID:              1,
			expectedWithdrawals: []model.Withdrawal{},
			expectedError:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			withdrawals, err := repo.GetWithdrawals(ctx, tt.userID)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedWithdrawals, withdrawals)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
