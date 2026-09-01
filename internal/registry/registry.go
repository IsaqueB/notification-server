package registry

import (
	"context"
	"errors"

	"github.com/IsaqueB/notification-server/internal/models"
)

var (
	ErrClientNotFound = errors.New("cliente não encontrado")
	ErrClientExists   = errors.New("cliente já registrado")
)

type Registry interface {
	Register(ctx context.Context, client *models.Client) error
	Unregister(ctx context.Context, clientID string) error
	GetClient(ctx context.Context, clientID string) (*models.Client, error)
	GetAllClients(ctx context.Context) ([]*models.Client, error)
	Heartbeat(ctx context.Context, clientID string) error
	CleanupInactive(ctx context.Context, maxAge int64) error
}
