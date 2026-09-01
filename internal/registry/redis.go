package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IsaqueB/notification-server/internal/config"
	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/go-redis/redis/v8"
)

type RedisRegistry struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRegistry(cfg config.RedisConfig) *RedisRegistry {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &RedisRegistry{
		client: client,
		ttl:    2 * time.Minute,
	}
}

func (r *RedisRegistry) Register(ctx context.Context, client *models.Client) error {
	data, err := json.Marshal(client)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("client:%s", client.ID)
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *RedisRegistry) Unregister(ctx context.Context, clientID string) error {
	key := fmt.Sprintf("client:%s", clientID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisRegistry) GetClient(ctx context.Context, clientID string) (*models.Client, error) {
	key := fmt.Sprintf("client:%s", clientID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrClientNotFound
		}
		return nil, err
	}

	var client models.Client
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, err
	}

	return &client, nil
}

func (r *RedisRegistry) GetAllClients(ctx context.Context) ([]*models.Client, error) {
	keys, err := r.client.Keys(ctx, "client:*").Result()
	if err != nil {
		return nil, err
	}

	clients := make([]*models.Client, 0, len(keys))
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var client models.Client
		if err := json.Unmarshal(data, &client); err != nil {
			continue
		}

		clients = append(clients, &client)
	}

	return clients, nil
}

func (r *RedisRegistry) Heartbeat(ctx context.Context, clientID string) error {
	key := fmt.Sprintf("client:%s", clientID)
	return r.client.Expire(ctx, key, r.ttl).Err()
}

func (r *RedisRegistry) CleanupInactive(ctx context.Context, maxAge int64) error {
	return nil
}

func (r *RedisRegistry) Close() error {
	return r.client.Close()
}
