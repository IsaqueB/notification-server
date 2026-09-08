package websocket

import (
	"context"
	"sync"

	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/IsaqueB/notification-server/pkg/logger"
	"github.com/gorilla/websocket"
)

const SEND_CHANNEL_SIZE = 256

type Pool struct {
	log        *logger.Logger
	clients    map[string]*models.WebsocketClient
	mu         sync.RWMutex
	register   chan *models.WebsocketClient
	unregister chan *models.WebsocketClient
}

func NewClient(id string, conn *websocket.Conn) *models.WebsocketClient {
	return &models.WebsocketClient{
		Id:   id,
		Conn: conn,
		Send: make(chan []byte, SEND_CHANNEL_SIZE),
	}
}

func NewPool(log *logger.Logger) *Pool {
	return &Pool{
		log:        log,
		clients:    make(map[string]*models.WebsocketClient),
		mu:         sync.RWMutex{},
		register:   make(chan *models.WebsocketClient),
		unregister: make(chan *models.WebsocketClient),
	}
}

func (p *Pool) Register(client *models.WebsocketClient) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clients[client.Id] = client
}

func (p *Pool) DisconnectClient(client *models.WebsocketClient) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	current, exists := p.clients[client.Id]
	if !exists || current != client {
		return nil
	}
	// Close connection
	closeMsg := websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "New connection established, closing old one")
	client.Conn.WriteMessage(websocket.CloseMessage, closeMsg)
	client.Conn.Close()
	// Remove from pool
	delete(p.clients, client.Id)
	close(client.Send)
	return nil
}

func (p *Pool) SendToClient(id string, message []byte) bool {
	p.mu.RLock()
	client, exists := p.clients[id]
	p.mu.RUnlock()

	if !exists {
		return false
	}

	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

func (p *Pool) Broadcast(message []byte) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, client := range p.clients {
		select {
		case client.Send <- message:
		default:

		}
	}
}

func (p *Pool) Shutdown(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// deadline := time.Now().Add(5 * time.Second)
	for _, client := range p.clients {
		p.DisconnectClient(client)
	}
}

func (p *Pool) IsClientConnected(clientId string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.clients[clientId]
	return exists
}
