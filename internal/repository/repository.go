//go:generate mockery --name=Repository --output=../mocks --outpkg=mocks --case=underscore
package repository

import (
	"context"
	"errors"
	"github.com/bezjen/gophermart/internal/model"
)

var (
	ErrNotFound   = errors.New("record not found")
	ErrUserExists = errors.New("user already exists")
)

type Repository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (int, error)
	GetUserByLogin(ctx context.Context, login string) (*model.DbUser, error)
	Ping(ctx context.Context) error
	Close() error
}
