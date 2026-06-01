package auth_test

import (
	"strings"
	"testing"
	"time"

	"sing-box-web-panel/internal/lib/auth"

	"github.com/pquerna/otp/totp"
)

func TestTOTPManager_GenerateSecret(t *testing.T) {
	tm := auth.NewTOTPManager("Shilka")

	key, err := tm.GenerateSecret("admin")
	if err != nil {
		t.Fatalf("GenerateSecret() error: %v", err)
	}

	if key.Secret() == "" {
		t.Error("secret should not be empty")
	}

	if !strings.HasPrefix(key.URL(), "otpauth://totp/") {
		t.Errorf("URL should start with otpauth://totp/, got: %s", key.URL())
	}

	if !strings.Contains(key.URL(), "Shilka") {
		t.Error("URL should contain issuer")
	}
}

func TestTOTPManager_Validate(t *testing.T) {
	tm := auth.NewTOTPManager("Shilka")

	key, err := tm.GenerateSecret("admin")
	if err != nil {
		t.Fatalf("GenerateSecret() error: %v", err)
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Fatalf("totp.GenerateCode() error: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("code length = %d, want 6", len(code))
	}

	if !tm.Validate(code, key.Secret()) {
		t.Error("Validate() should return true for valid code")
	}

	if tm.Validate("000000", key.Secret()) {
		t.Error("Validate() should return false for wrong code")
	}
}

func TestTOTPManager_ValidateEmptySecret(t *testing.T) {
	tm := auth.NewTOTPManager("Shilka")

	if tm.Validate("123456", "") {
		t.Error("Validate() should return false for empty secret")
	}
}

func TestTOTPManager_ValidateInvalidSecret(t *testing.T) {
	tm := auth.NewTOTPManager("Shilka")

	if tm.Validate("123456", "not-base32-!!!") {
		t.Error("Validate() should return false for invalid base32 secret")
	}
}

func TestGenerateRecoveryCode(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := auth.GenerateRecoveryCode()
		if err != nil {
			t.Fatalf("GenerateRecoveryCode() error: %v", err)
		}

		if len(code) != 9 {
			t.Errorf("code length = %d, want 9", len(code))
		}

		if code[4] != '-' {
			t.Errorf("code[4] = %c, want '-'", code[4])
		}

		if codes[code] {
			t.Errorf("duplicate code generated: %s", code)
		}
		codes[code] = true
	}
}

func TestTOTPManager_SecretFromURI(t *testing.T) {
	tm := auth.NewTOTPManager("Shilka")

	key, err := tm.GenerateSecret("admin")
	if err != nil {
		t.Fatalf("GenerateSecret() error: %v", err)
	}

	secret, err := tm.SecretFromURI(key.URL())
	if err != nil {
		t.Fatalf("SecretFromURI() error: %v", err)
	}

	if secret == "" {
		t.Error("secret should not be empty")
	}
}

func TestTOTPManager_SecretFromURIInvalid(t *testing.T) {
	tm := auth.NewTOTPManager("Shilka")

	_, err := tm.SecretFromURI("not-a-valid-uri")
	if err == nil {
		t.Error("SecretFromURI() should return error for invalid URI")
	}
}
