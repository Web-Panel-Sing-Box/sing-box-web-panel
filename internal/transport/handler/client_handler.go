package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sing-box-web-panel/internal/domain"
	svcclient "sing-box-web-panel/internal/services/client"
)

type ClientHandler struct {
	svc        *svcclient.Service
	subBaseURL string
	log        *slog.Logger
}

func NewClientHandler(svc *svcclient.Service, subBaseURL string, log *slog.Logger) *ClientHandler {
	return &ClientHandler{svc: svc, subBaseURL: strings.TrimRight(subBaseURL, "/"), log: log}
}

func (h *ClientHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/clients", h.List)
	mux.HandleFunc("POST /api/clients", h.Create)
	mux.HandleFunc("GET /api/clients/{id}", h.Get)
	mux.HandleFunc("PUT /api/clients/{id}", h.Update)
	mux.HandleFunc("DELETE /api/clients/{id}", h.Delete)
	mux.HandleFunc("POST /api/clients/{id}/reset-traffic", h.ResetTraffic)
	mux.HandleFunc("POST /api/clients/{id}/status", h.SetStatus)
}

func (h *ClientHandler) subURL(token string) string {
	if h.subBaseURL == "" {
		return "/sub/" + token
	}
	return h.subBaseURL + "/sub/" + token
}

type clientDTO struct {
	ID                 string `json:"id"`
	NodeID             string `json:"nodeId,omitempty"`
	RemoteID           string `json:"remoteId,omitempty"`
	Name               string `json:"name"`
	UUID               string `json:"uuid"`
	InboundID          string `json:"inboundId"`
	UsedDown           int64  `json:"usedDown"`
	UsedUp             int64  `json:"usedUp"`
	TotalQuota         int64  `json:"totalQuota"`
	Expiry             string `json:"expiry"`
	Status             string `json:"status"`
	Subscription       string `json:"subscription"`
	SubToken           string `json:"subToken,omitempty"`
	Enabled            bool   `json:"enabled"`
	StartAfterFirstUse bool   `json:"startAfterFirstUse"`
	Online             bool   `json:"online"`
}

func (h *ClientHandler) toDTO(c *domain.Client) clientDTO {
	dto := clientDTO{
		ID:                 strconv.FormatInt(c.ID, 10),
		RemoteID:           c.RemoteID,
		Name:               c.Name,
		UUID:               c.UUID,
		InboundID:          strconv.FormatInt(c.InboundID, 10),
		UsedDown:           c.UsedDown,
		UsedUp:             c.UsedUp,
		TotalQuota:         c.TotalQuota,
		Expiry:             formatTimePtr(c.Expiry),
		Status:             string(c.Status),
		Subscription:       h.subURL(c.SubToken),
		SubToken:           c.SubToken,
		Enabled:            c.Enabled,
		StartAfterFirstUse: c.StartAfterFirstUse,
		Online:             isOnline(c.LastUsedAt),
	}
	if c.NodeID != nil {
		dto.NodeID = strconv.FormatInt(*c.NodeID, 10)
	}
	return dto
}

// List godoc
//
//	@Summary	List clients
//	@Tags		clients
//	@Produce	json
//	@Security	BearerAuth
//	@Param		inboundId	query	int	false	"Filter by inbound ID"
//	@Success	200	{array}	clientDTO
//	@Router		/clients [get]
func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	var filter *int64
	if q := r.URL.Query().Get("inboundId"); q != "" {
		id, err := strconv.ParseInt(q, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid inboundId"})
			return
		}
		filter = &id
	}
	clients, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeServiceError(w, h.log, "list clients", err)
		return
	}
	out := make([]clientDTO, 0, len(clients))
	for i := range clients {
		out = append(out, h.toDTO(&clients[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

type createClientRequest struct {
	Name               string `json:"name"`
	InboundID          string `json:"inboundId"`
	TotalQuota         int64  `json:"totalQuota"`
	Expiry             string `json:"expiry"`
	StartAfterFirstUse bool   `json:"startAfterFirstUse"`
}

// Create godoc
//
//	@Summary	Create a client
//	@Tags		clients
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		createClientRequest	true	"Client"
//	@Success	201		{object}	clientDTO
//	@Router		/clients [post]
func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createClientRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	inboundID, err := strconv.ParseInt(req.InboundID, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid inboundId"})
		return
	}
	expiry, err := parseTimePtr(req.Expiry)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid expiry"})
		return
	}
	c, err := h.svc.Create(r.Context(), svcclient.CreateInput{
		Name:               req.Name,
		InboundID:          inboundID,
		TotalQuota:         req.TotalQuota,
		Expiry:             expiry,
		StartAfterFirstUse: req.StartAfterFirstUse,
	})
	if err != nil {
		writeServiceError(w, h.log, "create client", err)
		return
	}
	writeJSON(w, http.StatusCreated, h.toDTO(c))
}

// Get godoc
//
//	@Summary	Get a client
//	@Tags		clients
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"Client ID"
//	@Success	200	{object}	clientDTO
//	@Router		/clients/{id} [get]
func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeServiceError(w, h.log, "get client", err)
		return
	}
	writeJSON(w, http.StatusOK, h.toDTO(c))
}

type updateClientRequest struct {
	Name               *string `json:"name"`
	InboundID          *string `json:"inboundId"`
	TotalQuota         *int64  `json:"totalQuota"`
	Expiry             *string `json:"expiry"`
	Status             *string `json:"status"`
	StartAfterFirstUse *bool   `json:"startAfterFirstUse"`
}

// Update godoc
//
//	@Summary	Update a client
//	@Tags		clients
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		int					true	"Client ID"
//	@Param		request	body		updateClientRequest	true	"Patch"
//	@Success	200		{object}	clientDTO
//	@Router		/clients/{id} [put]
func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req updateClientRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	in := svcclient.UpdateInput{
		Name:               req.Name,
		TotalQuota:         req.TotalQuota,
		StartAfterFirstUse: req.StartAfterFirstUse,
	}
	if req.InboundID != nil {
		inboundID, err := strconv.ParseInt(*req.InboundID, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid inboundId"})
			return
		}
		in.InboundID = &inboundID
	}
	if req.Expiry != nil {
		expiry, err := parseTimePtr(*req.Expiry)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid expiry"})
			return
		}
		in.Expiry = expiry
	}
	if req.Status != nil {
		status := domain.ClientStatus(*req.Status)
		in.Status = &status
	}

	c, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeServiceError(w, h.log, "update client", err)
		return
	}
	writeJSON(w, http.StatusOK, h.toDTO(c))
}

// Delete godoc
//
//	@Summary	Delete a client
//	@Tags		clients
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Client ID"
//	@Success	200	{object}	map[string]string
//	@Router		/clients/{id} [delete]
func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeServiceError(w, h.log, "delete client", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ResetTraffic godoc
//
//	@Summary	Reset a client's traffic counters
//	@Tags		clients
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"Client ID"
//	@Success	200	{object}	clientDTO
//	@Router		/clients/{id}/reset-traffic [post]
func (h *ClientHandler) ResetTraffic(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	c, err := h.svc.ResetTraffic(r.Context(), id)
	if err != nil {
		writeServiceError(w, h.log, "reset client traffic", err)
		return
	}
	writeJSON(w, http.StatusOK, h.toDTO(c))
}

type setStatusRequest struct {
	Status string `json:"status"`
}

// SetStatus godoc
//
//	@Summary	Set a client's status
//	@Tags		clients
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		int					true	"Client ID"
//	@Param		request	body		setStatusRequest	true	"Status"
//	@Success	200		{object}	clientDTO
//	@Router		/clients/{id}/status [post]
func (h *ClientHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req setStatusRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	c, err := h.svc.SetStatus(r.Context(), id, domain.ClientStatus(req.Status))
	if err != nil {
		writeServiceError(w, h.log, "set client status", err)
		return
	}
	writeJSON(w, http.StatusOK, h.toDTO(c))
}

const onlineThreshold = 5 * time.Minute

func isOnline(lastUsedAt *time.Time) bool {
	return lastUsedAt != nil && time.Since(*lastUsedAt) < onlineThreshold
}
