package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IsaqueB/notification-server/internal/config"
	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/IsaqueB/notification-server/pkg/logger"
	"github.com/go-redis/redis/v8"
)

type RedisBroker struct {
	log        *logger.Logger
	client     *redis.Client
	mapChannel map[models.Topic]string
}

func NewRedisBroker(cfg config.RedisConfig, topics []models.Topic, log *logger.Logger) (*RedisBroker, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	mapChannel := make(map[models.Topic]string)
	for _, topic := range topics {
		mapChannel[topic] = fmt.Sprintf("%sChannel", topic)
	}

	return &RedisBroker{
		client:     client,
		log:        log,
		mapChannel: mapChannel,
	}, nil
}

func (b *RedisBroker) PublishNotification(ctx context.Context, notification *models.Notification) error {
	b.log.Debug("Trying to publish notification to topic", notification.Title, notification.Topic)
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	b.log.Debug("Publishing notification to topic", notification.Topic, b.mapChannel[notification.Topic], notification.Topic, notification.ClientID)
	return b.client.Publish(ctx, b.mapChannel[notification.Topic], data).Err()
}

func (b *RedisBroker) PublishNotificationWithTopic(ctx context.Context, topic models.Topic, notification *models.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return b.client.Publish(ctx, string(topic), data).Err()
}

func (b *RedisBroker) SubscribeNotifications(ctx context.Context, topic models.Topic) (<-chan *models.Notification, error) {
	pubsub := b.client.Subscribe(ctx, b.mapChannel[topic])
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
	b.log.Debug("Subscribed to topic!", topic, b.mapChannel[topic])
	return notifications, nil
}

func (b *RedisBroker) Close() error {
	return b.client.Close()
}
