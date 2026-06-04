package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
	"sing-box-web-panel/internal/services/sublink"
)

// ClientByToken resolves a client from its subscription token.
type ClientByToken interface {
	GetBySubToken(ctx context.Context, token string) (*domain.Client, error)
	GetByID(ctx context.Context, id int64) (*domain.Client, error)
}

// InboundByID resolves an inbound by id.
type InboundByID interface {
	GetByID(ctx context.Context, id int64) (*domain.Inbound, error)
}

// SettingGetter reads a single panel setting.
type SettingGetter interface {
	Get(ctx context.Context, key string) (string, error)
}

type SubscriptionHandler struct {
	clients     ClientByToken
	inbounds    InboundByID
	settings    SettingGetter
	subBaseURL  string
	defaultHost string
	log         *slog.Logger
}

func NewSubscriptionHandler(clients ClientByToken, inbounds InboundByID, settings SettingGetter, subBaseURL, defaultHost string, log *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		clients:     clients,
		inbounds:    inbounds,
		settings:    settings,
		subBaseURL:  strings.TrimRight(subBaseURL, "/"),
		defaultHost: defaultHost,
		log:         log,
	}
}

func (h *SubscriptionHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /sub/{token}", h.Serve)              // public
	mux.HandleFunc("GET /api/subscription/{token}", h.Serve) // public
	mux.HandleFunc("GET /api/clients/{id}/links", h.Links)   // authenticated
}

// Serve godoc
//
//	@Summary	Fetch a subscription (public)
//	@Description	Returns the client's connection config. ?format=base64|plain|json (default base64).
//	@Tags		subscription
//	@Produce	plain
//	@Param		token	path	string	true	"Subscription token"
//	@Param		format	query	string	false	"base64 | plain | json"
//	@Success	200	{string}	string
//	@Failure	404	{object}	map[string]string
//	@Router		/subscription/{token} [get]
func (h *SubscriptionHandler) Serve(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	client, err := h.clients.GetBySubToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.log.Error("subscription lookup", slog.String("error", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !client.Enabled || client.Status != domain.ClientStatusActive {
		http.Error(w, "subscription disabled", http.StatusForbidden)
		return
	}

	inbound, err := h.inbounds.GetByID(r.Context(), client.InboundID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	format := r.URL.Query().Get("format")
	result, err := sublink.Render(format, inbound, client, h.resolveHost(r.Context(), r))
	if err != nil {
		h.log.Error("render subscription", slog.String("error", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Profile-Update-Interval", "12")
	w.WriteHeader(http.StatusOK)
	w.Write(result.Body)
}

type clientLinksDTO struct {
	Link         string `json:"link"`
	ShareLink    string `json:"shareLink"`
	Subscription string `json:"subscription"`
}

// Links godoc
//
//	@Summary	Get a client's share link and subscription URL
//	@Tags		clients
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"Client ID"
//	@Success	200	{object}	clientLinksDTO
//	@Router		/clients/{id}/links [get]
func (h *SubscriptionHandler) Links(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	client, err := h.clients.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, h.log, "client links", err)
		return
	}
	inbound, err := h.inbounds.GetByID(r.Context(), client.InboundID)
	if err != nil {
		writeServiceError(w, h.log, "client links inbound", err)
		return
	}
	sub := "/sub/" + client.SubToken
	if h.subBaseURL != "" {
		sub = h.subBaseURL + "/sub/" + client.SubToken
	}
	link := sublink.BuildLink(inbound, client, h.resolveHost(r.Context(), r))
	writeJSON(w, http.StatusOK, clientLinksDTO{
		Link:         link,
		ShareLink:    link,
		Subscription: sub,
	})
}

func (h *SubscriptionHandler) resolveHost(ctx context.Context, r *http.Request) string {
	if h.settings != nil {
		if v, err := h.settings.Get(ctx, domain.SettingInboundHost); err == nil && v != "" {
			return v
		}
	}
	if h.defaultHost != "" {
		return h.defaultHost
	}
	return sublink.HostFromRequestHost(r.Host)
}
