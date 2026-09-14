package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"strings"
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
			h.log.Error("webhook middleware auth", "Could not find signature in header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		message := os.Getenv("BRACOMIL_BLING_WEBHOOK_INVOICE_ISSUED")
		if message == "" {
			h.log.Error("webhook middleware auth", "Could not get get bling webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		secret := os.Getenv("AUTH_SECRET")
		if secret == "" {
			h.log.Error("webhook middleware auth", "Could not get auth secret in env")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sign := base64.RawURLEncoding.EncodeToString(auth.Sign_HS256([]byte(message), []byte(secret)))
		if signature != sign {
			h.log.Error("webhook middleware auth", "signature in header was not equal to calculated")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) webhookBlingAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature := strings.TrimPrefix(r.Header.Get("X-Bling-Signature-256"), "sha256=")
		if signature == "" {
			h.log.Error("webhook middleware auth", "Could not find signature in header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		message, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			h.log.Error("webhook middleware auth", "Could not find payload to encode")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(message))

		secret := os.Getenv("BLING_CLIENT_SECRET")
		defer func() { secret = "" }()
		if secret == "" {
			h.log.Error("webhook middleware auth", "Could not get auth secret in env")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		secretB, _ := hex.DecodeString(secret)
		defer func() { secret = "" }()

		sign := hex.EncodeToString(auth.Sign_HS256([]byte(message), secretB))
		if signature != sign {
			h.log.Error("webhook middleware auth", "signature in header was not equal to calculated")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) sheetsAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature := strings.TrimPrefix(r.Header.Get("X-Sheets-Signature-256"), "sha256=")
		if signature == "" {
			h.log.Error("webhook middleware auth", "Could not find signature in header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		message, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			h.log.Error("webhook middleware auth", "Could not find payload to encode")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(message))

		secret := os.Getenv("SHEETS_CLIENT_SECRET")
		defer func() { secret = "" }()
		if secret == "" {
			h.log.Error("webhook middleware auth", "Could not get auth secret in env")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		secretB, _ := hex.DecodeString(secret)
		defer func() { secret = "" }()

		sign := hex.EncodeToString(auth.Sign_HS256([]byte(message), secretB))
		if signature != sign {
			h.log.Error("webhook middleware auth", "signature in header was not equal to calculated")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
