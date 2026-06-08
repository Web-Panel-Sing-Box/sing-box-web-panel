package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sing-box-web-panel/internal/services/singbox"
	"sing-box-web-panel/internal/transport/handler"
)

type fakeProcessManager struct {
	status singbox.Status
}

func (f fakeProcessManager) Start(context.Context) error   { return nil }
func (f fakeProcessManager) Stop(context.Context) error    { return nil }
func (f fakeProcessManager) Restart(context.Context) error { return nil }
func (f fakeProcessManager) Reload(context.Context) error  { return nil }
func (f fakeProcessManager) Status(context.Context) (singbox.Status, error) {
	return f.status, nil
}

func TestCoreStatusIncludesLastError(t *testing.T) {
	pm := fakeProcessManager{status: singbox.Status{
		Running:   false,
		Version:   "sing-box 1.2.3",
		Uptime:    5 * time.Second,
		LastError: "clash-api: bind 127.0.0.1:9090: address already in use",
	}}
	h := handler.NewCoreHandler(pm, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/core/status", nil)
	h.Status(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["lastError"] != "clash-api: bind 127.0.0.1:9090: address already in use" {
		t.Fatalf("lastError = %v", body["lastError"])
	}
}
