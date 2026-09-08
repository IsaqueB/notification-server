package config

import (
	"os"
	"strconv"
	"time"

	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/IsaqueB/notification-server/pkg/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	WebSocket WebSocketConfig
	Redis     RedisConfig
	Auth      AuthConfig
	LogLevel  int
	Topics    []models.Topic
}

type ServerConfig struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type WebSocketConfig struct {
	Address        string
	MaxConnections int
	MaxMessageSize int
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

type AuthConfig struct {
	Secret     string
	Expiration time.Duration
}

func Load() *Config {
	// Carrega .env (se existir)
	godotenv.Load()

	// Se APP_ENV for definido, tenta carregar .env.{APP_ENV}
	if env := os.Getenv("APP_ENV"); env != "" {
		godotenv.Load(".env." + env)
	}

	return &Config{
		Server: ServerConfig{
			Address:      getEnv("SERVER_ADDRESS", ":8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		WebSocket: WebSocketConfig{
			Address:        getEnv("WEBSOCKET_ADDRESS", "/ws"),
			MaxConnections: getIntEnv("WEBSOCKET_MAX_CONNECTIONS", 1000),
			MaxMessageSize: getIntEnv("WEBSOCKET_MAX_MESSAGE_SIZE", 512),
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			// Secret:     getEnv("AUTH_SECRET", "super-secret-key"),
			Expiration: getDurationEnv("AUTH_EXPIRATION", 24*time.Hour),
		},
		LogLevel: getIntEnv("LOG_LEVEL", logger.DEBUG),
		Topics:   models.GetAllNotificationTopics(),
	}
}

// Helpers para converter .env
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if dur, err := time.ParseDuration(value); err == nil {
			return dur
		}
	}
	return defaultValue
}
