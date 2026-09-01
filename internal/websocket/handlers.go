package websocket

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/IsaqueB/notification-server/internal/broker"
	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/IsaqueB/notification-server/pkg/logger"
	"github.com/gorilla/websocket"
)

// TODO: Correnctly check origin
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	pool   *Pool
	broker broker.Broker
	log    *logger.Logger
}

func NewHandler(pool *Pool, broker broker.Broker, log *logger.Logger) *Handler {
	return &Handler{
		pool:   pool,
		broker: broker,
		log:    log,
	}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("client_id")
	if clientId == "" {
		http.Error(w, "client id is obrigatory", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("Erro no upgrade Websocket:", err)
		return
	}

	client := NewClient(clientId, conn)
	h.pool.Register(client)

	go h.writePump(client)
	go h.readPump(client)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func (h *Handler) writePump(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case message, ok := <-client.Send:
			// Closes pump if channel is closed
			if !ok {
				return
			}
			// Envia pela conexão WebSocket
			client.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			// Envia Ping para verificar se Windows está vivo
			client.Conn.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

func (h *Handler) readPump(client *Client) {
	for {
		_, message, err := client.Conn.ReadMessage()
		h.log.Error("Error reading message from", client.Id, err)

		var msg map[string]interface{}
		json.Unmarshal(message, &msg)

		switch msg["type"] {
		case models.HEARTBEAT:
			client.LastSeen = time.Now() // Atualiza: "Windows está vivo!"
		case models.ACK:
			h.log.Info(client.Id, "Client confirmou recebimento")
		default:
			h.log.Error("Mensagem recebida não corresponde a nenhum tipo cadastrado")
		}
	}
}
