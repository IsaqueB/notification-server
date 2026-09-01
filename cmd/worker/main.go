package worker

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IsaqueB/notification-server/internal/broker"
	"github.com/IsaqueB/notification-server/internal/config"
	"github.com/IsaqueB/notification-server/internal/websocket"
	"github.com/IsaqueB/notification-server/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New()

	redisBroker, err := broker.NewRedisBroker(cfg.Redis)
	if err != nil {
		log.Fatal("Erro ao conectar no Redis:", err)
	}
	defer redisBroker.Close()

	pool := websocket.NewPool()
	wsHandler := websocket.NewHandler(pool, redisBroker, log)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)
	mux.HandleFunc("/health", wsHandler.HandleHealth)

	server := &http.Server{
		Addr:         cfg.WebSocket.Address,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Worker WebSocket rodando em", cfg.WebSocket.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Erro no worker:", err)
		}
	}()

	ctx := context.Background()
	go pool.ConsumeMessages(ctx, redisBroker)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Desligando Worker...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool.Shutdown(shutdownCtx)
	server.Shutdown(shutdownCtx)

	log.Info("Worker desligado")
}
