package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IsaqueB/notification-server/internal/broker"
	"github.com/IsaqueB/notification-server/internal/middleware"
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

	r.Use(h.loggingMiddleware, middleware.Cors)

	r.HandleFunc("/notify", h.HandleNotify).Methods("POST")
	r.HandleFunc("/notify/all", h.HandleNotifyAll).Methods("POST")
	r.HandleFunc("/clients", h.HandleListClients).Methods("GET")
	r.HandleFunc("/clients/{id}", h.HandleGetClient).Methods("GET")
	r.HandleFunc("/health", h.HandleHealth).Methods("GET")

	webhookBlingRouter := r.NewRoute().Subrouter()
	webhookBlingRouter.Use(h.webhookBlingInvoiceIssuedAuthorization)
	webhookBlingRouter.HandleFunc("/webhook/bling/invoice-issued", h.HandleBlingInvoiceIssued).Methods("POST")

	return r
}

func (h *Handler) HandleNotify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientID     string              `json:"client_id"`
		Notification models.Notification `json:"notification"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("Failed to decode notify request", err)
		h.respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if req.Notification.Title == "" {
		h.respondError(w, http.StatusBadRequest, "Campos obrigatórios faltando")
		return
	}

	req.Notification.ClientID = req.ClientID
	req.Notification.CreatedAt = time.Now()

	if req.Notification.Topic == models.EMPTY {
		req.Notification.Topic = models.ALL
	}

	if req.Notification.Topic == models.PRIVATE && req.Notification.ClientID == "" {
		h.respondError(w, http.StatusBadRequest, "Necessário client_id para notificações privadas")
		return
	}

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

func (h *Handler) HandleBlingInvoiceIssued(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid body")
	}
	event := models.BlingWebhookEvent[models.BlingWebhookPayloadInvoice]{}
	if err := json.Unmarshal(body, &event); err != nil {
		h.respondError(w, http.StatusBadRequest, "Erro ao fazer unmarshal do evento")
		return
	}
	// Parse from event to notification
	notification := &models.Notification{
		Topic:    models.INVOICE_ISSUED,
		Title:    "Nova Nota Fiscal Emitida!",
		Message:  fmt.Sprintf("A NF %s foi emitida pela SEFAZ", event.Data.Number),
		Sound:    true,
		Priority: "critical",
	}

	if err := h.broker.PublishNotification(r.Context(), notification); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Erro ao publicar notificação")
		return
	}

	h.respondJSON(w, http.StatusOK, nil)
}
