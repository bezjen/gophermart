package main

import (
	"context"
	"errors"
	"github.com/bezjen/gophermart/internal/config"
	"github.com/bezjen/gophermart/internal/config/db"
	"github.com/bezjen/gophermart/internal/handler"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/bezjen/gophermart/internal/router"
	"github.com/bezjen/gophermart/internal/service"
	"log"
	"net/http"
	"time"
)

func main() {
	config.ParseConfig()
	cfg := config.AppConfig

	gophermartLogger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Printf("Error during logger initialization: %v", err)
		return
	}

	storage, err := db.InitDB(cfg)
	if err != nil {
		log.Printf("Error during storage initialization: %v", err)
		return
	}
	defer func(storage repository.Repository) {
		err = storage.Close()
		if err != nil {
			log.Printf("Error during storage close cleanly: %v", err)
		}
	}(storage)
	pingHandler := handler.NewPingHandler(gophermartLogger, storage)
	authorizer := service.NewAuthorizer([]byte(cfg.SecretKey), storage, gophermartLogger)
	authHandler := handler.NewAuthHandler(gophermartLogger, authorizer)
	orderService := service.NewUserOrderService(storage, gophermartLogger)
	orderHandler := handler.NewOrderHandler(gophermartLogger, orderService)
	balanceService := service.NewUserBalanceService(storage, gophermartLogger, orderService)
	balanceHandler := handler.NewBalanceHandler(gophermartLogger, balanceService)
	gophermartRouter := router.NewRouter(gophermartLogger, *pingHandler,
		*authHandler, *orderHandler, *balanceHandler, authorizer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accrualService := service.NewAccrualRestService(cfg.AccrualAddr, storage)
	go accrualService.StartOrderProcessingWorker(ctx, 1*time.Second) // TODO: move to config

	go func() {
		if err = http.ListenAndServe(cfg.RunAddr, gophermartRouter); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server failed to start: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
}
