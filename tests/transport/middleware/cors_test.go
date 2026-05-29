package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sing-box-web-panel/internal/transport/middleware"
)

func TestCORS_OptionsPreflight(t *testing.T) {
	handler := middleware.CORS(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Allow-Methods header missing")
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("Allow-Origin header missing or wrong")
	}
}

func TestCORS_OptionsWithoutOrigin(t *testing.T) {
	handler := middleware.CORS(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestCORS_AllowedOrigin(t *testing.T) {
	allowedOrigins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}

	called := false
	handler := middleware.CORS(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should have been called for allowed origin")
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("Allow-Origin header missing or wrong")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	allowedOrigins := []string{"http://localhost:3000"}

	handler := middleware.CORS(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for disallowed origin")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCORS_SameOrigin(t *testing.T) {
	allowedOrigins := []string{"http://localhost:3000"}

	called := false
	handler := middleware.CORS(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("same-origin request should be allowed regardless of allowedOrigins")
	}
}

func TestCORS_SameOrigin_LocalhostVSLoopback(t *testing.T) {
	allowedOrigins := []string{"http://localhost:3000"}

	called := false
	handler := middleware.CORS(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("localhost and 127.0.0.1 should be treated as same-origin")
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	called := false
	handler := middleware.CORS([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("requests without Origin header should pass through")
	}
}
