package handler

import (
	"net/http"
	"strconv"
	"strings"

	"sing-box-web-panel/internal/services/logbuf"
	"sing-box-web-panel/internal/transport/middleware"
)

type LogsHandler struct {
	buf *logbuf.Buffer
}

func NewLogsHandler(buf *logbuf.Buffer) *LogsHandler {
	return &LogsHandler{buf: buf}
}

func (h *LogsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/logs", h.List)
	mux.HandleFunc("POST /api/logs/frontend", h.Frontend)
}

type logEntryDTO struct {
	ID        string            `json:"id"`
	T         int64             `json:"t"`
	Level     string            `json:"level"`
	Source    string            `json:"source"`
	Message   string            `json:"message"`
	RequestID string            `json:"requestId,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// List godoc
//
//	@Summary	Recent log lines
//	@Description	Returns recent core and panel log lines from the in-memory ring buffer.
//	@Tags		logs
//	@Produce	json
//	@Security	BearerAuth
//	@Param		level	query	string	false	"info | warn | error"
//	@Param		source	query	string	false	"panel | core | frontend"
//	@Param		q		query	string	false	"substring filter"
//	@Param		limit	query	int		false	"max lines (default 200)"
//	@Success	200	{array}	logEntryDTO
//	@Router		/logs [get]
func (h *LogsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	entries := h.buf.Recent(limit, r.URL.Query().Get("level"), r.URL.Query().Get("source"), r.URL.Query().Get("q"))
	out := make([]logEntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, logEntryDTO{
			ID:        e.ID,
			T:         e.T,
			Level:     e.Level,
			Source:    e.Source,
			Message:   e.Message,
			RequestID: e.RequestID,
			Fields:    e.Fields,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type frontendLogRequest struct {
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

// Frontend godoc
//
//	@Summary	Record a frontend log line
//	@Description	Stores authenticated frontend runtime errors in the in-memory log buffer.
//	@Tags		logs
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		frontendLogRequest	true	"Log line"
//	@Success	202		{object}	map[string]string
//	@Failure	400		{object}	map[string]string
//	@Router		/logs/frontend [post]
func (h *LogsHandler) Frontend(w http.ResponseWriter, r *http.Request) {
	var req frontendLogRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	level := strings.ToLower(strings.TrimSpace(req.Level))
	if level == "" {
		level = "error"
	}
	if level != "info" && level != "warn" && level != "error" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid log level"})
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}
	if len(message) > 1000 {
		message = message[:1000]
	}
	if len(req.Fields) > 20 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "too many fields"})
		return
	}
	fields := make(map[string]string, len(req.Fields)+1)
	for k, v := range req.Fields {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if len(k) > 80 {
			k = k[:80]
		}
		if len(v) > 500 {
			v = v[:500]
		}
		fields[k] = v
	}
	if id := middleware.AdminID(r); id > 0 {
		fields["admin_id"] = strconv.FormatInt(id, 10)
	}
	h.buf.AppendEntry(logbuf.Entry{
		Level:     level,
		Source:    logbuf.SourceFrontend,
		Message:   message,
		RequestID: middleware.RequestID(r.Context()),
		Fields:    fields,
	})
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "logged"})
}
