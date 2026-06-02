package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"sing-box-web-panel/internal/lib/auth"
	"sing-box-web-panel/internal/transport/middleware"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

func TestAuth_PublicPaths(t *testing.T) {
	jwt := auth.NewJWTManager("secret", time.Hour)

	publicPaths := []string{"/", "/health", "/api/auth/login", "/api/auth/login/totp", "/api/auth/login/recovery", "/api/auth/logout"}

	for _, path := range publicPaths {
		t.Run(path, func(t *testing.T) {
			called := false
			handler := middleware.Auth(jwt, testLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}))

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if !called {
				t.Errorf("public path %s should pass through unauthorized", path)
			}
		})
	}
}

func TestAuth_SwaggerPath(t *testing.T) {
	jwt := auth.NewJWTManager("secret", time.Hour)

	called := false
	handler := middleware.Auth(jwt, testLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("swagger paths should pass through unauthorized")
	}
}

func TestAuth_ProtectedWithoutToken(t *testing.T) {
	jwt := auth.NewJWTManager("secret", time.Hour)

	handler := middleware.Auth(jwt, testLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ProtectedWithValidToken(t *testing.T) {
	jwt := auth.NewJWTManager("secret", time.Hour)
	token, _ := jwt.Create(42)

	called := false
	var capturedID int64
	handler := middleware.Auth(jwt, testLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		capturedID = middleware.AdminID(r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called with valid token")
	}
	if capturedID != 42 {
		t.Errorf("AdminID = %d, want 42", capturedID)
	}
}

func TestAuth_ProtectedWithValidCookie(t *testing.T) {
	jwt := auth.NewJWTManager("secret", time.Hour)
	token, _ := jwt.Create(7)

	called := false
	handler := middleware.Auth(jwt, testLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called with valid cookie")
	}
}

func TestAuth_ProtectedWithInvalidToken(t *testing.T) {
	jwt := auth.NewJWTManager("secret", time.Hour)

	handler := middleware.Auth(jwt, testLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer garbage-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
