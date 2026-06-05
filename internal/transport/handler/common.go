package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"sing-box-web-panel/internal/repo"
	svcclient "sing-box-web-panel/internal/services/client"
	svcinbound "sing-box-web-panel/internal/services/inbound"
	svcnode "sing-box-web-panel/internal/services/node"
)

const timeRFC3339 = time.RFC3339

// idParam parses the {id} path value as a positive int64.
func idParam(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func parsePositiveID(w http.ResponseWriter, raw string, message string) (int64, bool) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
		return 0, false
	}
	return id, true
}

func matchesNodeID(w http.ResponseWriter, raw string, nodeID int64) bool {
	id, ok := parsePositiveID(w, raw, "invalid nodeId")
	if !ok {
		return false
	}
	if id != nodeID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot move resource between nodes"})
		return false
	}
	return true
}

// decodeJSON decodes the request body into dst, writing a 400 on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	return true
}

// parseTimePtr parses an RFC3339 timestamp; an empty string yields nil.
func parseTimePtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// formatTimePtr renders an optional timestamp as RFC3339 (empty when nil).
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// writeServiceError maps service/repository errors onto HTTP status codes.
func writeServiceError(w http.ResponseWriter, log *slog.Logger, op string, err error) {
	switch {
	case errors.Is(err, repo.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, repo.ErrExist):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already exists"})
	case errors.Is(err, svcinbound.ErrValidation), errors.Is(err, svcclient.ErrValidation), errors.Is(err, svcnode.ErrValidation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, svcclient.ErrInboundMissing):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, svcnode.ErrRemote):
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "remote node error"})
	case errors.Is(err, svcnode.ErrNodeUnreachable):
		status := http.StatusBadGateway // 502
		detail := "unreachable"
		var ue *svcnode.UnreachableError
		if errors.As(err, &ue) {
			detail = ue.Detail
			if ue.Timeout {
				status = http.StatusGatewayTimeout // 504
			}
		}
		log.Warn(op, slog.String("reason", "node unreachable"), slog.String("detail", detail))
		writeJSON(w, status, map[string]string{"error": "node unreachable", "detail": detail})
	default:
		log.Error(op, slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}
