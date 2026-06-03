package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/services/scheduler"
)

type SchedulerHandler struct {
	svc *scheduler.Service
	log *slog.Logger
}

func NewSchedulerHandler(svc *scheduler.Service, log *slog.Logger) *SchedulerHandler {
	return &SchedulerHandler{svc: svc, log: log}
}

func (h *SchedulerHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/scheduled-tasks", h.List)
	mux.HandleFunc("POST /api/scheduled-tasks", h.Create)
	mux.HandleFunc("PUT /api/scheduled-tasks/{id}", h.Update)
	mux.HandleFunc("DELETE /api/scheduled-tasks/{id}", h.Delete)
}

type taskDTO struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CronExpr   string  `json:"cronExpr"`
	Action     string  `json:"action"`
	ParamsJSON string  `json:"paramsJson"`
	Enabled    bool    `json:"enabled"`
	LastRunAt  *string `json:"lastRunAt"`
	NextRunAt  *string `json:"nextRunAt"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

func toTaskDTO(t *domain.ScheduledTask) taskDTO {
	dto := taskDTO{
		ID:         strconv.FormatInt(t.ID, 10),
		Name:       t.Name,
		CronExpr:   t.CronExpr,
		Action:     t.Action,
		ParamsJSON: t.ParamsJSON,
		Enabled:    t.Enabled,
		CreatedAt:  t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  t.UpdatedAt.Format(time.RFC3339),
	}
	if t.LastRunAt != nil {
		s := t.LastRunAt.Format(time.RFC3339)
		dto.LastRunAt = &s
	}
	if t.NextRunAt != nil {
		s := t.NextRunAt.Format(time.RFC3339)
		dto.NextRunAt = &s
	}
	return dto
}

func (h *SchedulerHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.svc.List(r.Context())
	if err != nil {
		h.log.Error("list tasks", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	out := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		out = append(out, toTaskDTO(&tasks[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

type createTaskRequest struct {
	Name       string `json:"name"`
	CronExpr   string `json:"cronExpr"`
	Action     string `json:"action"`
	ParamsJSON string `json:"paramsJson"`
	Enabled    bool   `json:"enabled"`
}

func (h *SchedulerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name == "" || req.CronExpr == "" || req.Action == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, cronExpr, action required"})
		return
	}
	if req.ParamsJSON == "" {
		req.ParamsJSON = "{}"
	}

	task := &domain.ScheduledTask{
		Name: req.Name, CronExpr: req.CronExpr, Action: req.Action,
		ParamsJSON: req.ParamsJSON, Enabled: req.Enabled,
	}

	if err := h.svc.Create(r.Context(), task); err != nil {
		h.log.Error("create task", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, toTaskDTO(task))
}

type updateTaskRequest struct {
	Name       string `json:"name"`
	CronExpr   string `json:"cronExpr"`
	Action     string `json:"action"`
	ParamsJSON string `json:"paramsJson"`
	Enabled    *bool  `json:"enabled"`
}

func (h *SchedulerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req updateTaskRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	task, err := h.svc.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	var existing *domain.ScheduledTask
	for i := range task {
		if task[i].ID == id {
			existing = &task[i]
			break
		}
	}
	if existing == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.CronExpr != "" {
		existing.CronExpr = req.CronExpr
	}
	if req.Action != "" {
		existing.Action = req.Action
	}
	if req.ParamsJSON != "" {
		existing.ParamsJSON = req.ParamsJSON
	} else {
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(req.ParamsJSON), &raw); err == nil {
			existing.ParamsJSON = req.ParamsJSON
		}
		existing.ParamsJSON = req.ParamsJSON
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	if err := h.svc.Update(r.Context(), existing); err != nil {
		h.log.Error("update task", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, toTaskDTO(existing))
}

func (h *SchedulerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.log.Error("delete task", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
