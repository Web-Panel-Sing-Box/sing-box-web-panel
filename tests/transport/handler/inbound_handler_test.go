package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
	svcinbound "sing-box-web-panel/internal/services/inbound"
	"sing-box-web-panel/internal/transport/handler"
)

type inboundFakeRepo struct {
	items  map[int64]*domain.Inbound
	nextID int64
}

func newInboundFakeRepo() *inboundFakeRepo {
	return &inboundFakeRepo{items: map[int64]*domain.Inbound{}}
}

func (r *inboundFakeRepo) Create(_ context.Context, ib *domain.Inbound) error {
	r.nextID++
	ib.ID = r.nextID
	cp := *ib
	r.items[ib.ID] = &cp
	return nil
}

func (r *inboundFakeRepo) GetByID(_ context.Context, id int64) (*domain.Inbound, error) {
	if ib, ok := r.items[id]; ok {
		cp := *ib
		return &cp, nil
	}
	return nil, repo.ErrNotFound
}

func (r *inboundFakeRepo) List(context.Context) ([]domain.Inbound, error) { return nil, nil }

func (r *inboundFakeRepo) Update(_ context.Context, ib *domain.Inbound) error {
	cp := *ib
	r.items[ib.ID] = &cp
	return nil
}

func (r *inboundFakeRepo) SetEnabled(_ context.Context, id int64, enabled bool) error {
	if ib, ok := r.items[id]; ok {
		ib.Enabled = enabled
		return nil
	}
	return repo.ErrNotFound
}

func (r *inboundFakeRepo) Delete(_ context.Context, id int64) error {
	delete(r.items, id)
	return nil
}

type inboundFakeCounter struct{}

func (inboundFakeCounter) CountByInbound(context.Context) (map[int64]int, error) {
	return map[int64]int{}, nil
}

func TestInboundHandlerCreateAllowInsecureExplicitFalse(t *testing.T) {
	repo := newInboundFakeRepo()
	mux := testInboundMux(repo)

	body, _ := json.Marshal(map[string]any{
		"remark":        "hy2",
		"protocol":      "hysteria2",
		"port":          8443,
		"tls":           "tls",
		"allowInsecure": false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/inbounds", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
	}
	if repo.items[1].Settings.AllowInsecure == nil || *repo.items[1].Settings.AllowInsecure {
		t.Fatalf("stored allow insecure = %#v, want explicit false", repo.items[1].Settings.AllowInsecure)
	}
	settings := responseSettings(t, rec.Body.Bytes())
	if got := settings["allowInsecure"]; got != false {
		t.Fatalf("response settings.allowInsecure = %#v, want false", got)
	}
}

func TestInboundHandlerCreateAllowInsecureAutoResponse(t *testing.T) {
	mux := testInboundMux(newInboundFakeRepo())

	body, _ := json.Marshal(map[string]any{
		"remark":   "hy2",
		"protocol": "hysteria2",
		"port":     8443,
		"tls":      "tls",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/inbounds", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
	}
	settings := responseSettings(t, rec.Body.Bytes())
	if got := settings["allowInsecure"]; got != true {
		t.Fatalf("response settings.allowInsecure = %#v, want true", got)
	}
}

func testInboundMux(repo *inboundFakeRepo) *http.ServeMux {
	svc := svcinbound.NewService(repo, inboundFakeCounter{}, nil)
	h := handler.NewInboundHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func responseSettings(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	settings, ok := resp["settings"].(map[string]any)
	if !ok {
		t.Fatalf("settings missing or invalid in response: %s", body)
	}
	return settings
}
