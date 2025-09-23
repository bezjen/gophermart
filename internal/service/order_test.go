package service

import (
	"context"
	"errors"
	"github.com/bezjen/gophermart/internal/mocks"
	"testing"
	"time"

	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserOrderService_CreateNewOrder(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		orderNumber   string
		mockSetup     func(*mocks.Repository)
		expectedError error
	}{
		{
			name:        "successful order creation",
			userID:      1,
			orderNumber: "49927398716",
			mockSetup: func(mockRepo *mocks.Repository) {
				mockRepo.On("CreateOrder", mock.Anything, 1, "49927398716").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "invalid order number",
			userID:        1,
			orderNumber:   "1234567890",
			mockSetup:     func(mockRepo *mocks.Repository) {},
			expectedError: ErrOrderNumber,
		},
		{
			name:        "repository error",
			userID:      1,
			orderNumber: "49927398716",
			mockSetup: func(mockRepo *mocks.Repository) {
				mockRepo.On("CreateOrder", mock.Anything, 1, "49927398716").Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name:          "empty order number",
			userID:        1,
			orderNumber:   "",
			mockSetup:     func(mockRepo *mocks.Repository) {},
			expectedError: ErrOrderNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.Repository{}
			testLogger, _ := logger.NewLogger("debug")
			service := NewUserOrderService(mockRepo, testLogger)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			err := service.CreateNewOrder(context.Background(), tt.userID, tt.orderNumber)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserOrderService_ValidateOrderNumber(t *testing.T) {
	service := &UserOrderService{}

	tests := []struct {
		name     string
		number   string
		expected error
	}{
		{"valid number", "49927398716", nil},
		{"invalid number", "1234567890", ErrOrderNumber},
		{"empty string", "", ErrOrderNumber},
		{"non-numeric", "abc123", ErrOrderNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateOrderNumber(tt.number)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserOrderService_GetOrders(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		userID         int
		mockSetup      func(*mocks.Repository)
		expectedOrders []model.Order
	}{
		{
			name:   "successful get orders",
			userID: 1,
			mockSetup: func(mockRepo *mocks.Repository) {
				expectedOrders := []model.Order{
					{Number: "49927398716", Status: "NEW", Accrual: 1, UploadedAt: now},
					{Number: "1234567812345670", Status: "PROCESSED", Accrual: 2, UploadedAt: now},
				}
				mockRepo.On("GetOrders", mock.Anything, 1).Return(expectedOrders, nil)
			},
			expectedOrders: []model.Order{
				{Number: "49927398716", Status: "NEW", Accrual: 1, UploadedAt: now},
				{Number: "1234567812345670", Status: "PROCESSED", Accrual: 2, UploadedAt: now},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.Repository{}
			testLogger, _ := logger.NewLogger("debug")
			service := NewUserOrderService(mockRepo, testLogger)

			tt.mockSetup(mockRepo)

			orders, err := service.GetOrders(context.Background(), tt.userID)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOrders, orders)

			mockRepo.AssertExpectations(t)
		})
	}
}
