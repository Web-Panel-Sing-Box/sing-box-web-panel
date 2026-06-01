package handler

import (
	"log/slog"
	"net/http"

	"sing-box-web-panel/internal/services/singbox"
)

type CoreHandler struct {
	pm      singbox.ProcessManager
	applier *singbox.Applier
	log     *slog.Logger
}

func NewCoreHandler(pm singbox.ProcessManager, applier *singbox.Applier, log *slog.Logger) *CoreHandler {
	return &CoreHandler{pm: pm, applier: applier, log: log}
}

func (h *CoreHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/core/status", h.Status)
	mux.HandleFunc("POST /api/core/start", h.Start)
	mux.HandleFunc("POST /api/core/stop", h.Stop)
	mux.HandleFunc("POST /api/core/restart", h.Restart)
	mux.HandleFunc("POST /api/core/reload", h.Reload)
	mux.HandleFunc("GET /api/core/version", h.Version)
	mux.HandleFunc("GET /api/core/config", h.Config)
}

type coreStatusDTO struct {
	Running       bool   `json:"running"`
	PID           int    `json:"pid"`
	Version       string `json:"version"`
	UptimeSeconds int64  `json:"uptimeSeconds"`
}

// Status godoc
//
//	@Summary	Core process status
//	@Tags		core
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	coreStatusDTO
//	@Router		/core/status [get]
func (h *CoreHandler) Status(w http.ResponseWriter, r *http.Request) {
	st, err := h.pm.Status(r.Context())
	if err != nil {
		h.log.Error("core status", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, coreStatusDTO{
		Running:       st.Running,
		PID:           st.PID,
		Version:       st.Version,
		UptimeSeconds: int64(st.Uptime.Seconds()),
	})
}

func (h *CoreHandler) action(w http.ResponseWriter, r *http.Request, op string, fn func() error) {
	if err := fn(); err != nil {
		h.log.Error("core "+op, slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": op + " ok"})
}

// Start godoc
//
//	@Summary	Start the core
//	@Description	Generates the initial config (if missing), then starts the core process.
//	@Tags		core
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	map[string]string
//	@Router		/core/start [post]
func (h *CoreHandler) Start(w http.ResponseWriter, r *http.Request) {
	if err := h.applier.ApplyIfMissing(r.Context()); err != nil {
		h.log.Error("core start: apply config", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "apply config: " + err.Error()})
		return
	}
	h.action(w, r, "start", func() error { return h.pm.Start(r.Context()) })
}

// Stop godoc
//
//	@Summary	Stop the core
//	@Tags		core
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	map[string]string
//	@Router		/core/stop [post]
func (h *CoreHandler) Stop(w http.ResponseWriter, r *http.Request) {
	h.action(w, r, "stop", func() error { return h.pm.Stop(r.Context()) })
}

// Restart godoc
//
//	@Summary	Restart the core
//	@Tags		core
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	map[string]string
//	@Router		/core/restart [post]
func (h *CoreHandler) Restart(w http.ResponseWriter, r *http.Request) {
	h.action(w, r, "restart", func() error { return h.pm.Restart(r.Context()) })
}

// Reload godoc
//
//	@Summary	Regenerate, validate and apply the config
//	@Description	Renders config from the database, runs `sing-box check`, installs it and reloads the core. A failed check returns 400 with the validation error and leaves the live config untouched.
//	@Tags		core
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	map[string]string
//	@Failure	400	{object}	map[string]string
//	@Router		/core/reload [post]
func (h *CoreHandler) Reload(w http.ResponseWriter, r *http.Request) {
	if err := h.applier.Apply(r.Context()); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "config applied"})
}

// Version godoc
//
//	@Summary	Core version
//	@Tags		core
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	map[string]string
//	@Router		/core/version [get]
func (h *CoreHandler) Version(w http.ResponseWriter, r *http.Request) {
	st, _ := h.pm.Status(r.Context())
	writeJSON(w, http.StatusOK, map[string]string{"version": st.Version})
}

// Config godoc
//
//	@Summary	Preview the generated config
//	@Tags		core
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	map[string]any
//	@Router		/core/config [get]
func (h *CoreHandler) Config(w http.ResponseWriter, r *http.Request) {
	data, err := h.applier.Preview(r.Context())
	if err != nil {
		h.log.Error("core config preview", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
