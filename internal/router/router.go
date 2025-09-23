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

	r.Use(gzipMiddleware.WithGzipRequestDecompression)
	r.Use(gzipMiddleware.WithGzipResponseCompression)

	r.Get("/ping", pingHandler.HandlePingRepository)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", authHandler.HandleRegister)
		r.Post("/login", authHandler.HandleLogin)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.WithAuth)
			r.Post("/orders", orderHandler.HandlePostNewOrder)
			r.Get("/orders", orderHandler.HandleGetOrders)
			r.Get("/balance", balanceHandler.HandleGetUserBalance)
			r.Post("/balance/withdraw", balanceHandler.HandlePostWithdraw)
			r.Get("/withdrawals", balanceHandler.HandleGetWithdrawals)
		})
	})

	return r
}
