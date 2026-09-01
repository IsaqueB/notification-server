package broker

import (
	"context"

	"github.com/IsaqueB/notification-server/internal/models"
)

type Broker interface {
	PublishNotification(ctx context.Context, notification *models.Notification) error
	SubscribeNotifications(ctx context.Context) (<-chan *models.Notification, error)
	Close() error
}
