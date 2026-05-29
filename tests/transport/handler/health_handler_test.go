package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sing-box-web-panel/internal/transport/handler"
)

func TestHealthHandler_Root(t *testing.T) {
	h := handler.NewHealthHandler()
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["name"] != "Singbox Web Panel" {
		t.Errorf("name = %q, want Singbox Web Panel", resp["name"])
	}

	if resp["version"] == "" {
		t.Error("version should not be empty")
	}
}

func TestHealthHandler_Health(t *testing.T) {
	h := handler.NewHealthHandler()
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["status"] != "ok" {
		t.Errorf("status = %q, want ok", resp["status"])
	}
}
