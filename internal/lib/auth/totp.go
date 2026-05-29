package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTPManager struct {
	issuer string
	period uint
	digits otp.Digits
}

func NewTOTPManager(issuer string) *TOTPManager {
	return &TOTPManager{
		issuer: issuer,
		period: 30,
		digits: otp.DigitsSix,
	}
}

func (m *TOTPManager) GenerateSecret(username string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      m.issuer,
		AccountName: username,
		Period:      m.period,
		Digits:      m.digits,
	})
}

func (m *TOTPManager) Validate(code string, secret string) bool {
	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now(),
		totp.ValidateOpts{
			Period:    m.period,
			Digits:    m.digits,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	return err == nil && valid
}

func (m *TOTPManager) SecretFromURI(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("parse otpauth uri: %w", err)
	}

	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", fmt.Errorf("parse query: %w", err)
	}

	secret := query.Get("secret")
	if secret == "" {
		return "", fmt.Errorf("secret not found in otpauth uri")
	}

	return secret, nil
}

func (m *TOTPManager) IsValidSecret(secret string) bool {
	if secret == "" {
		return false
	}

	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		decoded, err = base32.StdEncoding.DecodeString(secret)
		if err != nil {
			return false
		}
	}

	return len(decoded) >= 10
}

func GenerateRecoveryCode() (string, error) {
	part1, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("generate recovery code: %w", err)
	}
	part2, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("generate recovery code: %w", err)
	}
	return fmt.Sprintf("%04d-%04d", part1.Int64(), part2.Int64()), nil
}
