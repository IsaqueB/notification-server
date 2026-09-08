package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IsaqueB/notification-server/internal/api"
	"github.com/IsaqueB/notification-server/internal/broker"
	"github.com/IsaqueB/notification-server/internal/config"
	"github.com/IsaqueB/notification-server/internal/registry"
	"github.com/IsaqueB/notification-server/pkg/logger"
)

func Run() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	redisBroker, err := broker.NewRedisBroker(cfg.Redis, cfg.Topics, log)
	if err != nil {
		log.Fatal("Erro ao conectar no Redis:", err)
	}
	defer redisBroker.Close()

	clientRegistry := registry.NewRedisRegistry(cfg.Redis)

	handler := api.NewHandler(redisBroker, clientRegistry, log)
	router := handler.SetupRoutes()

	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("API Server rodando em", cfg.Server.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Erro no servidor:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Desligando API Server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Erro no shutdown:", err)
	}

	log.Info("API Server desligado")
}
