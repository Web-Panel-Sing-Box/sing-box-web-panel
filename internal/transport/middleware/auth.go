package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"sing-box-web-panel/internal/lib/auth"
)

type contextKey string

const AdminIDKey contextKey = "admin_id"

var publicPaths = map[string]bool{
	"/api/auth/login":          true,
	"/api/auth/login/totp":     true,
	"/api/auth/login/recovery": true,
	"/api/auth/logout":         true,
	"/api":                     true,
	"/api/health":              true,
}

func Auth(jwt *auth.JWTManager, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if publicPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			if isPublicPrefix(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			token := extractToken(r)
			if token == "" {
				log.Warn("auth failed: no token",
					slog.String("path", r.URL.Path),
					slog.String("remote", r.RemoteAddr),
				)
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			adminID, err := jwt.Validate(token)
			if err != nil {
				log.Warn("auth failed: invalid token",
					slog.String("path", r.URL.Path),
					slog.String("remote", r.RemoteAddr),
					slog.String("error", err.Error()),
				)
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AdminIDKey, adminID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// publicPrefixes are path prefixes served without authentication: the Swagger
// UI and the public subscription endpoints (token is the credential).
var publicPrefixes = []string{
	"/swagger",
	"/sub/",
	"/api/subscription/",
}

func isPublicPrefix(path string) bool {
	for _, p := range publicPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	// Frontend SPA paths: anything not under /api or /swagger is a frontend
	// asset or route handled by the embedded SPA and needs no auth.
	if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/swagger") {
		return true
	}
	return false
}

func extractToken(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}

	return ""
}

func AdminID(r *http.Request) int64 {
	id, _ := r.Context().Value(AdminIDKey).(int64)
	return id
}
