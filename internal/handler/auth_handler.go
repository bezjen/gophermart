package handler

import (
	"encoding/json"
	"errors"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/bezjen/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type AuthHandler struct {
	logger     *logger.Logger
	authorizer service.Authorizer
}

func NewAuthHandler(logger *logger.Logger, authorizer service.Authorizer) *AuthHandler {
	return &AuthHandler{
		logger:     logger,
		authorizer: authorizer,
	}
}

func (h *AuthHandler) HandleRegister(rw http.ResponseWriter, r *http.Request) {
	var user model.ApiUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(rw, "Invalid request format", http.StatusBadRequest)
		return
	}

	token, err := h.authorizer.Register(r.Context(), user.Login, user.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			http.Error(rw, "Login already taken", http.StatusConflict)
			return
		}
		h.logger.Error("Failed to register user",
			zap.Error(err),
			zap.String("login", user.Login),
			zap.String("password", user.Password),
		)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Authorization", "Bearer "+token)
	rw.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) HandleLogin(rw http.ResponseWriter, r *http.Request) {
	var user model.ApiUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(rw, "Invalid request format", http.StatusBadRequest)
		return
	}

	token, err := h.authorizer.Login(r.Context(), user.Login, user.Password)
	if err != nil {
		http.Error(rw, "Invalid login/password pair", http.StatusUnauthorized)
		return
	}

	rw.Header().Set("Authorization", "Bearer "+token)
	rw.WriteHeader(http.StatusOK)
}
