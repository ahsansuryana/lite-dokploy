package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/lite-dokploy/backend/internal/auth"
)

type ctxKey string

const CtxUsername ctxKey = "username"

func Auth(manager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !manager.Enabled() {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := manager.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUsername, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
