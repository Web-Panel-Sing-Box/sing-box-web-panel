package service_test

import (
	"context"
	"testing"
	"time"

	"sing-box-web-panel/internal/domain"
	libauth "sing-box-web-panel/internal/lib/auth"
	"sing-box-web-panel/internal/repo"
	"sing-box-web-panel/internal/services/auth"
)

type mockAdminRepo struct {
	admins map[int64]*domain.Admin
}

func newMockAdminRepo() *mockAdminRepo {
	return &mockAdminRepo{admins: make(map[int64]*domain.Admin)}
}

func (m *mockAdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	admin.ID = int64(len(m.admins) + 1)
	m.admins[admin.ID] = admin
	return nil
}

func (m *mockAdminRepo) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	admin, ok := m.admins[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return admin, nil
}

func (m *mockAdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.Username == username {
			return a, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (m *mockAdminRepo) Update(ctx context.Context, admin *domain.Admin) error {
	if _, ok := m.admins[admin.ID]; !ok {
		return repo.ErrNotFound
	}
	m.admins[admin.ID] = admin
	return nil
}

func (m *mockAdminRepo) Count(ctx context.Context) (int, error) {
	return len(m.admins), nil
}

type mockRecoveryRepo struct {
	codes []domain.RecoveryCode
}

func newMockRecoveryRepo() *mockRecoveryRepo {
	return &mockRecoveryRepo{}
}

func (m *mockRecoveryRepo) Create(ctx context.Context, code *domain.RecoveryCode) error {
	code.ID = int64(len(m.codes) + 1)
	m.codes = append(m.codes, *code)
	return nil
}

func (m *mockRecoveryRepo) FindUnusedByAdminID(ctx context.Context, adminID int64) ([]domain.RecoveryCode, error) {
	var result []domain.RecoveryCode
	for _, c := range m.codes {
		if c.AdminID == adminID && !c.IsUsed {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockRecoveryRepo) GetByAdminIDAndHash(ctx context.Context, adminID int64, codeHash string) (*domain.RecoveryCode, error) {
	for i, c := range m.codes {
		if c.AdminID == adminID && c.CodeHash == codeHash && !c.IsUsed {
			return &m.codes[i], nil
		}
	}
	return nil, repo.ErrNotFound
}

func (m *mockRecoveryRepo) MarkUsed(ctx context.Context, id int64) error {
	for i := range m.codes {
		if m.codes[i].ID == id {
			m.codes[i].IsUsed = true
			return nil
		}
	}
	return repo.ErrNotFound
}

func (m *mockRecoveryRepo) DeleteByAdminID(ctx context.Context, adminID int64) error {
	filtered := m.codes[:0]
	for _, c := range m.codes {
		if c.AdminID != adminID {
			filtered = append(filtered, c)
		}
	}
	m.codes = filtered
	return nil
}

func newTestService() *auth.Service {
	hasher := libauth.NewArgon2Hasher(64*1024, 3, 2)
	jwt := libauth.NewJWTManager("test-secret", time.Hour)
	totp := auth.NewTOTPAdapter(libauth.NewTOTPManager("Test"))

	return auth.NewService(
		newMockAdminRepo(),
		newMockRecoveryRepo(),
		hasher,
		jwt,
		totp,
		libauth.GenerateRecoveryCode,
	)
}

func TestSeedAdmin(t *testing.T) {
	svc := newTestService()

	err := svc.SeedAdmin(context.Background(), "admin", "password123")
	if err != nil {
		t.Fatalf("SeedAdmin() error: %v", err)
	}

	admin, err := svc.GetAdmin(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetAdmin() error: %v", err)
	}

	if admin.Username != "admin" {
		t.Errorf("username = %q, want admin", admin.Username)
	}
}

func TestSeedAdmin_NoOpIfAlreadyExists(t *testing.T) {
	svc := newTestService()

	svc.SeedAdmin(context.Background(), "first", "pass1")
	svc.SeedAdmin(context.Background(), "second", "pass2")

	admin, _ := svc.GetAdmin(context.Background(), 1)
	if admin.Username != "first" {
		t.Errorf("SeedAdmin should not create second admin, got: %s", admin.Username)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestService()

	svc.SeedAdmin(context.Background(), "admin", "correct-password")

	result, err := svc.Login(context.Background(), "admin", "correct-password", "")
	if err != nil {
		t.Fatalf("Login() error: %v", err)
	}

	if result.Token == "" {
		t.Error("Login should return a token")
	}
	if result.RequiresTOTP {
		t.Error("Login should not require TOTP when not enabled")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	svc := newTestService()

	svc.SeedAdmin(context.Background(), "admin", "correct-password")

	_, err := svc.Login(context.Background(), "admin", "wrong-password", "")
	if err == nil {
		t.Error("Login should fail with wrong password")
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	svc := newTestService()

	_, err := svc.Login(context.Background(), "nobody", "password", "")
	if err == nil {
		t.Error("Login should fail for unknown user")
	}
}

func TestSetupTOTP(t *testing.T) {
	svc := newTestService()

	svc.SeedAdmin(context.Background(), "admin", "password")

	qrURI, err := svc.SetupTOTP(context.Background(), 1)
	if err != nil {
		t.Fatalf("SetupTOTP() error: %v", err)
	}

	if qrURI == "" {
		t.Error("SetupTOTP should return a QR URI")
	}

	admin, _ := svc.GetAdmin(context.Background(), 1)
	if admin.TOTPSecret == "" {
		t.Error("admin should have a TOTP secret after setup")
	}
	if admin.IsTOTPEnabled {
		t.Error("TOTP should not be enabled until confirmed")
	}
}

func TestChangePassword(t *testing.T) {
	svc := newTestService()

	svc.SeedAdmin(context.Background(), "admin", "old-password")

	err := svc.ChangePassword(context.Background(), 1, "old-password", "new-password")
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}

	_, err = svc.Login(context.Background(), "admin", "old-password", "")
	if err == nil {
		t.Error("old password should no longer work")
	}

	result, err := svc.Login(context.Background(), "admin", "new-password", "")
	if err != nil {
		t.Errorf("login with new password should succeed: %v", err)
	}
	if result.Token == "" {
		t.Error("new password login should return token")
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	svc := newTestService()

	svc.SeedAdmin(context.Background(), "admin", "correct")

	err := svc.ChangePassword(context.Background(), 1, "wrong", "new-password")
	if err == nil {
		t.Error("ChangePassword should fail with wrong current password")
	}
}

func TestGetAdmin_NotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.GetAdmin(context.Background(), 999)
	if err == nil {
		t.Error("GetAdmin should fail for nonexistent admin")
	}
}

func TestConfirmTOTP_Validation(t *testing.T) {
	svc := newTestService()
	svc.SeedAdmin(context.Background(), "admin", "password")
	svc.SetupTOTP(context.Background(), 1)

	_, err := svc.ConfirmTOTP(context.Background(), 1, "000000")
	if err == nil {
		t.Error("ConfirmTOTP should fail with wrong code")
	}
}

func TestDisableTOTP_NotSetup(t *testing.T) {
	svc := newTestService()
	svc.SeedAdmin(context.Background(), "admin", "password")

	err := svc.DisableTOTP(context.Background(), 1, "000000")
	if err == nil {
		t.Error("DisableTOTP should fail when TOTP is not enabled")
	}
}

func TestLoginRecovery_UserNotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.LoginRecovery(context.Background(), "nobody", "1234-5678")
	if err == nil {
		t.Error("LoginRecovery should fail for unknown user")
	}
}

func TestVerifyRecoveryCode_NoCodes(t *testing.T) {
	svc := newTestService()
	svc.SeedAdmin(context.Background(), "admin", "password")

	_, err := svc.VerifyRecoveryCode(context.Background(), 1, "1234-5678")
	if err == nil {
		t.Error("VerifyRecoveryCode should fail when no recovery codes exist")
	}
}
