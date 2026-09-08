package worker

import (
	"context"
	"net/http"
	"time"

	"github.com/IsaqueB/notification-server/internal/broker"
	"github.com/IsaqueB/notification-server/internal/config"
	"github.com/IsaqueB/notification-server/internal/dispatcher"
	"github.com/IsaqueB/notification-server/internal/websocket"
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

	dispatcher := dispatcher.NewDispatcher(cfg.Topics, log)
	go dispatcher.ConsumeMessages(context.Background(), redisBroker)

	pool := websocket.NewPool(log)
	wsHandlerShutdownCh := make(chan bool)
	wsHandler := websocket.NewHandler(pool, dispatcher, log, wsHandlerShutdownCh)

	go wsHandler.HandleDisconnects()

	router := websocket.NewRouter(wsHandler)

	server := &http.Server{
		Addr:         cfg.WebSocket.Address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("Worker WebSocket rodando em", cfg.WebSocket.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Erro no worker:", err)
	}
	// go func() {
	// }()

	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit

	// log.Info("Desligando Worker...")
	// shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	// wsHandlerShutdownCh <- true
	// pool.Shutdown(shutdownCtx)
	// dispatcher.Shutdown(shutdownCtx)
	// server.Shutdown(shutdownCtx)

	// log.Info("Worker desligado")
}
