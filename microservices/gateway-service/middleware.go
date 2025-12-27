package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only protect /admin* routes
		if !strings.HasPrefix(r.URL.Path, "/admin") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			slog.ErrorContext(r.Context(), "missing authorization header", "path", r.URL.Path)
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			slog.ErrorContext(r.Context(), "invalid authorization header format", "path", r.URL.Path)
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		jwtSecret := os.Getenv("APP_JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "supersecretkey" // consistent with auth service default
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			slog.ErrorContext(r.Context(), "invalid token", "path", r.URL.Path, "error", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Check for "role" claim
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if role, ok := claims["role"].(string); ok {
				if role == "admin" {
					// Authorized
					ctx := context.WithValue(r.Context(), "user_token", token)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		slog.ErrorContext(r.Context(), "forbidden: admin role required", "path", r.URL.Path)
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
	})
}
