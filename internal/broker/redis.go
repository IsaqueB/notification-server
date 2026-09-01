package broker

import (
	"context"
	"encoding/json"

	"github.com/IsaqueB/notification-server/internal/config"
	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/go-redis/redis/v8"
)

const notificationChannel = "notifications"

type RedisBroker struct {
	client *redis.Client
}

func NewRedisBroker(cfg config.RedisConfig) (*RedisBroker, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisBroker{client: client}, nil
}

func (b *RedisBroker) PublishNotification(ctx context.Context, notification *models.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return b.client.Publish(ctx, notificationChannel, data).Err()
}

func (b *RedisBroker) SubscribeNotifications(ctx context.Context) (<-chan *models.Notification, error) {
	pubsub := b.client.Subscribe(ctx, notificationChannel)

	notifications := make(chan *models.Notification)

	go func() {
		defer pubsub.Close()
		defer close(notifications)

		for msg := range pubsub.Channel() {
			var notification models.Notification
			if err := json.Unmarshal([]byte(msg.Payload), &notification); err != nil {
				continue
			}

			notifications <- &notification
		}
	}()

	return notifications, nil
}

func (b *RedisBroker) Close() error {
	return b.client.Close()
}
