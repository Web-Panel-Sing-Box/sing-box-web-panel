package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sing-box-web-panel/internal/transport/middleware"
)

func TestLogger_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := middleware.Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "method=GET") {
		t.Error("log should contain method")
	}
	if !strings.Contains(output, "path=/health") {
		t.Error("log should contain path")
	}
	if !strings.Contains(output, "status=200") {
		t.Errorf("log should contain status, got: %s", output)
	}
	if !strings.Contains(output, "size=2") {
		t.Errorf("log should contain size, got: %s", output)
	}
	if !strings.Contains(output, "duration=") {
		t.Error("log should contain duration")
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("response should include X-Request-ID")
	}
	if !strings.Contains(output, "request_id=") {
		t.Error("log should contain request_id")
	}
}

func TestLogger_UsesSafeIncomingRequestID(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := middleware.Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "req_test_123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") != "req_test_123" {
		t.Fatalf("X-Request-ID = %q, want req_test_123", rec.Header().Get("X-Request-ID"))
	}
	if !strings.Contains(buf.String(), "request_id=req_test_123") {
		t.Fatalf("log should contain incoming request id, got: %s", buf.String())
	}
}

func TestLogger_LogsWarnFor4xx(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	handler := middleware.Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "WARN") || !strings.Contains(output, "status=404") {
		t.Errorf("4xx should be logged as WARN, got: %s", output)
	}
}

func TestLogger_LogsErrorFor5xx(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := middleware.Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/crash", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "ERROR") || !strings.Contains(output, "status=500") {
		t.Errorf("5xx should be logged as ERROR, got: %s", output)
	}
}
