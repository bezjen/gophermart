package handler

import (
	"encoding/json"
	"errors"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/middleware"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/bezjen/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type BalanceHandler struct {
	logger         *logger.Logger
	balanceService service.BalanceService
}

func NewBalanceHandler(logger *logger.Logger, balanceService service.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		logger:         logger,
		balanceService: balanceService,
	}
}

func (h *BalanceHandler) HandleGetUserBalance(rw http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user balance",
			zap.Error(err),
			zap.Int("userID", userID),
		)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(balance)
}

func (h *BalanceHandler) HandlePostWithdraw(rw http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req model.Withdrawal
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "Invalid request format", http.StatusBadRequest)
		return
	}

	err := h.balanceService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, repository.ErrNotEnoughBalance) {
			http.Error(rw, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		if errors.Is(err, service.ErrOrderNumber) {
			http.Error(rw, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		}
		h.logger.Error("Failed to withdraw",
			zap.Error(err),
			zap.Int("userID", userID),
			zap.String("orderNumber", req.Order),
			zap.Float64("sum", req.Sum),
		)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
