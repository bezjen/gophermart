package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/mocks"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/bezjen/gophermart/internal/repository"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bezjen/gophermart/internal/handler"
	"github.com/bezjen/gophermart/internal/middleware"
	"github.com/bezjen/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderHandler_HandlePostNewOrder(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful order creation",
			requestBody:    "12345678903",
			mockError:      nil,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "invalid order number format",
			requestBody:    "invalid",
			mockError:      service.ErrOrderNumber,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Invalid order number format\n",
		},
		{
			name:           "order already exists for same user",
			requestBody:    "12345678903",
			mockError:      repository.ErrOrderExists,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "order already exists for different user",
			requestBody:    "12345678903",
			mockError:      repository.ErrOrderExistsOther,
			expectedStatus: http.StatusConflict,
			expectedBody:   "Order already uploaded by another user\n",
		},
		{
			name:           "internal server error",
			requestBody:    "12345678903",
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.OrderService)
			testLogger, _ := logger.NewLogger("debug")

			handler := handler.NewOrderHandler(testLogger, mockService)

			if tt.requestBody != "" {
				mockService.On("CreateNewOrder", mock.Anything, 1, tt.requestBody).
					Return(tt.mockError)
			}

			req := httptest.NewRequest("POST", "/orders", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.HandlePostNewOrder(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestOrderHandler_HandleGetOrders(t *testing.T) {
	tests := []struct {
		name           string
		mockOrders     []model.Order
		expectedStatus int
		expectedOrders []model.Order
	}{
		{
			name: "successful get orders",
			mockOrders: []model.Order{
				{Number: "12345678903", Status: "NEW"},
				{Number: "98765432103", Status: "PROCESSED"},
			},
			expectedStatus: http.StatusOK,
			expectedOrders: []model.Order{
				{Number: "12345678903", Status: "NEW"},
				{Number: "98765432103", Status: "PROCESSED"},
			},
		},
		{
			name:           "no orders found",
			mockOrders:     []model.Order{},
			expectedStatus: http.StatusNoContent,
			expectedOrders: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.OrderService)
			testLogger, _ := logger.NewLogger("debug")

			handler := handler.NewOrderHandler(testLogger, mockService)

			mockService.On("GetOrders", mock.Anything, 1).
				Return(tt.mockOrders, nil)

			req := httptest.NewRequest("GET", "/orders", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.HandleGetOrders(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var responseOrders []model.Order
				err := json.Unmarshal(rr.Body.Bytes(), &responseOrders)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOrders, responseOrders)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			}

			mockService.AssertExpectations(t)
		})
	}
}
