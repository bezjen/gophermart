package handler

import (
	"context"
	"encoding/json"
	"github.com/bezjen/gophermart/internal/mocks"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/bezjen/gophermart/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/middleware"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBalanceHandler_HandleGetUserBalance(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.BalanceService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful balance retrieval",
			setupMock: func(m *mocks.BalanceService) {
				m.On("GetBalance", mock.Anything, 1).Return(&model.Balance{
					Current:   100.5,
					Withdrawn: 50.0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"current":100.5,"withdrawn":50}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.BalanceService{}
			tt.setupMock(mockService)

			testLogger, _ := logger.NewLogger("debug")
			handler := NewBalanceHandler(testLogger, mockService)

			req := httptest.NewRequest(http.MethodGet, "/balance", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
			req = req.WithContext(ctx)

			rw := httptest.NewRecorder()

			handler.HandleGetUserBalance(rw, req)

			assert.Equal(t, tt.expectedStatus, rw.Code)
			assert.Contains(t, rw.Body.String(), tt.expectedBody)

			mockService.AssertExpectations(t)
		})
	}
}

func TestBalanceHandler_HandlePostWithdraw(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*mocks.BalanceService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful withdrawal",
			requestBody: `{"order":"1234567890","sum":50.0}`,
			setupMock: func(m *mocks.BalanceService) {
				m.On("Withdraw", mock.Anything, 1, "1234567890", 50.0).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "invalid json",
			requestBody:    `{"order":"123","sum":invalid}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request format\n",
		},
		{
			name:        "insufficient funds",
			requestBody: `{"order":"1234567890","sum":100.0}`,
			setupMock: func(m *mocks.BalanceService) {
				m.On("Withdraw", mock.Anything, 1, "1234567890", 100.0).Return(repository.ErrNotEnoughBalance)
			},
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   "Insufficient funds\n",
		},
		{
			name:        "invalid order number",
			requestBody: `{"order":"invalid","sum":50.0}`,
			setupMock: func(m *mocks.BalanceService) {
				m.On("Withdraw", mock.Anything, 1, "invalid", 50.0).Return(service.ErrOrderNumber)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Invalid order number format\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.BalanceService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			testLogger, _ := logger.NewLogger("debug")
			handler := NewBalanceHandler(testLogger, mockService)

			req := httptest.NewRequest(http.MethodPost, "/withdraw", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
			req = req.WithContext(ctx)

			rw := httptest.NewRecorder()

			handler.HandlePostWithdraw(rw, req)

			assert.Equal(t, tt.expectedStatus, rw.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rw.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBalanceHandler_HandleGetWithdrawals(t *testing.T) {
	testTime := time.Date(2025, 9, 18, 14, 42, 15, 0, time.UTC)
	tests := []struct {
		name           string
		setupMock      func(*mocks.BalanceService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful withdrawals getting",
			setupMock: func(m *mocks.BalanceService) {
				m.On("GetWithdrawals", mock.Anything, 1).Return([]model.Withdrawal{
					{
						Order:       "1234567890",
						Sum:         50.0,
						ProcessedAt: testTime,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"order":"1234567890","sum":50,"processed_at":"2025-09-18T14:42:15Z"}]`,
		},
		{
			name: "no withdrawals",
			setupMock: func(m *mocks.BalanceService) {
				m.On("GetWithdrawals", mock.Anything, 1).Return([]model.Withdrawal{}, nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.BalanceService{}
			tt.setupMock(mockService)

			testLogger, _ := logger.NewLogger("debug")
			handler := NewBalanceHandler(testLogger, mockService)

			req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
			req = req.WithContext(ctx)

			rw := httptest.NewRecorder()

			handler.HandleGetWithdrawals(rw, req)

			assert.Equal(t, tt.expectedStatus, rw.Code)
			if tt.expectedBody != "" {
				var expected []model.Withdrawal
				json.Unmarshal([]byte(tt.expectedBody), &expected)

				var actual []model.Withdrawal
				json.Unmarshal(rw.Body.Bytes(), &actual)

				assert.Equal(t, expected, actual)
			}

			mockService.AssertExpectations(t)
		})
	}
}
