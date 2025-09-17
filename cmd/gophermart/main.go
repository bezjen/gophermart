package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bezjen/gophermart/internal/config"
	"github.com/bezjen/gophermart/internal/config/db"
	"github.com/bezjen/gophermart/internal/handler"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/bezjen/gophermart/internal/router"
	"github.com/bezjen/gophermart/internal/service"
	"log"
	"net/http"
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
	gophermartRouter := router.NewRouter(gophermartLogger, *pingHandler, *authHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		serverAddr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
		if err = http.ListenAndServe(serverAddr, gophermartRouter); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server failed to start: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
}
