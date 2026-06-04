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

const (
	SourcePanel    = "panel"
	SourceCore     = "core"
	SourceFrontend = "frontend"
)

// Entry is a single buffered log line.
type Entry struct {
	ID        string
	T         int64 // unix milliseconds
	Level     string
	Source    string
	Message   string
	RequestID string
	Fields    map[string]string
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
	b.AppendEntry(Entry{Level: level, Source: SourcePanel, Message: message})
}

// AppendEntry adds a structured log entry, evicting the oldest when full.
func (b *Buffer) AppendEntry(entry Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.seq++
	if entry.Level == "" {
		entry.Level = "info"
	}
	if entry.Source == "" {
		entry.Source = SourcePanel
	}
	entry.ID = strconv.FormatUint(b.seq, 10)
	entry.T = time.Now().UnixMilli()
	entry.Fields = RedactFields(entry.Fields)
	if len(b.entries) >= b.max {
		b.entries = append(b.entries[1:], entry)
	} else {
		b.entries = append(b.entries, entry)
	}
}

// Recent returns up to limit entries (newest last), optionally filtered by
// level/source and a case-insensitive substring query.
func (b *Buffer) Recent(limit int, level, source, query string) []Entry {
	b.mu.Lock()
	defer b.mu.Unlock()

	query = strings.ToLower(query)
	out := make([]Entry, 0, len(b.entries))
	for _, e := range b.entries {
		if level != "" && level != "all" && e.Level != level {
			continue
		}
		if source != "" && source != "all" && e.Source != source {
			continue
		}
		if query != "" && !matchesQuery(e, query) {
			continue
		}
		out = append(out, e)
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

func matchesQuery(e Entry, query string) bool {
	if strings.Contains(strings.ToLower(e.Message), query) ||
		strings.Contains(strings.ToLower(e.RequestID), query) ||
		strings.Contains(strings.ToLower(e.Source), query) {
		return true
	}
	for k, v := range e.Fields {
		if strings.Contains(strings.ToLower(k), query) || strings.Contains(strings.ToLower(v), query) {
			return true
		}
	}
	return false
}

// Writer returns an io.Writer that splits input into lines and appends them with
// a heuristically detected level. Suitable for a subprocess's stdout/stderr.
func (b *Buffer) Writer() *LineWriter {
	return b.SourceWriter(SourceCore)
}

// SourceWriter returns a line writer that tags all lines with source.
func (b *Buffer) SourceWriter(source string) *LineWriter {
	if source == "" {
		source = SourcePanel
	}
	return &LineWriter{buf: b, source: source}
}

// LineWriter accumulates bytes and flushes complete lines into the buffer.
type LineWriter struct {
	mu      sync.Mutex
	buf     *Buffer
	source  string
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
			w.buf.AppendEntry(Entry{Level: detectLevel(line), Source: w.source, Message: line})
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
	attrs []slog.Attr
}

// NewTeeHandler wraps inner so every emitted record is also captured in buf.
func NewTeeHandler(inner slog.Handler, buf *Buffer) *TeeHandler {
	return &TeeHandler{inner: inner, buf: buf}
}

func (h *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
	fields := make(map[string]string, r.NumAttrs()+len(h.attrs))
	for _, a := range h.attrs {
		addAttr(fields, "", a)
	}
	r.Attrs(func(a slog.Attr) bool {
		addAttr(fields, "", a)
		return true
	})
	requestID := fields["request_id"]
	delete(fields, "request_id")
	h.buf.AppendEntry(Entry{
		Level:     levelString(r.Level),
		Source:    SourcePanel,
		Message:   r.Message,
		RequestID: requestID,
		Fields:    fields,
	})
	return h.inner.Handle(ctx, r)
}

func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nextAttrs := append([]slog.Attr{}, h.attrs...)
	nextAttrs = append(nextAttrs, attrs...)
	return &TeeHandler{inner: h.inner.WithAttrs(attrs), buf: h.buf, attrs: nextAttrs}
}

func (h *TeeHandler) WithGroup(name string) slog.Handler {
	return &TeeHandler{inner: h.inner.WithGroup(name), buf: h.buf, attrs: h.attrs}
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

func addAttr(fields map[string]string, prefix string, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	key := attr.Key
	if prefix != "" {
		key = prefix + "." + key
	}
	if attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			addAttr(fields, key, child)
		}
		return
	}
	if key == "" {
		return
	}
	fields[key] = attrValueString(attr.Value)
}

func attrValueString(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time().UTC().Format(time.RFC3339)
	default:
		return fmt.Sprint(v.Any())
	}
}

// RedactFields replaces sensitive values while preserving field names.
func RedactFields(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if SensitiveKey(k) {
			out[k] = "[redacted]"
			continue
		}
		out[k] = v
	}
	return out
}

func SensitiveKey(key string) bool {
	k := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	k = strings.ReplaceAll(k, " ", "_")
	sensitive := []string{
		"password",
		"token",
		"secret",
		"authorization",
		"private_key",
		"privatekey",
	}
	for _, s := range sensitive {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}
