package middleware

import (
	"context"
	"net/http"
	"strings"

	"sing-box-web-panel/internal/lib/auth"
)

type contextKey string

const AdminIDKey contextKey = "admin_id"

var publicPaths = map[string]bool{
	"/api/auth/login":          true,
	"/api/auth/login/recovery": true,
	"/api/auth/logout":         true,
	"/":                        true,
	"/health":                  true,
}

func Auth(jwt *auth.JWTManager) func(http.Handler) http.Handler {
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
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			adminID, err := jwt.Validate(token)
			if err != nil {
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
