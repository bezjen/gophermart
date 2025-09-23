package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/mocks"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func setupTest(handler http.Handler) (*AccrualRestService, *mocks.OrderService, *httptest.Server) {
	server := httptest.NewServer(handler)
	mockOrderSvc := new(mocks.OrderService)
	testLogger, _ := logger.NewLogger("debug")
	accrualService := NewAccrualRestService(server.URL, mockOrderSvc, testLogger)
	return accrualService, mockOrderSvc, server
}

func TestAccrualRestService_processPendingOrders_Success(t *testing.T) {
	orderNumber := "12345"
	accrualResponse := model.AccrualResponse{
		Order:   orderNumber,
		Status:  model.AccrualOrderStatusProcessed,
		Accrual: 150.5,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/"+orderNumber, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accrualResponse)
	})

	accrualService, mockOrderSvc, server := setupTest(handler)
	defer server.Close()

	pendingOrders := []model.Order{{Number: orderNumber, Status: model.OrderStatusProcessing}}
	mockOrderSvc.On("GetPendingOrders", mock.Anything, mock.Anything).Return(pendingOrders, nil).Once()

	updatedOrder := model.Order{
		Number:  orderNumber,
		Status:  model.OrderStatusProcessed,
		Accrual: 150.5,
	}
	mockOrderSvc.On("UpdateOrderWithBalance", mock.Anything, updatedOrder).Return(nil).Once()

	err := accrualService.processPendingOrders(context.Background())

	assert.NoError(t, err)
	mockOrderSvc.AssertExpectations(t)
}

func TestAccrualRestService_processPendingOrders_NoContent(t *testing.T) {
	orderNumber := "54321"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/"+orderNumber, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})

	accrualService, mockOrderSvc, server := setupTest(handler)
	defer server.Close()

	pendingOrders := []model.Order{{Number: orderNumber, Status: model.OrderStatusProcessing}}
	mockOrderSvc.On("GetPendingOrders", mock.Anything, mock.Anything).Return(pendingOrders, nil).Once()

	err := accrualService.processPendingOrders(context.Background())

	assert.NoError(t, err)
	mockOrderSvc.AssertNotCalled(t, "UpdateOrderWithBalance", mock.Anything, mock.Anything)
	mockOrderSvc.AssertExpectations(t)
}

func TestAccrualRestService_processPendingOrders_InternalServerError(t *testing.T) {
	orderNumber := "11223"

	// Настраиваем HTTP сервер
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	accrualService, mockOrderSvc, server := setupTest(handler)
	defer server.Close()

	// Настраиваем мок OrderService
	pendingOrders := []model.Order{{Number: orderNumber, Status: model.OrderStatusProcessing}}
	mockOrderSvc.On("GetPendingOrders", mock.Anything, mock.Anything).Return(pendingOrders, nil).Once()

	// Выполняем тест
	err := accrualService.processPendingOrders(context.Background())

	// Проверяем результат
	assert.NoError(t, err) // processPendingOrders не возвращает ошибку, а логирует ее
	mockOrderSvc.AssertExpectations(t)
}

func TestAccrualRestService_processPendingOrders_NoPendingOrders(t *testing.T) {
	accrualService, mockOrderSvc, server := setupTest(nil) // Handler не нужен
	defer server.Close()

	// Настраиваем мок OrderService на возврат пустого среза
	mockOrderSvc.On("GetPendingOrders", mock.Anything, mock.Anything).Return([]model.Order{}, nil).Once()

	// Выполняем тест
	err := accrualService.processPendingOrders(context.Background())

	// Проверяем результат
	assert.NoError(t, err)
	mockOrderSvc.AssertExpectations(t)
}

func TestAccrualRestService_workerPool(t *testing.T) {
	orderCount := 10
	orders := make([]model.Order, 0, orderCount)
	for i := 0; i < orderCount; i++ {
		orders = append(orders, model.Order{Number: fmt.Sprintf("order-%d", i)})
	}

	var wg sync.WaitGroup
	wg.Add(orderCount)

	// Настраиваем HTTP сервер
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Каждый вызов обрабатывается успешно
		orderNumber := r.URL.Path[len("/api/orders/"):]
		response := model.AccrualResponse{
			Order:   orderNumber,
			Status:  model.AccrualOrderStatusProcessed,
			Accrual: 10,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		wg.Done()
	})

	accrualService, mockOrderSvc, server := setupTest(handler)
	defer server.Close()

	mockOrderSvc.On("GetPendingOrders", mock.Anything, mock.Anything).Return(orders, nil).Once()
	// Ожидаем вызов обновления для каждого заказа
	mockOrderSvc.On("UpdateOrderWithBalance", mock.Anything, mock.Anything).Return(nil).Times(orderCount)

	// Выполняем тест
	err := accrualService.processPendingOrders(context.Background())
	assert.NoError(t, err)

	// Ждем, пока все запросы к серверу будут обработаны
	wg.Wait()

	mockOrderSvc.AssertExpectations(t)
}
