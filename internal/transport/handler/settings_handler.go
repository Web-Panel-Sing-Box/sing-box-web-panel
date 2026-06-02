package handler

import (
	"log/slog"
	"net/http"

	"sing-box-web-panel/internal/services/settings"
)

type SettingsHandler struct {
	svc *settings.Service
	log *slog.Logger
}

func NewSettingsHandler(svc *settings.Service, log *slog.Logger) *SettingsHandler {
	return &SettingsHandler{svc: svc, log: log}
}

func (h *SettingsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings", h.Get)
	mux.HandleFunc("PUT /api/settings", h.Put)
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	all, err := h.svc.All(r.Context())
	if err != nil {
		h.log.Error("settings get", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load settings"})
		return
	}
	writeJSON(w, http.StatusOK, all)
}

func (h *SettingsHandler) Put(w http.ResponseWriter, r *http.Request) {
	var patch map[string]string
	if !decodeJSON(w, r, &patch) {
		return
	}

	if err := h.svc.PatchAndRebuild(r.Context(), patch); err != nil {
		h.log.Error("settings put", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save settings"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "saved"})
}
