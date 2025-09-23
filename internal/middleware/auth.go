package middleware

import (
	"context"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/service"
	"net/http"
	"strings"
)

type userIDKey string

const (
	CookieName           = "user_token"
	UserIDKey  userIDKey = "userID"
)

type AuthMiddleware struct {
	authorizer service.Authorizer
	logger     *logger.Logger
}

func NewAuthMiddleware(authorizer service.Authorizer, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authorizer: authorizer,
		logger:     logger,
	}
}

func (m *AuthMiddleware) WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Could not find bearer token in Authorization header", http.StatusUnauthorized)
			return
		}

		userID, err := m.authorizer.ParseToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		updateCookie(w, tokenString)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func updateCookie(w http.ResponseWriter, newToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    newToken,
		Path:     "/",
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
