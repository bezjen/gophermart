package handler

import (
	"encoding/json"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/middleware"
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
