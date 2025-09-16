package router

import (
	"github.com/bezjen/gophermart/internal/handler"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(logger *logger.Logger,
	pingHandler handler.PingHandler,
) *chi.Mux {
	r := chi.NewRouter()
	gzipMiddleware := middleware.NewGzipMiddleware(logger)

	r.Use(
		gzipMiddleware.WithGzipRequestDecompression,
		gzipMiddleware.WithGzipResponseCompression)

	r.Get("/ping", pingHandler.HandlePingRepository)

	return r
}
