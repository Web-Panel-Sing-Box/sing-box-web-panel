package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidTOTP         = errors.New("invalid totp code")
	ErrInvalidRecoveryCode = errors.New("invalid recovery code")
	ErrTOTPNotSetup        = errors.New("totp not yet set up")
)
