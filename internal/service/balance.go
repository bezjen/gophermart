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
	Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error)
}

type UserBalanceService struct {
	storage      repository.Repository
	logger       *logger.Logger
	orderService OrderService
}

func NewUserBalanceService(storage repository.Repository, logger *logger.Logger, orderService OrderService) *UserBalanceService {
	return &UserBalanceService{
		storage:      storage,
		logger:       logger,
		orderService: orderService,
	}
}

func (s *UserBalanceService) GetBalance(ctx context.Context, userID int) (*model.Balance, error) {
	return s.storage.GetBalance(ctx, userID)
}

func (s *UserBalanceService) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	if !s.orderService.ValidateOrderNumber(orderNumber) {
		return ErrOrderNumber
	}
	return s.storage.Withdraw(ctx, userID, orderNumber, sum)
}

func (s *UserBalanceService) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	return s.storage.GetWithdrawals(ctx, userID)
}
