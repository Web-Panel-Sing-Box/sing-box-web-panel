package handler

import (
	"net/http"
	"strconv"

	"sing-box-web-panel/internal/services/logbuf"
)

type LogsHandler struct {
	buf *logbuf.Buffer
}

func NewLogsHandler(buf *logbuf.Buffer) *LogsHandler {
	return &LogsHandler{buf: buf}
}

func (h *LogsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/logs", h.List)
}

type logEntryDTO struct {
	ID      string `json:"id"`
	T       int64  `json:"t"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// List godoc
//
//	@Summary	Recent log lines
//	@Description	Returns recent core and panel log lines from the in-memory ring buffer.
//	@Tags		logs
//	@Produce	json
//	@Security	BearerAuth
//	@Param		level	query	string	false	"info | warn | error"
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
	entries := h.buf.Recent(limit, r.URL.Query().Get("level"), r.URL.Query().Get("q"))
	out := make([]logEntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, logEntryDTO{ID: e.ID, T: e.T, Level: e.Level, Message: e.Message})
	}
	writeJSON(w, http.StatusOK, out)
}
