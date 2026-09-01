package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/IsaqueB/notification-server/internal/broker"
	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/IsaqueB/notification-server/internal/registry"
	"github.com/IsaqueB/notification-server/pkg/logger"
	"github.com/gorilla/mux"
)

type Handler struct {
	broker   broker.Broker
	registry registry.Registry
	log      *logger.Logger
}

func NewHandler(b broker.Broker, r registry.Registry, log *logger.Logger) *Handler {
	return &Handler{
		broker:   b,
		registry: r,
		log:      log,
	}
}

func (h *Handler) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/notify", h.HandleNotify).Methods("POST")
	r.HandleFunc("/api/notify/all", h.HandleNotifyAll).Methods("POST")
	r.HandleFunc("/api/clients", h.HandleListClients).Methods("GET")
	r.HandleFunc("/api/clients/{id}", h.HandleGetClient).Methods("GET")
	r.HandleFunc("/api/health", h.HandleHealth).Methods("GET")

	return r
}

func (h *Handler) HandleNotify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientID     string              `json:"client_id"`
		Notification models.Notification `json:"notification"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if req.ClientID == "" || req.Notification.Title == "" {
		h.respondError(w, http.StatusBadRequest, "Campos obrigatórios faltando")
		return
	}

	req.Notification.ClientID = req.ClientID
	req.Notification.CreatedAt = time.Now()

	if err := h.broker.PublishNotification(r.Context(), &req.Notification); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Erro ao publicar notificação")
		return
	}

	h.respondJSON(w, http.StatusOK, models.NotificationResponse{
		Success:   true,
		Message:   "Notificação enviada",
		Timestamp: time.Now(),
		ClientID:  req.ClientID,
	})
}

func (h *Handler) HandleNotifyAll(w http.ResponseWriter, r *http.Request) {
	var notification models.Notification

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		h.respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	notification.CreatedAt = time.Now()

	if err := h.broker.PublishNotification(r.Context(), &notification); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Erro ao publicar notificação")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "Notificação broadcast enviada",
		"timestamp": time.Now(),
	})
}

func (h *Handler) HandleListClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.registry.GetAllClients(r.Context())
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Erro ao listar clientes")
		return
	}

	h.respondJSON(w, http.StatusOK, clients)
}

func (h *Handler) HandleGetClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	client, err := h.registry.GetClient(r.Context(), clientID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Cliente não encontrado")
		return
	}

	h.respondJSON(w, http.StatusOK, client)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}
