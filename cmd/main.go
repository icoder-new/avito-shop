package main

import (
	"context"
	"fmt"
	"github.com/icoder-new/avito-shop/api"
	"github.com/icoder-new/avito-shop/api/handler"
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/internal/service"
	"github.com/icoder-new/avito-shop/internal/storage/postgres"
	"github.com/icoder-new/avito-shop/pkg/jwt"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig("./config/config.yml")
	if err != nil {
		panic(fmt.Sprintf("error loading config: %v", err))
	}

	log, err := logger.New(cfg.Settings.Logger)
	if err != nil {
		panic(fmt.Sprintf("error initializing logger: %v", err))
	}
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := postgres.NewStorage(ctx, log, cfg.GetDSN(), cfg.Settings.DB)
	if err != nil {
		log.Fatal("failed to initialize storage", zap.Error(err))
	}
	defer storage.CloseDB()

	manager, err := jwt.NewTokenManager(cfg.Credentials.JWT)

	services := service.NewService(cfg, log, storage, manager)
	handlers := handler.NewHandler(cfg, log, services, manager)
	router := api.SetUpRoutes(handlers, log)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Settings.App.Port),
		Handler:      router,
		ReadTimeout:  cfg.Settings.Server.ReadTimeout,
		WriteTimeout: cfg.Settings.Server.WriteTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("starting server",
			zap.String("port", srv.Addr),
			zap.String("mode", cfg.Settings.App.Mode),
		)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatal("server error", zap.Error(err))
	case sig := <-shutdown:
		log.Info("shutdown signal received",
			zap.String("signal", sig.String()),
		)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Settings.Server.ShutdownTimeout)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed", zap.Error(err))
			if err := srv.Close(); err != nil {
				log.Fatal("forced shutdown failed", zap.Error(err))
			}
		}
	}
}
