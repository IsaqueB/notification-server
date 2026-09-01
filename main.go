package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Logger struct{}

func (l *Logger) Log(s string) {
	fmt.Println(s)
}

func (l *Logger) Error(e error) {
	fmt.Println("** ERROR! **", e)
}

var logger = Logger{}

type Client struct {
	ClientId string
	Conn     *websocket.Conn
	Send     chan []byte
}

type Server struct {
	mutex   *sync.Mutex
	Clients map[string]*Client

	register   chan *Client
	unregister chan *Client
	upgrader   websocket.Upgrader
}

func NewServer() *Server {
	s := &Server{
		mutex:   &sync.Mutex{},
		Clients: make(map[string]*Client, 0),

		unregister: make(chan *Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return s
}

func (s *Server) WsHandler(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("client_id")
	authToken := r.Header.Get("Authorization")

	if err := s.authenticate(clientId, authToken); err != nil {
		if err.Error() == "Usuário conectado" {
			logger.Log("Desconectando usuário " + clientId)
			s.unregister <- s.Clients[clientId]
		}
		logger.Error(err)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
		return
	}

	client := &Client{
		ClientId: clientId,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	s.register <- client

	logger.Log("✅ Cliente conectado!")

	go s.readPump(client)
	go s.writePump(client)
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.Clients[client.ClientId] = client
			s.mutex.Unlock()
		case client := <-s.unregister:
			s.mutex.Lock()
			client.Conn.Close()
			delete(s.Clients, client.ClientId)
			s.mutex.Unlock()
		}
	}
}

func (s *Server) authenticate(clientId, authToken string) error {
	if clientId == "" {
		return fmt.Errorf("ClientId inválido")
	}
	if authToken == "" {
		return fmt.Errorf("AuthToken inválido")
	}

	if authToken != fmt.Sprintf("AUTH_%s", clientId) {
		return fmt.Errorf("Unauthorized")
	}

	if _, exists := s.Clients[clientId]; exists {
		return fmt.Errorf("Usuário conectado")
	}

	return nil
}

func (s *Server) readPump(client *Client) {
	defer func() {
		s.unregister <- client
		client.Conn.Close()
	}()

	for {
		_, msg, err := client.Conn.ReadMessage()
		if err != nil {
			logger.Error(err)
			break
		}
		logger.Log(fmt.Sprintf("📩 %s enviou: %s", client.ClientId, string(msg)))
	}
}

func (s *Server) writePump(client *Client) {
	defer client.Conn.Close()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			client.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func main() {
	server := NewServer()
	// go server.Run()

	http.HandleFunc("/ws", server.WsHandler)

	logger.Log("Server rodando em :8080!")
	go http.ListenAndServe(":8080", nil)
}
