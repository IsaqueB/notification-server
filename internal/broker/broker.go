package broker

import (
	"context"

	"github.com/IsaqueB/notification-server/internal/models"
)

type Broker interface {
	PublishNotification(ctx context.Context, notification *models.Notification) error
	// PublishNotificationWithTopic(ctx context.Context, topic models.Topic, notification *models.Notification) error
	// SubscribeNotifications(ctx context.Context) (<-chan *models.Notification, error)

	SubscribeNotifications(ctx context.Context, topic models.Topic) (<-chan *models.Notification, error)
	Close() error
}
