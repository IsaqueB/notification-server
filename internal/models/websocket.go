package models

import (
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	Id   string
	Conn *websocket.Conn
	Send chan []byte

	LastSeen time.Time
}
