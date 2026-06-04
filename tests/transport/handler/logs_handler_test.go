package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sing-box-web-panel/internal/services/logbuf"
	"sing-box-web-panel/internal/transport/handler"
)

func TestLogsHandler_ListFiltersStructuredEntries(t *testing.T) {
	buf := logbuf.New(10)
	buf.AppendEntry(logbuf.Entry{Level: "info", Source: logbuf.SourcePanel, Message: "server started"})
	buf.AppendEntry(logbuf.Entry{Level: "error", Source: logbuf.SourceFrontend, Message: "render failed", Fields: map[string]string{"component": "Dashboard"}})

	h := handler.NewLogsHandler(buf)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/logs?source=frontend&q=dashboard", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp) != 1 {
		t.Fatalf("logs = %d, want 1", len(resp))
	}
	if resp[0]["source"] != "frontend" {
		t.Fatalf("source = %v, want frontend", resp[0]["source"])
	}
}

func TestLogsHandler_FrontendAppendsRedactedEntry(t *testing.T) {
	buf := logbuf.New(10)
	h := handler.NewLogsHandler(buf)
	mux := http.NewServeMux()
	h.Register(mux)

	body := bytes.NewBufferString(`{"level":"error","message":"render failed","fields":{"token":"abc","component":"Dashboard"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/logs/frontend", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", rec.Code)
	}
	got := buf.Recent(10, "error", "frontend", "")
	if len(got) != 1 {
		t.Fatalf("logs = %d, want 1", len(got))
	}
	if got[0].Fields["token"] != "[redacted]" {
		t.Fatalf("token field = %q, want redacted", got[0].Fields["token"])
	}
	if got[0].Fields["component"] != "Dashboard" {
		t.Fatalf("component field = %q", got[0].Fields["component"])
	}
}
