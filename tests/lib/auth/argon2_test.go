package auth_test

import (
	"strings"
	"testing"

	"sing-box-web-panel/internal/lib/auth"
)

func TestArgon2Hasher_HashAndVerify(t *testing.T) {
	hasher := auth.NewArgon2Hasher(64*1024, 3, 2)

	hash, err := hasher.Hash("super-secret-password")
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("hash should start with $argon2id$v=19$, got: %s", hash)
	}

	ok, err := hasher.Verify("super-secret-password", hash)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if !ok {
		t.Error("Verify() should return true for correct password")
	}

	ok, err = hasher.Verify("wrong-password", hash)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if ok {
		t.Error("Verify() should return false for wrong password")
	}
}

func TestArgon2Hasher_VerifyInvalidFormat(t *testing.T) {
	hasher := auth.NewArgon2Hasher(64*1024, 3, 2)

	_, err := hasher.Verify("password", "not-a-valid-hash")
	if err == nil {
		t.Error("Verify() should return error for invalid hash format")
	}
}

func TestArgon2Hasher_Uniqueness(t *testing.T) {
	hasher := auth.NewArgon2Hasher(64*1024, 3, 2)

	h1, _ := hasher.Hash("password")
	h2, _ := hasher.Hash("password")

	if h1 == h2 {
		t.Error("hashes of the same password should be unique due to random salt")
	}
}

func TestArgon2Hasher_VerifyWithDifferentParams(t *testing.T) {
	h1 := auth.NewArgon2Hasher(16*1024, 1, 1)
	h2 := auth.NewArgon2Hasher(64*1024, 3, 2)

	hash, err := h1.Hash("password")
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	ok, err := h2.Verify("password", hash)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if !ok {
		t.Error("Verify() should work even when hasher params differ from stored params")
	}
}

func TestArgon2Hasher_EmptyPassword(t *testing.T) {
	hasher := auth.NewArgon2Hasher(64*1024, 3, 2)

	hash, err := hasher.Hash("")
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	ok, err := hasher.Verify("", hash)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if !ok {
		t.Error("Verify() should work with empty password")
	}
}
