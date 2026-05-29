package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sing-box-web-panel/internal/domain"
	libauth "sing-box-web-panel/internal/lib/auth"
	"sing-box-web-panel/internal/repo"
	"sing-box-web-panel/internal/services/auth"
	"sing-box-web-panel/internal/transport/handler"
	"sing-box-web-panel/internal/transport/middleware"
)

type fakeAdminRepo struct {
	admins map[int64]*adminRecord
}

type adminRecord struct {
	id              int64
	username        string
	password        string
	totpSecret      string
	isTOTPEnabled   bool
	totpConfirmedAt *time.Time
}

func newFakeAdminRepo() *fakeAdminRepo {
	return &fakeAdminRepo{admins: make(map[int64]*adminRecord)}
}

func (r *fakeAdminRepo) Create(ctx context.Context, a *domain.Admin) error {
	r.admins[a.ID] = &adminRecord{
		id:              a.ID,
		username:        a.Username,
		password:        a.PasswordHash,
		totpSecret:      a.TOTPSecret,
		isTOTPEnabled:   a.IsTOTPEnabled,
		totpConfirmedAt: a.TOTPConfirmedAt,
	}
	return nil
}

func (r *fakeAdminRepo) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	rec, ok := r.admins[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return &domain.Admin{
		ID:              rec.id,
		Username:        rec.username,
		PasswordHash:    rec.password,
		TOTPSecret:      rec.totpSecret,
		IsTOTPEnabled:   rec.isTOTPEnabled,
		TOTPConfirmedAt: rec.totpConfirmedAt,
	}, nil
}

func (r *fakeAdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	for _, rec := range r.admins {
		if rec.username == username {
			return &domain.Admin{
				ID:              rec.id,
				Username:        rec.username,
				PasswordHash:    rec.password,
				TOTPSecret:      rec.totpSecret,
				IsTOTPEnabled:   rec.isTOTPEnabled,
				TOTPConfirmedAt: rec.totpConfirmedAt,
			}, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *fakeAdminRepo) Update(ctx context.Context, a *domain.Admin) error {
	if _, ok := r.admins[a.ID]; !ok {
		return repo.ErrNotFound
	}
	r.admins[a.ID] = &adminRecord{
		id:              a.ID,
		username:        a.Username,
		password:        a.PasswordHash,
		totpSecret:      a.TOTPSecret,
		isTOTPEnabled:   a.IsTOTPEnabled,
		totpConfirmedAt: a.TOTPConfirmedAt,
	}
	return nil
}

func (r *fakeAdminRepo) Count(ctx context.Context) (int, error) {
	return len(r.admins), nil
}

type fakeRecoveryRepo struct{}

func newFakeRecoveryRepo() *fakeRecoveryRepo {
	return &fakeRecoveryRepo{}
}

func (r *fakeRecoveryRepo) Create(ctx context.Context, code *domain.RecoveryCode) error  { return nil }
func (r *fakeRecoveryRepo) FindUnusedByAdminID(ctx context.Context, adminID int64) ([]domain.RecoveryCode, error) {
	return nil, nil
}
func (r *fakeRecoveryRepo) GetByAdminIDAndHash(ctx context.Context, adminID int64, codeHash string) (*domain.RecoveryCode, error) {
	return nil, repo.ErrNotFound
}
func (r *fakeRecoveryRepo) MarkUsed(ctx context.Context, id int64) error                     { return nil }
func (r *fakeRecoveryRepo) DeleteByAdminID(ctx context.Context, adminID int64) error         { return nil }

func testAuthService() *auth.Service {
	hasher := libauth.NewArgon2Hasher(64*1024, 3, 2)
	jwt := libauth.NewJWTManager("test-secret", time.Hour)
	totpAdapter := auth.NewTOTPAdapter(libauth.NewTOTPManager("Test"))

	adminRepo := newFakeAdminRepo()
	adminRepo.admins[1] = &adminRecord{
		id:       1,
		username: "admin",
		password: mustHash(hasher, "admin"),
	}

	return auth.NewService(adminRepo, newFakeRecoveryRepo(), hasher, jwt, totpAdapter, libauth.GenerateRecoveryCode)
}

func mustHash(hasher *libauth.Argon2Hasher, password string) string {
	hash, err := hasher.Hash(password)
	if err != nil {
		panic(err)
	}
	return hash
}

func withAdminID(r *http.Request, id int64) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.AdminIDKey, id)
	return r.WithContext(ctx)
}

func TestAuthHandler_Login_BadRequest(t *testing.T) {
	svc := testAuthService()
	h := handler.NewAuthHandler(svc, nil)

	mux := http.NewServeMux()
	h.Register(mux)

	body := bytes.NewReader([]byte(`not-json`))
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAuthHandler_Login_EmptyCredentials(t *testing.T) {
	svc := testAuthService()
	h := handler.NewAuthHandler(svc, nil)

	mux := http.NewServeMux()
	h.Register(mux)

	body, _ := json.Marshal(map[string]string{"username": "", "password": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := testAuthService()
	h := handler.NewAuthHandler(svc, nil)

	mux := http.NewServeMux()
	h.Register(mux)

	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "admin",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["token"] == "" {
		t.Error("response should contain a token")
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "token" {
		t.Error("response should set token cookie")
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc := testAuthService()
	h := handler.NewAuthHandler(svc, nil)

	mux := http.NewServeMux()
	h.Register(mux)

	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "wrong",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	svc := testAuthService()
	h := handler.NewAuthHandler(svc, nil)

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "token" && c.MaxAge < 0 {
			found = true
		}
	}
	if !found {
		t.Error("logout should clear the token cookie")
	}
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	svc := testAuthService()
	h := handler.NewAuthHandler(svc, nil)

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthHandler_Me_Success(t *testing.T) {
	svc := testAuthService()
	h := handler.NewAuthHandler(svc, nil)

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req = withAdminID(req, 1)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["username"] != "admin" {
		t.Errorf("username = %v, want admin", resp["username"])
	}
}

func TestAuthHandler_LoginRecovery(t *testing.T) {
	svc := testAuthService()
	h := handler.NewAuthHandler(svc, nil)

	mux := http.NewServeMux()
	h.Register(mux)

	body, _ := json.Marshal(map[string]string{"username": "admin", "code": "1234-5678"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login/recovery", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 (no recovery codes exist)", rec.Code)
	}
}
