package handler

import (
	"encoding/json"
	"errors"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/middleware"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/bezjen/gophermart/internal/service"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type OrderHandler struct {
	logger       *logger.Logger
	orderService service.OrderService
}

func NewOrderHandler(logger *logger.Logger, orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		logger:       logger,
		orderService: orderService,
	}
}

func (h *OrderHandler) HandlePostNewOrder(rw http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read body",
			zap.Error(err),
			zap.Int("userID", userID),
		)
		http.Error(rw, "Cannot read request body", http.StatusInternalServerError)
		return
	}
	orderNumber := string(body)

	err = h.orderService.CreateNewOrder(r.Context(), userID, orderNumber)
	if err != nil {
		if errors.Is(err, service.ErrOrderNumber) {
			http.Error(rw, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, repository.ErrOrderExists) {
			rw.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, repository.ErrOrderExistsOther) {
			http.Error(rw, "Order already uploaded by another user", http.StatusConflict)
			return
		}
		h.logger.Error("Failed to create new order",
			zap.Error(err),
			zap.Int("userID", userID),
			zap.String("orderNumber", orderNumber),
		)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) HandleGetOrders(rw http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	orders, err := h.orderService.GetOrders(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user orders",
			zap.Error(err),
			zap.Int("userID", userID),
		)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(orders)
}
