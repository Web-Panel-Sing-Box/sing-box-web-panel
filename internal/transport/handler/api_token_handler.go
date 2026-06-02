package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/services/apitoken"
)

type APITokenHandler struct {
	svc *apitoken.Service
	log *slog.Logger
}

func NewAPITokenHandler(svc *apitoken.Service, log *slog.Logger) *APITokenHandler {
	return &APITokenHandler{svc: svc, log: log}
}

func (h *APITokenHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/api-tokens", h.List)
	mux.HandleFunc("POST /api/api-tokens", h.Create)
	mux.HandleFunc("POST /api/api-tokens/{id}/toggle", h.Toggle)
	mux.HandleFunc("DELETE /api/api-tokens/{id}", h.Delete)
}

type apiTokenDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	TokenPrefix string `json:"tokenPrefix"`
	Scopes      string `json:"scopes"`
	Enabled     bool   `json:"enabled"`
	LastUsedAt  string `json:"lastUsedAt,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

type createdAPITokenDTO struct {
	apiTokenDTO
	Token string `json:"token"`
}

func toAPITokenDTO(t *domain.APIToken) apiTokenDTO {
	return apiTokenDTO{
		ID:          strconv.FormatInt(t.ID, 10),
		Name:        t.Name,
		TokenPrefix: t.TokenPrefix,
		Scopes:      t.Scopes,
		Enabled:     t.Enabled,
		LastUsedAt:  formatTimePtr(t.LastUsedAt),
		CreatedAt:   t.CreatedAt.UTC().Format(timeRFC3339),
	}
}

type createAPITokenRequest struct {
	Name   string `json:"name"`
	Scopes string `json:"scopes"`
}

func (h *APITokenHandler) List(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.svc.List(r.Context())
	if err != nil {
		writeServiceError(w, h.log, "list api tokens", err)
		return
	}
	out := make([]apiTokenDTO, 0, len(tokens))
	for i := range tokens {
		out = append(out, toAPITokenDTO(&tokens[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *APITokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAPITokenRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	created, err := h.svc.Create(r.Context(), req.Name, req.Scopes)
	if err != nil {
		writeServiceError(w, h.log, "create api token", err)
		return
	}
	dto := createdAPITokenDTO{apiTokenDTO: toAPITokenDTO(&created.Token), Token: created.Raw}
	writeJSON(w, http.StatusCreated, dto)
}

func (h *APITokenHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.SetEnabled(r.Context(), id, req.Enabled); err != nil {
		writeServiceError(w, h.log, "toggle api token", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (h *APITokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeServiceError(w, h.log, "delete api token", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
