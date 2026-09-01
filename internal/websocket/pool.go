package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/IsaqueB/notification-server/internal/broker"
	"github.com/gorilla/websocket"
)

const SEND_CHANNEL_SIZE = 256

type Pool struct {
	clients    map[string]*Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	Id   string
	Conn *websocket.Conn
	Send chan []byte

	LastSeen time.Time
}

func NewClient(id string, conn *websocket.Conn) *Client {
	return &Client{
		Id:   id,
		Conn: conn,
		Send: make(chan []byte, SEND_CHANNEL_SIZE),
	}
}

func NewPool() *Pool {
	return &Pool{
		clients:    make(map[string]*Client),
		mu:         sync.RWMutex{},
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (p *Pool) Register(client *Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clients[client.Id] = client
}

func (p *Pool) Unregister(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists := p.clients[id]; exists {
		close(client.Send)
		delete(p.clients, id)
	}
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

	for clientID, client := range p.clients {
		select {
		case client.Send <- message:
		default:
			go p.Unregister(clientID)
		}
	}
}

func (p *Pool) ConsumeMessages(ctx context.Context, broker broker.Broker) {
	notifications, _ := broker.SubscribeNotifications(ctx)

	for notification := range notifications {
		data, _ := json.Marshal(notification)

		if notification.ClientID != "" {
			p.SendToClient(notification.ClientID, data)
		} else {
			p.Broadcast(data)
		}
	}
}

func (p *Pool) Shutdown(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	deadline := time.Now().Add(5 * time.Second)
	for id, client := range p.clients {
		err := client.Conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"),
			deadline,
		)
		// Cannot use unregister channel or method because mutex is locked while for is running.
		if err != nil {
			go p.Unregister(id)
		}
	}
}
