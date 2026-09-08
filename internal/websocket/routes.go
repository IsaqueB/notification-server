package websocket

import (
	"github.com/IsaqueB/notification-server/internal/middleware"
	"github.com/gorilla/mux"
)

// func (h *Handler) SetupRoutes() *mux.Router {
// 	r := mux.NewRouter()

// 	r.Use(h.loggingMiddleware, h.corsMiddleware)

// 	r.HandleFunc("/api/notify", h.HandleNotify).Methods("POST")
// 	r.HandleFunc("/api/notify/all", h.HandleNotifyAll).Methods("POST")
// 	r.HandleFunc("/api/clients", h.HandleListClients).Methods("GET")
// 	r.HandleFunc("/api/clients/{id}", h.HandleGetClient).Methods("GET")
// 	r.HandleFunc("/api/health", h.HandleHealth).Methods("GET")

// 	return r
// }

func NewRouter(handler *Handler) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.Cors)

	r.HandleFunc("/health", handler.HandleHealth)
	r.HandleFunc("/login", handler.HandleLogin)

	authenticationRequired := r.NewRoute().Subrouter()
	authenticationRequired.Use(JWTAuthenticationMiddleware)
	authenticationRequired.HandleFunc("/ws", handler.HandleWebSocket)

	return r
}
