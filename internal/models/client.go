package models

import (
	"sync"
	"time"
)

type Client struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Token     string                 `json:"-"`
	Endpoint  string                 `json:"endpoint"`
	LastSeen  time.Time              `json:"last_seen"`
	Connected bool                   `json:"connected"`
	OS        string                 `json:"os"`
	Version   string                 `json:"version"`
	IPAddress string                 `json:"ip_address"`
	Metadata  map[string]interface{} `json:"metadata"`
	mu        sync.RWMutex
}

func (c *Client) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastSeen = time.Now()
	c.Connected = true
}

func (c *Client) SetDisconnected() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Connected = false
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Connected
}
