package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/username/progetto/shared/pkg/deduplication"
)

func NewAdminMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					slog.WarnContext(r.Context(), "token expired", "path", r.URL.Path)
					http.Error(w, "token expired", http.StatusUnauthorized)
					return
				}
				slog.ErrorContext(r.Context(), "invalid token", "path", r.URL.Path, "error", err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				slog.ErrorContext(r.Context(), "invalid token", "path", r.URL.Path)
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
}

// NewDeduplicationMiddleware creates a middleware that deduplicates requests based on X-Request-ID header.
func NewDeduplicationMiddleware(deduplicator deduplication.Deduplicator, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check for X-Request-ID
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				// No ID = No Deduplication
				next.ServeHTTP(w, r)
				return
			}

			// 2. Check Uniqueness
			// Use a prefix for gateway to avoid collisions with other services if sharing same redis
			// But RedisDeduplicator might already have a prefix.
			// Let's assume the deduplicator instance is configured with the correct prefix.
			unique, err := deduplicator.IsUnique(r.Context(), reqID, ttl)
			if err != nil {
				slog.ErrorContext(r.Context(), "deduplication check failed", "error", err)
				// Fail open or closed? If Redis is down, we probably shouldn't block all traffic unless critical.
				// But strict exactly-once requires fail closed.
				// Let's return 500 for now.
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if !unique {
				slog.WarnContext(r.Context(), "duplicate request detected", "request_id", reqID)
				http.Error(w, "duplicate request", http.StatusConflict)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewLoggingMiddleware creates a middleware that logs HTTP requests start and end.
func NewLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logger.InfoContext(r.Context(), "Request started",
				"method", r.Method,
				"path", r.URL.Path,
			)

			// Simple wrapper to capture status code
			ww := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			logger.InfoContext(r.Context(), "Request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.statusCode,
				"duration", duration,
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
func (lrw *loggingResponseWriter) Flush() {
	if flusher, ok := lrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
