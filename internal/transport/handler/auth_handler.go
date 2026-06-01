package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	svcauth "sing-box-web-panel/internal/services/auth"
	"sing-box-web-panel/internal/transport/middleware"
)

const maxBodySize = 1 << 14

type AuthHandler struct {
	svc *svcauth.Service
	log *slog.Logger
}

func NewAuthHandler(svc *svcauth.Service, log *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

func (h *AuthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/login/totp", h.LoginTOTP)
	mux.HandleFunc("POST /api/auth/login/recovery", h.LoginRecovery)
	mux.HandleFunc("GET /api/auth/me", h.withAuth(h.Me))
	mux.HandleFunc("POST /api/auth/logout", h.Logout)
	mux.HandleFunc("POST /api/auth/totp/setup", h.withAuth(h.SetupTOTP))
	mux.HandleFunc("POST /api/auth/totp/confirm", h.withAuth(h.ConfirmTOTP))
	mux.HandleFunc("POST /api/auth/totp/disable", h.withAuth(h.DisableTOTP))
	mux.HandleFunc("POST /api/auth/change-password", h.withAuth(h.ChangePassword))
}

func (h *AuthHandler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := middleware.AdminID(r)
		if id == 0 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next(w, r)
	}
}

type loginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"admin"`
}

// Login godoc
//
//	@Summary		Authenticate admin
//	@Description	Logs in with username and password. Returns a JWT token, or temp_token + requires_totp if 2FA is enabled.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		loginRequest	true	"Login credentials"
//	@Success		200		{object}	map[string]string	"token"
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]any		"requires_totp with temp_token"
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if err.Error() == "http: request body too large" {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password required"})
		return
	}

	result, err := h.svc.Login(r.Context(), req.Username, req.Password, "")
	if err != nil {
		if errors.Is(err, svcauth.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		h.log.Error("login", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if result.RequiresTOTP {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"requires_totp": true,
			"temp_token":    result.TempToken,
		})
		return
	}

	h.setTokenCookie(w, r, result.Token)
	writeJSON(w, http.StatusOK, map[string]string{"token": result.Token})
}

type loginTOTPRequest struct {
	TempToken string `json:"temp_token" example:"eyJhbG..."`
	Code      string `json:"code"       example:"123456"`
}

// LoginTOTP godoc
//
//	@Summary		Complete TOTP login
//	@Description	Completes 2FA login using the temp_token returned by /login and a TOTP code.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		loginTOTPRequest	true	"TOTP login payload"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/login/totp [post]
func (h *AuthHandler) LoginTOTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req loginTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.TempToken == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "temp_token and code required"})
		return
	}

	token, err := h.svc.LoginTOTP(r.Context(), req.TempToken, req.Code)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid totp code"})
		return
	}

	h.setTokenCookie(w, r, token)
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

type loginRecoveryRequest struct {
	Username string `json:"username" example:"admin"`
	Code     string `json:"code"     example:"1234-5678"`
}

// LoginRecovery godoc
//
//	@Summary		Login with recovery code
//	@Description	Authenticates using a one-time recovery code (bypasses TOTP).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		loginRecoveryRequest	true	"Recovery login credentials"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/login/recovery [post]
func (h *AuthHandler) LoginRecovery(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req loginRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Username == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and code required"})
		return
	}

	token, err := h.svc.LoginRecovery(r.Context(), req.Username, req.Code)
	if err != nil {
		if errors.Is(err, svcauth.ErrInvalidRecoveryCode) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid recovery code"})
			return
		}
		h.log.Error("login recovery", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	h.setTokenCookie(w, r, token)
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// Me godoc
//
//	@Summary		Get current admin profile
//	@Description	Returns the authenticated admin's profile without secrets.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	meResponse
//	@Failure		401	{object}	map[string]string
//	@Router			/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.AdminID(r)
	admin, err := h.svc.GetAdmin(r.Context(), adminID)
	if err != nil {
		h.log.Error("me", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		ID:            admin.ID,
		Username:      admin.Username,
		IsTOTPEnabled: admin.IsTOTPEnabled,
		TOTPConfirmed: admin.TOTPConfirmedAt,
		CreatedAt:     admin.CreatedAt,
	})
}

type meResponse struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	IsTOTPEnabled bool       `json:"is_totp_enabled"`
	TOTPConfirmed *time.Time `json:"totp_confirmed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Logout godoc
//
//	@Summary		Logout
//	@Description	Clears the auth token cookie.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// SetupTOTP godoc
//
//	@Summary		Initiate TOTP setup
//	@Description	Generates a TOTP secret and returns a QR code URI for the admin to scan.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Router			/auth/totp/setup [post]
func (h *AuthHandler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.AdminID(r)
	qrURI, err := h.svc.SetupTOTP(r.Context(), adminID)
	if err != nil {
		h.log.Error("setup totp", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"qr_uri": qrURI})
}

type confirmTOTPRequest struct {
	Code string `json:"code" example:"123456"`
}

// ConfirmTOTP godoc
//
//	@Summary		Confirm TOTP setup
//	@Description	Verifies the TOTP code and enables 2FA. Returns 8 one-time recovery codes.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		confirmTOTPRequest	true	"TOTP code"
//	@Success		200		{object}	map[string]any
//	@Failure		400		{object}	map[string]string
//	@Router			/auth/totp/confirm [post]
func (h *AuthHandler) ConfirmTOTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req confirmTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	adminID := middleware.AdminID(r)
	codes, err := h.svc.ConfirmTOTP(r.Context(), adminID, req.Code)
	if err != nil {
		if errors.Is(err, svcauth.ErrInvalidTOTP) || errors.Is(err, svcauth.ErrTOTPNotSetup) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		h.log.Error("confirm totp", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":        "totp enabled",
		"recovery_codes": codes,
	})
}

type disableTOTPRequest struct {
	Code string `json:"code" example:"123456"`
}

// DisableTOTP godoc
//
//	@Summary		Disable TOTP
//	@Description	Disables 2FA. Requires a valid TOTP code or recovery code.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		disableTOTPRequest	true	"TOTP or recovery code"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Router			/auth/totp/disable [post]
func (h *AuthHandler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req disableTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	adminID := middleware.AdminID(r)
	if err := h.svc.DisableTOTP(r.Context(), adminID, req.Code); err != nil {
		if errors.Is(err, svcauth.ErrInvalidTOTP) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		h.log.Error("disable totp", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "totp disabled"})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" example:"oldpass"`
	NewPassword     string `json:"new_password"     example:"newpass"`
}

// ChangePassword godoc
//
//	@Summary		Change admin password
//	@Description	Changes the password for the currently authenticated admin.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		changePasswordRequest	true	"Password change payload"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Router			/auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	adminID := middleware.AdminID(r)
	if err := h.svc.ChangePassword(r.Context(), adminID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, svcauth.ErrInvalidCredentials) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid current password"})
			return
		}
		h.log.Error("change password", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password changed"})
}

func (h *AuthHandler) setTokenCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
