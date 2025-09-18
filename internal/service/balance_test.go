package service

import (
	"context"
	"github.com/bezjen/gophermart/internal/mocks"
	"github.com/bezjen/gophermart/internal/repository"
	"testing"
	"time"

	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserBalanceService_GetBalance(t *testing.T) {
	tests := []struct {
		name            string
		userID          int
		mockSetup       func(*mocks.Repository)
		expectedBalance *model.Balance
	}{
		{
			name:   "successful get balance",
			userID: 1,
			mockSetup: func(mockRepo *mocks.Repository) {
				expectedBalance := &model.Balance{Current: 100.5, Withdrawn: 50.0}
				mockRepo.On("GetBalance", mock.Anything, 1).Return(expectedBalance, nil)
			},
			expectedBalance: &model.Balance{Current: 100.5, Withdrawn: 50.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.Repository{}
			mockOrderService := &mocks.OrderService{}
			testLogger, _ := logger.NewLogger("debug")
			service := NewUserBalanceService(mockRepo, testLogger, mockOrderService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			balance, err := service.GetBalance(context.Background(), tt.userID)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBalance, balance)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserBalanceService_Withdraw(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		orderNumber   string
		sum           float64
		mockSetup     func(*mocks.Repository, *mocks.OrderService)
		expectedError error
	}{
		{
			name:        "successful withdrawal",
			userID:      1,
			orderNumber: "49927398716",
			sum:         50.0,
			mockSetup: func(mockRepo *mocks.Repository, mockOrderService *mocks.OrderService) {
				mockOrderService.On("ValidateOrderNumber", "49927398716").Return(true)
				mockRepo.On("Withdraw", mock.Anything, 1, "49927398716", 50.0).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "invalid order number",
			userID:      1,
			orderNumber: "1234567890",
			sum:         50.0,
			mockSetup: func(mockRepo *mocks.Repository, mockOrderService *mocks.OrderService) {
				mockOrderService.On("ValidateOrderNumber", "1234567890").Return(false)
			},
			expectedError: ErrOrderNumber,
		},
		{
			name:        "repository error",
			userID:      1,
			orderNumber: "49927398716",
			sum:         50.0,
			mockSetup: func(mockRepo *mocks.Repository, mockOrderService *mocks.OrderService) {
				mockOrderService.On("ValidateOrderNumber", "49927398716").Return(true)
				mockRepo.On("Withdraw", mock.Anything, 1, "49927398716", 50.0).Return(repository.ErrNotEnoughBalance)
			},
			expectedError: repository.ErrNotEnoughBalance,
		},
		{
			name:        "zero sum withdrawal",
			userID:      1,
			orderNumber: "49927398716",
			sum:         0.0,
			mockSetup: func(mockRepo *mocks.Repository, mockOrderService *mocks.OrderService) {
				mockOrderService.On("ValidateOrderNumber", "49927398716").Return(true)
				mockRepo.On("Withdraw", mock.Anything, 1, "49927398716", 0.0).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "negative sum withdrawal",
			userID:      1,
			orderNumber: "49927398716",
			sum:         -10.0,
			mockSetup: func(mockRepo *mocks.Repository, mockOrderService *mocks.OrderService) {
				mockOrderService.On("ValidateOrderNumber", "49927398716").Return(true)
				mockRepo.On("Withdraw", mock.Anything, 1, "49927398716", -10.0).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.Repository{}
			mockOrderService := &mocks.OrderService{}
			testLogger, _ := logger.NewLogger("debug")
			service := NewUserBalanceService(mockRepo, testLogger, mockOrderService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockOrderService)
			}

			err := service.Withdraw(context.Background(), tt.userID, tt.orderNumber, tt.sum)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockOrderService.AssertExpectations(t)
		})
	}
}

func TestUserBalanceService_GetWithdrawals(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name                string
		userID              int
		mockSetup           func(*mocks.Repository)
		expectedWithdrawals []model.Withdrawal
	}{
		{
			name:   "successful get withdrawals",
			userID: 1,
			mockSetup: func(mockRepo *mocks.Repository) {
				expectedWithdrawals := []model.Withdrawal{
					{Order: "49927398716", Sum: 50.0, ProcessedAt: now},
					{Order: "1234567812345670", Sum: 25.0, ProcessedAt: now},
				}
				mockRepo.On("GetWithdrawals", mock.Anything, 1).Return(expectedWithdrawals, nil)
			},
			expectedWithdrawals: []model.Withdrawal{
				{Order: "49927398716", Sum: 50.0, ProcessedAt: now},
				{Order: "1234567812345670", Sum: 25.0, ProcessedAt: now},
			},
		},
		{
			name:   "no withdrawals found",
			userID: 1,
			mockSetup: func(mockRepo *mocks.Repository) {
				mockRepo.On("GetWithdrawals", mock.Anything, 1).Return([]model.Withdrawal{}, nil)
			},
			expectedWithdrawals: []model.Withdrawal{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.Repository{}
			mockOrderService := &mocks.OrderService{}
			testLogger, _ := logger.NewLogger("debug")
			service := NewUserBalanceService(mockRepo, testLogger, mockOrderService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			withdrawals, err := service.GetWithdrawals(context.Background(), tt.userID)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedWithdrawals, withdrawals)

			mockRepo.AssertExpectations(t)
		})
	}
}
