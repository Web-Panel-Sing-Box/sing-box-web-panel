package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sing-box-web-panel/internal/services/updater"
	"sing-box-web-panel/internal/transport/handler"
)

type panelFakeReleaseClient struct {
	release updater.Release
}

func (c panelFakeReleaseClient) Latest(context.Context, string) (updater.Release, error) {
	return c.release, nil
}

type panelFakeRunner struct{}

func (panelFakeRunner) Run(context.Context, string, time.Duration) ([]byte, error) {
	return []byte("ok"), nil
}

func TestPanelHandler_Version(t *testing.T) {
	svc := updater.New(updater.Config{
		Repo:           "owner/repo",
		CurrentVersion: "1.0.0",
	}, panelFakeReleaseClient{release: updater.Release{Version: "v1.2.0", URL: "https://example.test/release"}}, panelFakeRunner{}, handlerDiscardLogger())

	h := handler.NewPanelHandler(svc, handlerDiscardLogger())
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/panel/version", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["currentVersion"] != "1.0.0" {
		t.Fatalf("currentVersion = %v", resp["currentVersion"])
	}
	if resp["latestVersion"] != "1.2.0" {
		t.Fatalf("latestVersion = %v", resp["latestVersion"])
	}
	if resp["updateAvailable"] != true {
		t.Fatalf("updateAvailable = %v, want true", resp["updateAvailable"])
	}
}

func TestPanelHandler_UpdateStartsHelper(t *testing.T) {
	script := filepath.Join(t.TempDir(), "update")
	if err := os.WriteFile(script, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	svc := updater.New(updater.Config{
		ScriptPath:     script,
		CurrentVersion: "1.0.0",
	}, panelFakeReleaseClient{}, panelFakeRunner{}, handlerDiscardLogger())

	h := handler.NewPanelHandler(svc, handlerDiscardLogger())
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/panel/update", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", rec.Code)
	}
}

func TestPanelHandler_UpdateReportsMissingHelper(t *testing.T) {
	svc := updater.New(updater.Config{
		ScriptPath:     filepath.Join(t.TempDir(), "missing"),
		CurrentVersion: "1.0.0",
	}, panelFakeReleaseClient{}, panelFakeRunner{}, handlerDiscardLogger())

	h := handler.NewPanelHandler(svc, handlerDiscardLogger())
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/panel/update", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func handlerDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
