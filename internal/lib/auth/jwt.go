package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret []byte
	expiry time.Duration
}

func NewJWTManager(secret string, expiry time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		expiry: expiry,
	}
}

func (m *JWTManager) Create(adminID int64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", adminID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

type totpPendingClaims struct {
	jwt.RegisteredClaims
	TOTPPending bool `json:"totp_pending"`
}

func (m *JWTManager) CreateTOTPPending(adminID int64) (string, error) {
	now := time.Now()
	claims := totpPendingClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", adminID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
		TOTPPending: true,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ValidateTOTPPending(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &totpPendingClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*totpPendingClaims)
	if !ok || !token.Valid || !claims.TOTPPending {
		return 0, fmt.Errorf("invalid totp pending token")
	}

	var adminID int64
	if _, err := fmt.Sscanf(claims.Subject, "%d", &adminID); err != nil {
		return 0, fmt.Errorf("parse subject: %w", err)
	}

	return adminID, nil
}

func (m *JWTManager) Validate(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &totpPendingClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*totpPendingClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	if claims.TOTPPending {
		return 0, fmt.Errorf("token is totp-pending, not full auth")
	}

	var adminID int64
	if _, err := fmt.Sscanf(claims.Subject, "%d", &adminID); err != nil {
		return 0, fmt.Errorf("parse subject: %w", err)
	}

	return adminID, nil
}
