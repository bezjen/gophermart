package handler

import (
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/repository"
	"go.uber.org/zap"
	"net/http"
)

type PingHandler struct {
	logger  *logger.Logger
	storage repository.Repository
}

func NewPingHandler(logger *logger.Logger, storage repository.Repository) *PingHandler {
	return &PingHandler{
		logger:  logger,
		storage: storage,
	}
}

func (h *PingHandler) HandlePingRepository(rw http.ResponseWriter, r *http.Request) {
	err := h.storage.Ping(r.Context())
	if err != nil {
		h.logger.Error("Failed to ping",
			zap.Error(err),
		)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
}
