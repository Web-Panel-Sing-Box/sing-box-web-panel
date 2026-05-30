package middleware_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"sing-box-web-panel/internal/transport/middleware"
)

func TestRateLimitBlocksAfterBurst(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mw := middleware.RateLimit("3/m", func(string) bool { return true }, log)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))

	codes := make([]int, 0, 5)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.RemoteAddr = "203.0.113.5:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		codes = append(codes, rec.Code)
	}

	// 3 allowed, then 429.
	for i := 0; i < 3; i++ {
		if codes[i] != http.StatusOK {
			t.Errorf("request %d: want 200, got %d", i, codes[i])
		}
	}
	if codes[3] != http.StatusTooManyRequests || codes[4] != http.StatusTooManyRequests {
		t.Errorf("requests 4,5: want 429, got %d,%d", codes[3], codes[4])
	}
}

func TestRateLimitPerIP(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mw := middleware.RateLimit("1/m", func(string) bool { return true }, log)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))

	call := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
		req.RemoteAddr = ip + ":1000"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := call("198.51.100.1"); got != http.StatusOK {
		t.Errorf("first ip first call: want 200, got %d", got)
	}
	// Different IP must have its own bucket.
	if got := call("198.51.100.2"); got != http.StatusOK {
		t.Errorf("second ip first call: want 200, got %d", got)
	}
	// First IP again exceeds its 1/m budget.
	if got := call("198.51.100.1"); got != http.StatusTooManyRequests {
		t.Errorf("first ip second call: want 429, got %d", got)
	}
}
