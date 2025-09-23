//go:generate mockery --name=OrderService --output=../mocks --case=underscore
package service

import (
	"context"
	"errors"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/bezjen/gophermart/internal/repository"
	"strconv"
)

var ErrOrderNumber = errors.New("invalid order number")

type OrderService interface {
	CreateNewOrder(ctx context.Context, userID int, orderNumber string) error
	GetOrders(ctx context.Context, userID int) ([]model.Order, error)
	GetPendingOrders(ctx context.Context, limit int) ([]model.Order, error)
	UpdateOrderWithBalance(ctx context.Context, order model.Order) error
	ValidateOrderNumber(orderNumber string) error
}

type UserOrderService struct {
	storage repository.Repository
	logger  *logger.Logger
}

func NewUserOrderService(storage repository.Repository, logger *logger.Logger) *UserOrderService {
	return &UserOrderService{
		storage: storage,
		logger:  logger,
	}
}

func (s *UserOrderService) CreateNewOrder(ctx context.Context, userID int, orderNumber string) error {
	err := s.ValidateOrderNumber(orderNumber)
	if err != nil {
		return err
	}
	return s.storage.CreateOrder(ctx, userID, orderNumber)
}

func (s *UserOrderService) GetOrders(ctx context.Context, userID int) ([]model.Order, error) {
	return s.storage.GetOrders(ctx, userID)
}

func (s *UserOrderService) ValidateOrderNumber(number string) error {
	if number == "" {
		return ErrOrderNumber
	}
	sum := 0
	n := len(number)
	for i := 0; i < n; i++ {
		digit, err := strconv.Atoi(string(number[n-1-i]))
		if err != nil {
			return ErrOrderNumber
		}
		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	if sum%10 != 0 {
		return ErrOrderNumber
	}
	return nil
}

func (s *UserOrderService) GetPendingOrders(ctx context.Context, limit int) ([]model.Order, error) {
	return s.storage.GetPendingOrders(ctx, limit)
}

func (s *UserOrderService) UpdateOrderWithBalance(ctx context.Context, order model.Order) error {
	return s.storage.UpdateOrderWithBalance(ctx, order)
}
