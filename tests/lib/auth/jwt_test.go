package auth_test

import (
	"testing"
	"time"

	"sing-box-web-panel/internal/lib/auth"
)

func TestJWTManager_CreateAndValidate(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Hour)

	token, err := mgr.Create(42)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	adminID, err := mgr.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	if adminID != 42 {
		t.Errorf("Validate() = %d, want 42", adminID)
	}
}

func TestJWTManager_ValidateExpiredToken(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", 0)

	token, err := mgr.Create(1)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	_, err = mgr.Validate(token)
	if err == nil {
		t.Error("Validate() should return error for expired token")
	}
}

func TestJWTManager_ValidateWithWrongSecret(t *testing.T) {
	mgr1 := auth.NewJWTManager("secret1", time.Hour)
	mgr2 := auth.NewJWTManager("secret2", time.Hour)

	token, _ := mgr1.Create(1)
	_, err := mgr2.Validate(token)
	if err == nil {
		t.Error("Validate() should fail with wrong secret")
	}
}

func TestJWTManager_ValidateTamperedToken(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Hour)

	token, _ := mgr.Create(1)
	tampered := token + "x"
	_, err := mgr.Validate(tampered)
	if err == nil {
		t.Error("Validate() should fail for tampered token")
	}
}

func TestJWTManager_CreateTOTPPending(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Hour)

	token, err := mgr.CreateTOTPPending(7)
	if err != nil {
		t.Fatalf("CreateTOTPPending() error: %v", err)
	}

	adminID, err := mgr.ValidateTOTPPending(token)
	if err != nil {
		t.Fatalf("ValidateTOTPPending() error: %v", err)
	}

	if adminID != 7 {
		t.Errorf("ValidateTOTPPending() = %d, want 7", adminID)
	}
}

func TestJWTManager_ValidateTOTPPendingWithFullToken(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Hour)

	token, _ := mgr.Create(1)
	_, err := mgr.ValidateTOTPPending(token)
	if err == nil {
		t.Error("ValidateTOTPPending() should fail for non-pending token")
	}
}

func TestJWTManager_ValidateShouldRejectTOTPPending(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Hour)

	token, _ := mgr.CreateTOTPPending(1)
	_, err := mgr.Validate(token)
	if err == nil {
		t.Error("Validate() should fail for TOTP pending token")
	}
}

func TestJWTManager_ValidateEmptyToken(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Hour)

	_, err := mgr.Validate("")
	if err == nil {
		t.Error("Validate() should fail for empty token")
	}
}

func TestJWTManager_DifferentAdminIDs(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Hour)

	for _, id := range []int64{0, 1, 999999, -1} {
		token, err := mgr.Create(id)
		if err != nil {
			t.Fatalf("Create(%d) error: %v", id, err)
		}

		got, err := mgr.Validate(token)
		if err != nil {
			t.Fatalf("Validate(%d) error: %v", id, err)
		}
		if got != id {
			t.Errorf("Validate(%d) = %d, want %d", id, got, id)
		}
	}
}
