package middleware

import (
	"context"
	"net/http"
	"project/auth"
	"strings"
)

type contextKey string

const (
	ContextUsetrID contextKey = "usetID"
	ContextRole    contextKey = "role"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)

			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)

			return
		}
		tokenString := parts[1]
		token, err := auth.ValidationToken(tokenString)
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)

			return
		}
		claims, ok := token.Claims.(auth.MyCustomClaims)
		if !ok {
			http.Error(w, "Invalid token claims!", http.StatusUnauthorized)

			return
		}
		ctx := context.WithValue(r.Context(), ContextUsetrID, claims.UserID)
		ctx = context.WithValue(ctx, ContextRole, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
