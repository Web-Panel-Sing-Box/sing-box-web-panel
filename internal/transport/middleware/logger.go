package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type requestIDKey struct{}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := requestIDFromHeader(r.Header.Get("X-Request-ID"))
			if requestID == "" {
				requestID = newRequestID()
			}
			w.Header().Set("X-Request-ID", requestID)
			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
			r = r.WithContext(ctx)
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Int("size", rw.size),
				slog.Duration("duration", duration),
				slog.String("remote", r.RemoteAddr),
				slog.String("request_id", requestID),
			}

			switch {
			case rw.status >= 500:
				log.LogAttrs(r.Context(), slog.LevelError, "request", attrs...)
			case rw.status >= 400:
				log.LogAttrs(r.Context(), slog.LevelWarn, "request", attrs...)
			default:
				log.LogAttrs(r.Context(), slog.LevelInfo, "request", attrs...)
			}
		})
	}
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func requestIDFromHeader(id string) string {
	id = strings.TrimSpace(id)
	if len(id) == 0 || len(id) > 128 {
		return ""
	}
	for _, r := range id {
		if r < 33 || r > 126 {
			return ""
		}
	}
	return id
}

func newRequestID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return hex.EncodeToString([]byte(time.Now().UTC().Format("20060102150405.000000000")))
}
