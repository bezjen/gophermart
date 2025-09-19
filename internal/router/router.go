package router

import (
	"github.com/bezjen/gophermart/internal/handler"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/middleware"
	"github.com/bezjen/gophermart/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewRouter(logger *logger.Logger,
	pingHandler handler.PingHandler,
	authHandler handler.AuthHandler,
	orderHandler handler.OrderHandler,
	balanceHandler handler.BalanceHandler,
	authorizer service.Authorizer,
) *chi.Mux {
	r := chi.NewRouter()
	gzipMiddleware := middleware.NewGzipMiddleware(logger)
	authMiddleware := middleware.NewAuthMiddleware(authorizer, logger)

	r.Use(
		gzipMiddleware.WithGzipRequestDecompression,
		gzipMiddleware.WithGzipResponseCompression)

	r.Get("/ping", pingHandler.HandlePingRepository)
	r.Post("/api/user/register", authHandler.HandleRegister)
	r.Post("/api/user/login", authHandler.HandleLogin)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.WithAuth)
		r.Post("/api/user/orders", orderHandler.HandlePostNewOrder)
		r.Get("/api/user/orders", orderHandler.HandleGetOrders)
		r.Get("/api/user/balance", balanceHandler.HandleGetUserBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.HandlePostWithdraw)
		r.Get("/api/user/withdrawals", balanceHandler.HandleGetWithdrawals)
	})

	return r
}
