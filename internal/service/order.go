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
	ValidateOrderNumber(orderNumber string) bool
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
	if !s.ValidateOrderNumber(orderNumber) {
		return ErrOrderNumber
	}
	return s.storage.CreateOrder(ctx, userID, orderNumber)
}

func (s *UserOrderService) GetOrders(ctx context.Context, userID int) ([]model.Order, error) {
	return s.storage.GetOrders(ctx, userID)
}

func (s *UserOrderService) ValidateOrderNumber(number string) bool {
	sum := 0
	n := len(number)
	for i := 0; i < n; i++ {
		digit, err := strconv.Atoi(string(number[n-1-i]))
		if err != nil {
			return false
		}
		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
