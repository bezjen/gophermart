//go:generate mockery --name=OrderService --output=../mocks --case=underscore
package service

import (
	"context"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/bezjen/gophermart/internal/repository"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int) (*model.Balance, error)
}

type UserBalanceService struct {
	storage repository.Repository
	logger  *logger.Logger
}

func NewUserBalanceService(storage repository.Repository, logger *logger.Logger) *UserBalanceService {
	return &UserBalanceService{
		storage: storage,
		logger:  logger,
	}
}

func (s *UserOrderService) GetBalance(ctx context.Context, userID int) (*model.Balance, error) {
	return s.storage.GetBalance(ctx, userID)
}
