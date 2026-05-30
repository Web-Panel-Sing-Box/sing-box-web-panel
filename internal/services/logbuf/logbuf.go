// Package logbuf keeps a bounded in-memory ring of recent log lines for the
// panel's Logs view. It captures panel logs through an slog tee handler and
// core logs through an io.Writer fed from the sing-box subprocess.
package logbuf

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Entry is a single buffered log line.
type Entry struct {
	ID      string
	T       int64 // unix milliseconds
	Level   string
	Message string
}

// Buffer is a concurrency-safe ring of the most recent entries.
type Buffer struct {
	mu      sync.Mutex
	entries []Entry
	max     int
	seq     uint64
}

func New(max int) *Buffer {
	if max <= 0 {
		max = 200
	}
	return &Buffer{max: max, entries: make([]Entry, 0, max)}
}

// Append adds a line, evicting the oldest when the buffer is full.
func (b *Buffer) Append(level, message string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.seq++
	e := Entry{
		ID:      strconv.FormatUint(b.seq, 10),
		T:       time.Now().UnixMilli(),
		Level:   level,
		Message: message,
	}
	if len(b.entries) >= b.max {
		b.entries = append(b.entries[1:], e)
	} else {
		b.entries = append(b.entries, e)
	}
}

// Recent returns up to limit entries (newest last), optionally filtered by
// level and a case-insensitive substring query.
func (b *Buffer) Recent(limit int, level, query string) []Entry {
	b.mu.Lock()
	defer b.mu.Unlock()

	query = strings.ToLower(query)
	out := make([]Entry, 0, len(b.entries))
	for _, e := range b.entries {
		if level != "" && level != "all" && e.Level != level {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(e.Message), query) {
			continue
		}
		out = append(out, e)
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

// Writer returns an io.Writer that splits input into lines and appends them with
// a heuristically detected level. Suitable for a subprocess's stdout/stderr.
func (b *Buffer) Writer() *LineWriter {
	return &LineWriter{buf: b}
}

// LineWriter accumulates bytes and flushes complete lines into the buffer.
type LineWriter struct {
	mu      sync.Mutex
	buf     *Buffer
	partial []byte
}

func (w *LineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.partial = append(w.partial, p...)
	for {
		i := bytes.IndexByte(w.partial, '\n')
		if i < 0 {
			break
		}
		line := strings.TrimRight(string(w.partial[:i]), "\r")
		w.partial = w.partial[i+1:]
		if line != "" {
			w.buf.Append(detectLevel(line), line)
		}
	}
	return len(p), nil
}

func detectLevel(line string) string {
	up := strings.ToUpper(line)
	switch {
	case strings.Contains(up, "ERROR"), strings.Contains(up, "FATAL"), strings.Contains(up, "PANIC"):
		return "error"
	case strings.Contains(up, "WARN"):
		return "warn"
	default:
		return "info"
	}
}

// --- slog tee handler ---

// TeeHandler forwards records to an inner handler and mirrors them into a Buffer.
type TeeHandler struct {
	inner slog.Handler
	buf   *Buffer
}

// NewTeeHandler wraps inner so every emitted record is also captured in buf.
func NewTeeHandler(inner slog.Handler, buf *Buffer) *TeeHandler {
	return &TeeHandler{inner: inner, buf: buf}
}

func (h *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
	var sb strings.Builder
	sb.WriteString(r.Message)
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&sb, " %s=%v", a.Key, a.Value.Any())
		return true
	})
	h.buf.Append(levelString(r.Level), sb.String())
	return h.inner.Handle(ctx, r)
}

func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TeeHandler{inner: h.inner.WithAttrs(attrs), buf: h.buf}
}

func (h *TeeHandler) WithGroup(name string) slog.Handler {
	return &TeeHandler{inner: h.inner.WithGroup(name), buf: h.buf}
}

func levelString(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "error"
	case l >= slog.LevelWarn:
		return "warn"
	default:
		return "info"
	}
}
