package websocket

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/IsaqueB/notification-server/internal/auth"
	"github.com/IsaqueB/notification-server/internal/dispatcher"
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
	pool               *Pool
	dispatcher         *dispatcher.Dispatcher
	log                *logger.Logger
	pingTimer          time.Duration
	pongAwaitTime      time.Duration
	DisconnectClientCh chan *models.WebsocketClient
	Shutdown           chan bool
}

func NewHandler(pool *Pool, dispatcher *dispatcher.Dispatcher, log *logger.Logger, shutdownCh chan bool) *Handler {
	return &Handler{
		pool:               pool,
		dispatcher:         dispatcher,
		log:                log,
		pingTimer:          15 * time.Second,
		pongAwaitTime:      20 * time.Second,
		DisconnectClientCh: make(chan *models.WebsocketClient, 256),
		Shutdown:           shutdownCh,
	}
}

// Router Handlers
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	clientIdAny := r.Context().Value("client_id")
	clientId, ok := clientIdAny.(string)
	if !ok {
		http.Error(w, "could not find client id in token", http.StatusInternalServerError)
	}
	h.log.Debug("Got ClientID from JWT Token", clientId)

	topicsAny := r.Context().Value("topics")
	topics, ok := topicsAny.([]models.Topic)
	if !ok {
		http.Error(w, "could not find topics for client in token", http.StatusInternalServerError)
	}
	h.log.Debug("Topics ANY:", topicsAny)
	h.log.Debug("Got Topics from JWT Token", topics)

	// Handle client starting other connection
	if h.pool.IsClientConnected(clientId) {
		h.DisconnectClientCh <- h.pool.clients[clientId]
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("Erro no upgrade Websocket:", err)
		return
	}

	h.log.Debug("New client connected", clientId, topics)
	client := NewClient(clientId, conn)
	h.pool.Register(client)
	h.dispatcher.RegisterUsingTopics(topics, client)
	h.log.Info("Novo cliente conectado id", clientId)

	go h.writePump(client)
	go h.readPump(client)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Provisory method for getting JWT token
	body := struct {
		ClientId string `json:"client_id"`
		Token    string `json:"token"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var token string
	expiresAt := time.Now().Add(5 * time.Second)
	switch body.ClientId {
	case "weverton":
		secret := os.Getenv("WEVERTON_LOGIN_SECRET")
		if secret == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if secret != body.Token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var err error
		token, err = auth.CreateJWT_HS256(
			"weverton",
			[]models.Topic{models.INVOICE_ISSUED, models.RECEIPT_RECEIVED},
			expiresAt,
		)
		if err != nil {
			http.Error(w, "error creating jwt token", http.StatusInternalServerError)
			return
		}
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "error formatting response", http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

// Websocket Handlers
func (h *Handler) DisconnectClient(client *models.WebsocketClient) error {
	if err := h.pool.DisconnectClient(client); err != nil {
		return err
	}
	if err := h.dispatcher.UnregisterFromAllTopics(client.Id, client.Send); err != nil {
		return err
	}
	return nil
}

func (h *Handler) HandleDisconnects() {
	for client := range h.DisconnectClientCh {
		h.DisconnectClient(client)
	}
}

func (h *Handler) writePump(client *models.WebsocketClient) {
	ticker := time.NewTicker(h.pingTimer)
	defer func() {
		ticker.Stop()
		if client != nil {
			h.log.Debug("writePump finalizado para", client.Id)
			h.DisconnectClientCh <- client
		}
	}()

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
			client.Conn.SetWriteDeadline(time.Now().Add(h.pingTimer))
			client.Conn.WriteMessage(websocket.PingMessage, nil)
		case <-h.Shutdown:
			return
		}
	}
}

func (h *Handler) readPump(client *models.WebsocketClient) {
	defer func() {
		if client != nil {
			client.Conn.Close()
		}
		h.log.Debug("readPump finalizado para", client.Id)
		h.DisconnectClientCh <- client
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(h.pongAwaitTime))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(h.pongAwaitTime))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				h.log.Error("Abnormal closure for client", client.Id, "disconnecting them", err)
				break
			}
		}
		if len(message) < 1 {
			return
		}

		var msg map[string]interface{}
		json.Unmarshal(message, &msg)

		switch msg["type"] {
		case models.ACK:
			h.log.Info(client.Id, "Client confirmou recebimento")
		case models.HEARTBEAT:
			client.LastSeen = time.Now() // Atualiza: "Windows está vivo!"
		default:
			h.log.Error("Mensagem recebida não corresponde a nenhum tipo cadastrado", message)
		}
	}
}
