package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimit returns middleware that enforces a per-IP token-bucket limit on
// requests whose path satisfies match. It is used both for general API limiting
// and (with a stricter rate and a login-path matcher) for brute-force
// protection on the auth endpoints. rate is "N/s", "N/m" or "N/h".
func RateLimit(rate string, match func(path string) bool, log *slog.Logger) func(http.Handler) http.Handler {
	count, window, err := parseRate(rate)
	if err != nil {
		log.Warn("invalid rate limit, disabling", slog.String("rate", rate), slog.String("error", err.Error()))
		return func(next http.Handler) http.Handler { return next }
	}
	lim := newLimiter(count, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if match(r.URL.Path) && !lim.allow(clientIP(r)) {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// LoginPathMatcher matches the authentication endpoints subject to stricter
// brute-force limiting.
func LoginPathMatcher(path string) bool {
	return strings.HasPrefix(path, "/api/auth/login")
}

// APIPathMatcher matches all API endpoints for general rate limiting.
func APIPathMatcher(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

func parseRate(s string) (int, time.Duration, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected N/unit, got %q", s)
	}
	count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || count <= 0 {
		return 0, 0, fmt.Errorf("invalid count in %q", s)
	}
	var window time.Duration
	switch strings.TrimSpace(parts[1]) {
	case "s":
		window = time.Second
	case "m":
		window = time.Minute
	case "h":
		window = time.Hour
	default:
		return 0, 0, fmt.Errorf("invalid unit in %q", s)
	}
	return count, window, nil
}

type bucket struct {
	tokens float64
	last   time.Time
}

type limiter struct {
	mu       sync.Mutex
	visitors map[string]*bucket
	capacity float64
	refill   float64 // tokens per second
}

func newLimiter(count int, window time.Duration) *limiter {
	l := &limiter{
		visitors: make(map[string]*bucket),
		capacity: float64(count),
		refill:   float64(count) / window.Seconds(),
	}
	go l.cleanupLoop()
	return l
}

func (l *limiter) allow(ip string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.visitors[ip]
	if !ok {
		l.visitors[ip] = &bucket{tokens: l.capacity - 1, last: now}
		return true
	}
	b.tokens += now.Sub(b.last).Seconds() * l.refill
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// cleanupLoop evicts idle visitors so the map does not grow unbounded.
func (l *limiter) cleanupLoop() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-15 * time.Minute)
		l.mu.Lock()
		for ip, b := range l.visitors {
			if b.last.Before(cutoff) {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
