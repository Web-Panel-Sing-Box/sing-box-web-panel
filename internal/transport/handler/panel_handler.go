package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"sing-box-web-panel/internal/services/updater"
)

type PanelHandler struct {
	updates *updater.Service
	log     *slog.Logger
}

func NewPanelHandler(updates *updater.Service, log *slog.Logger) *PanelHandler {
	return &PanelHandler{updates: updates, log: log}
}

func (h *PanelHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/panel/version", h.Version)
	mux.HandleFunc("POST /api/panel/update", h.Update)
}

type panelVersionDTO struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseURL      string `json:"releaseURL"`
	CheckedAt       string `json:"checkedAt"`
	Status          string `json:"status"`
}

// Version godoc
//
//	@Summary	Panel version and update status
//	@Tags		panel
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	panelVersionDTO
//	@Router		/panel/version [get]
func (h *PanelHandler) Version(w http.ResponseWriter, r *http.Request) {
	status, err := h.updates.VersionStatus(r.Context())
	if err != nil {
		h.log.Error("panel version", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, toPanelVersionDTO(status))
}

// Update godoc
//
//	@Summary	Start panel update
//	@Description	Starts the configured update helper asynchronously. Returns 409 if an update is already running.
//	@Tags		panel
//	@Produce	json
//	@Security	BearerAuth
//	@Success	202	{object}	panelVersionDTO
//	@Failure	409	{object}	map[string]string
//	@Failure	503	{object}	map[string]string
//	@Router		/panel/update [post]
func (h *PanelHandler) Update(w http.ResponseWriter, r *http.Request) {
	status, err := h.updates.StartUpdate(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, updater.ErrAlreadyRunning):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "update already running"})
		case errors.Is(err, updater.ErrNotConfigured):
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "update helper is not configured"})
		default:
			h.log.Error("panel update", slog.String("error", err.Error()))
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return
	}
	writeJSON(w, http.StatusAccepted, toPanelVersionDTO(status))
}

func toPanelVersionDTO(status updater.Status) panelVersionDTO {
	checkedAt := ""
	if !status.CheckedAt.IsZero() {
		checkedAt = status.CheckedAt.UTC().Format(time.RFC3339)
	}
	return panelVersionDTO{
		CurrentVersion:  status.CurrentVersion,
		LatestVersion:   status.LatestVersion,
		UpdateAvailable: status.UpdateAvailable,
		ReleaseURL:      status.ReleaseURL,
		CheckedAt:       checkedAt,
		Status:          status.Status,
	}
}
