package websocket

import (
	"context"
	"net/http"
	"strings"

	"github.com/IsaqueB/notification-server/internal/auth"
)

func JWTAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("client_id")
		if id == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		payload, err := auth.JWTAuthentication(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctxWithClientId := context.WithValue(r.Context(), "client_id", payload.ClientId)
		ctxWithTopics := context.WithValue(ctxWithClientId, "topics", payload.Topics)

		rWithValues := r.WithContext(ctxWithTopics)
		next.ServeHTTP(w, rWithValues)
	})
}
