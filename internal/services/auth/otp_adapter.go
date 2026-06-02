package auth

import (
	"fmt"

	blauth "sing-box-web-panel/internal/lib/auth"
)

type TOTPAdapter struct {
	manager *blauth.TOTPManager
}

func NewTOTPAdapter(manager *blauth.TOTPManager) *TOTPAdapter {
	return &TOTPAdapter{manager: manager}
}

func (a *TOTPAdapter) GenerateSecret(username string) (*otpKey, error) {
	key, err := a.manager.GenerateSecret(username)
	if err != nil {
		return nil, fmt.Errorf("generate totp secret: %w", err)
	}

	secret, err := a.manager.SecretFromURI(key.URL())
	if err != nil {
		return nil, fmt.Errorf("extract secret: %w", err)
	}

	return &otpKey{
		Secret: secret,
		URL:    key.URL(),
	}, nil
}

func (a *TOTPAdapter) BuildKeyURI(username, secret string) string {
	return a.manager.BuildKeyURI(username, secret)
}

func (a *TOTPAdapter) Validate(code, secret string) bool {
	return a.manager.Validate(code, secret)
}
