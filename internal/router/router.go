package router

import (
	"github.com/bezjen/gophermart/internal/handler"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(logger *logger.Logger,
	pingHandler handler.PingHandler,
	authHandler handler.AuthHandler,
) *chi.Mux {
	r := chi.NewRouter()
	gzipMiddleware := middleware.NewGzipMiddleware(logger)

	r.Use(
		gzipMiddleware.WithGzipRequestDecompression,
		gzipMiddleware.WithGzipResponseCompression)

	r.Get("/ping", pingHandler.HandlePingRepository)
	r.Post("/register", authHandler.HandleRegister)
	r.Post("/login", authHandler.HandleLogin)

	return r
}
