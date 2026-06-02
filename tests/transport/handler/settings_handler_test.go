package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"sing-box-web-panel/internal/services/settings"
	"sing-box-web-panel/internal/transport/handler"
)

type settingsFakeRepo struct {
	mu   sync.RWMutex
	data map[string]string
}

func newSettingsFakeRepo() *settingsFakeRepo {
	return &settingsFakeRepo{data: make(map[string]string)}
}

func (r *settingsFakeRepo) All(_ context.Context) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.data))
	for k, v := range r.data {
		out[k] = v
	}
	return out, nil
}

func (r *settingsFakeRepo) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
	return nil
}

type nopTrigger struct{}

func (nopTrigger) Trigger() {}

func testSettingsHandler(repo *settingsFakeRepo) *handler.SettingsHandler {
	svc := settings.New(repo, nopTrigger{})
	return handler.NewSettingsHandler(svc, slog.Default())
}

func TestSettingsHandler_Get_Empty(t *testing.T) {
	h := testSettingsHandler(newSettingsFakeRepo())
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("want empty map, got %v", resp)
	}
}

func TestSettingsHandler_Get_ReturnsSavedValues(t *testing.T) {
	repo := newSettingsFakeRepo()
	repo.Set(nil, "panel_name", "Test")
	repo.Set(nil, "log_level", "debug")

	h := testSettingsHandler(repo)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["panel_name"] != "Test" {
		t.Errorf("panel_name = %q, want Test", resp["panel_name"])
	}
	if resp["log_level"] != "debug" {
		t.Errorf("log_level = %q, want debug", resp["log_level"])
	}
}

func TestSettingsHandler_Put_SavesAndReturnsOk(t *testing.T) {
	repo := newSettingsFakeRepo()
	h := testSettingsHandler(repo)
	mux := http.NewServeMux()
	h.Register(mux)

	body, _ := json.Marshal(map[string]string{
		"panel_name": "MyPanel",
		"log_level":  "warn",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["ok"] != "saved" {
		t.Errorf("ok = %q, want saved", resp["ok"])
	}

	// Verify data persisted
	all, _ := repo.All(nil)
	if all["panel_name"] != "MyPanel" {
		t.Errorf("panel_name = %q, want MyPanel", all["panel_name"])
	}
	if all["log_level"] != "warn" {
		t.Errorf("log_level = %q, want warn", all["log_level"])
	}
}

func TestSettingsHandler_Put_InvalidJSON(t *testing.T) {
	repo := newSettingsFakeRepo()
	h := testSettingsHandler(repo)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader([]byte("not-json")))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestSettingsHandler_Put_Overwrites(t *testing.T) {
	repo := newSettingsFakeRepo()
	repo.Set(nil, "log_level", "info")

	h := testSettingsHandler(repo)
	mux := http.NewServeMux()
	h.Register(mux)

	body, _ := json.Marshal(map[string]string{"log_level": "error"})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	all, _ := repo.All(nil)
	if all["log_level"] != "error" {
		t.Errorf("log_level = %q, want error (overwritten)", all["log_level"])
	}
}
