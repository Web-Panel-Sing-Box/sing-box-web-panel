package domain

import "time"

type Admin struct {
	ID              int64
	Username        string
	PasswordHash    string
	TOTPSecret      string
	IsTOTPEnabled   bool
	TOTPConfirmedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type RecoveryCode struct {
	ID        int64
	AdminID   int64
	CodeHash  string
	IsUsed    bool
	UsedAt    *time.Time
	CreatedAt time.Time
}
