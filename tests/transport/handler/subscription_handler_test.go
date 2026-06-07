package handler_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
	"sing-box-web-panel/internal/transport/handler"
)

type subscriptionFakeClients struct {
	byToken map[string]*domain.Client
	byID    map[int64]*domain.Client
}

func (r subscriptionFakeClients) GetBySubToken(_ context.Context, token string) (*domain.Client, error) {
	if c, ok := r.byToken[token]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, repo.ErrNotFound
}

func (r subscriptionFakeClients) GetByID(_ context.Context, id int64) (*domain.Client, error) {
	if c, ok := r.byID[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, repo.ErrNotFound
}

type subscriptionFakeInbounds struct {
	byID map[int64]*domain.Inbound
}

func (r subscriptionFakeInbounds) GetByID(_ context.Context, id int64) (*domain.Inbound, error) {
	if ib, ok := r.byID[id]; ok {
		cp := *ib
		return &cp, nil
	}
	return nil, repo.ErrNotFound
}

func testSubscriptionMux(ib *domain.Inbound, c *domain.Client) *http.ServeMux {
	clients := subscriptionFakeClients{
		byToken: map[string]*domain.Client{c.SubToken: c},
		byID:    map[int64]*domain.Client{c.ID: c},
	}
	inbounds := subscriptionFakeInbounds{byID: map[int64]*domain.Inbound{ib.ID: ib}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewSubscriptionHandler(clients, inbounds, nil, "", "", log)
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestSubscriptionNaiveJSONSelfSignedReturns400(t *testing.T) {
	ib := &domain.Inbound{
		ID: 7, Protocol: domain.ProtocolNaive, Port: 38119,
		TLS: domain.TLSModeTLS, SNI: "panel.example",
	}
	c := &domain.Client{
		ID: 9, InboundID: 7, Name: "carol", Password: "pw",
		Status: domain.ClientStatusActive, Enabled: true, SubToken: "tok",
	}
	mux := testSubscriptionMux(ib, c)

	req := httptest.NewRequest(http.MethodGet, "/sub/tok?format=json", nil)
	req.Host = "panel.example"
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "naive json subscription requires trusted TLS") {
		t.Fatalf("body = %q, want trusted TLS error", rec.Body.String())
	}
}

func TestSubscriptionNaivePlainStillReturnsShareLink(t *testing.T) {
	ib := &domain.Inbound{
		ID: 7, Protocol: domain.ProtocolNaive, Port: 38119,
		TLS: domain.TLSModeTLS,
	}
	c := &domain.Client{
		ID: 9, InboundID: 7, Name: "carol", Password: "pw",
		Status: domain.ClientStatusActive, Enabled: true, SubToken: "tok",
	}
	mux := testSubscriptionMux(ib, c)

	req := httptest.NewRequest(http.MethodGet, "/sub/tok?format=plain", nil)
	req.Host = "panel.example"
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if !strings.HasPrefix(rec.Body.String(), "naive+https://carol:pw@panel.example:38119") {
		t.Fatalf("link = %q", rec.Body.String())
	}
}
