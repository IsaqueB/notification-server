package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"time"

	"github.com/IsaqueB/notification-server/internal/auth"
)

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		h.log.Info(
			"Método:", r.Method,
			"Path:", r.URL.Path,
			"Duração:", time.Since(start),
		)
	})
}

func (h *Handler) authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		clientId, err := auth.JWTAuthentication(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		rWithValue := r.WithContext(context.WithValue(r.Context(), "client_id", clientId))
		next.ServeHTTP(w, rWithValue)
	})
}

func (h *Handler) webhookBlingInvoiceIssuedAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("X-Notification-Signature-256")
		if signature == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		message := os.Getenv("BRACOMIL_BLING_WEBHOOK_INVOICE_ISSUED")
		if message == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		secret := os.Getenv("AUTH_SECRET")
		if secret == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sign := base64.RawURLEncoding.EncodeToString(auth.Sign_HS256([]byte(message), []byte(secret)))
		if signature != sign {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
