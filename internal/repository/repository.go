//go:generate mockery --name=Repository --output=../mocks --outpkg=mocks --case=underscore
package repository

import (
	"context"
	"errors"
	"github.com/bezjen/gophermart/internal/model"
)

var (
	ErrNotFound         = errors.New("record not found")
	ErrUserExists       = errors.New("user already exists")
	ErrOrderExists      = errors.New("order already exists")
	ErrOrderExistsOther = errors.New("order already uploaded by other user")
)

type Repository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (int, error)
	GetUserByLogin(ctx context.Context, login string) (*model.DbUser, error)

	CreateOrder(ctx context.Context, userID int, orderNumber string) error
	GetOrders(ctx context.Context, userID int) ([]model.Order, error)

	Ping(ctx context.Context) error
	Close() error
}
