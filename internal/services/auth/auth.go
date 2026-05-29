package auth

import (
	"context"
	"fmt"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
)

type AdminRepository interface {
	Create(ctx context.Context, admin *domain.Admin) error
	GetByID(ctx context.Context, id int64) (*domain.Admin, error)
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)
	Update(ctx context.Context, admin *domain.Admin) error
	Count(ctx context.Context) (int, error)
}

type RecoveryCodeRepository interface {
	Create(ctx context.Context, code *domain.RecoveryCode) error
	FindUnusedByAdminID(ctx context.Context, adminID int64) ([]domain.RecoveryCode, error)
	GetByAdminIDAndHash(ctx context.Context, adminID int64, codeHash string) (*domain.RecoveryCode, error)
	MarkUsed(ctx context.Context, id int64) error
	DeleteByAdminID(ctx context.Context, adminID int64) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encoded string) (bool, error)
}

type TokenManager interface {
	Create(adminID int64) (string, error)
	Validate(tokenString string) (int64, error)
}

type TOTPProvider interface {
	GenerateSecret(username string) (*otpKey, error)
	Validate(code, secret string) bool
}

type otpKey struct {
	Secret string
	URL    string
}

type RecoveryCodeGenerator func() (string, error)

type LoginResult struct {
	Token        string `json:"token,omitempty"`
	RequiresTOTP bool   `json:"requires_totp"`
	AdminID      int64  `json:"-"`
}

type Service struct {
	admins         AdminRepository
	recoveryCodes  RecoveryCodeRepository
	passwordHasher PasswordHasher
	tokenManager   TokenManager
	totp           TOTPProvider
	generateCode   RecoveryCodeGenerator
}

func NewService(
	admins AdminRepository,
	recoveryCodes RecoveryCodeRepository,
	passwordHasher PasswordHasher,
	tokenManager TokenManager,
	totp TOTPProvider,
	generateCode RecoveryCodeGenerator,
) *Service {
	return &Service{
		admins:         admins,
		recoveryCodes:  recoveryCodes,
		passwordHasher: passwordHasher,
		tokenManager:   tokenManager,
		totp:           totp,
		generateCode:   generateCode,
	}
}

func (s *Service) Login(ctx context.Context, username, password, totpCode string) (*LoginResult, error) {
	admin, err := s.admins.GetByUsername(ctx, username)
	if err != nil {
		if err == repo.ErrNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: %w", err)
	}

	ok, err := s.passwordHasher.Verify(password, admin.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}

	if admin.IsTOTPEnabled && admin.TOTPConfirmedAt != nil {
		if totpCode == "" {
			return &LoginResult{RequiresTOTP: true}, nil
		}
		if !s.totp.Validate(totpCode, admin.TOTPSecret) {
			return nil, ErrInvalidTOTP
		}
	}

	token, err := s.tokenManager.Create(admin.ID)
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &LoginResult{
		Token:   token,
		AdminID: admin.ID,
	}, nil
}

func (s *Service) SetupTOTP(ctx context.Context, adminID int64) (string, error) {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return "", err
	}

	key, err := s.totp.GenerateSecret(admin.Username)
	if err != nil {
		return "", fmt.Errorf("generate totp secret: %w", err)
	}

	admin.TOTPSecret = key.Secret
	admin.IsTOTPEnabled = false
	admin.TOTPConfirmedAt = nil

	if err := s.admins.Update(ctx, admin); err != nil {
		return "", fmt.Errorf("save totp secret: %w", err)
	}

	return key.URL, nil
}

func (s *Service) ConfirmTOTP(ctx context.Context, adminID int64, code string) ([]string, error) {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	if admin.TOTPSecret == "" {
		return nil, ErrTOTPNotSetup
	}

	if !s.totp.Validate(code, admin.TOTPSecret) {
		return nil, ErrInvalidTOTP
	}

	now := timeNow()
	admin.IsTOTPEnabled = true
	admin.TOTPConfirmedAt = &now

	if err := s.admins.Update(ctx, admin); err != nil {
		return nil, fmt.Errorf("confirm totp: %w", err)
	}

	if err := s.recoveryCodes.DeleteByAdminID(ctx, admin.ID); err != nil {
		return nil, fmt.Errorf("delete old recovery codes: %w", err)
	}

	return s.generateRecoveryCodes(ctx, admin.ID)
}

func (s *Service) VerifyRecoveryCode(ctx context.Context, adminID int64, code string) (string, error) {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return "", err
	}

	codeHash, err := s.passwordHasher.Hash(code)
	if err != nil {
		return "", fmt.Errorf("hash recovery code: %w", err)
	}

	rc, err := s.recoveryCodes.GetByAdminIDAndHash(ctx, admin.ID, codeHash)
	if err != nil {
		if err == repo.ErrNotFound {
			return "", ErrInvalidRecoveryCode
		}
		return "", err
	}

	if err := s.recoveryCodes.MarkUsed(ctx, rc.ID); err != nil {
		return "", fmt.Errorf("mark recovery code used: %w", err)
	}

	token, err := s.tokenManager.Create(admin.ID)
	if err != nil {
		return "", fmt.Errorf("create token: %w", err)
	}

	return token, nil
}

func (s *Service) ChangePassword(ctx context.Context, adminID int64, currentPassword, newPassword string) error {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	ok, err := s.passwordHasher.Verify(currentPassword, admin.PasswordHash)
	if err != nil {
		return fmt.Errorf("verify current password: %w", err)
	}
	if !ok {
		return ErrInvalidCredentials
	}

	hash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	admin.PasswordHash = hash
	return s.admins.Update(ctx, admin)
}

func (s *Service) SeedAdmin(ctx context.Context, username, password string) error {
	count, err := s.admins.Count(ctx)
	if err != nil {
		return fmt.Errorf("count admins: %w", err)
	}
	if count > 0 {
		return nil
	}

	hash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := &domain.Admin{
		Username:     username,
		PasswordHash: hash,
	}

	return s.admins.Create(ctx, admin)
}

func (s *Service) GetAdmin(ctx context.Context, adminID int64) (*domain.Admin, error) {
	return s.admins.GetByID(ctx, adminID)
}

func (s *Service) LoginRecovery(ctx context.Context, username, code string) (string, error) {
	admin, err := s.admins.GetByUsername(ctx, username)
	if err != nil {
		if err == repo.ErrNotFound {
			return "", ErrInvalidRecoveryCode
		}
		return "", err
	}

	return s.VerifyRecoveryCode(ctx, admin.ID, code)
}

func (s *Service) DisableTOTP(ctx context.Context, adminID int64, code string) error {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	if !s.totp.Validate(code, admin.TOTPSecret) {
		if _, err := s.VerifyRecoveryCode(ctx, adminID, code); err != nil {
			return ErrInvalidTOTP
		}
	}

	admin.TOTPSecret = ""
	admin.IsTOTPEnabled = false
	admin.TOTPConfirmedAt = nil

	if err := s.admins.Update(ctx, admin); err != nil {
		return fmt.Errorf("disable totp: %w", err)
	}

	if err := s.recoveryCodes.DeleteByAdminID(ctx, adminID); err != nil {
		return fmt.Errorf("delete recovery codes: %w", err)
	}

	return nil
}

func (s *Service) generateRecoveryCodes(ctx context.Context, adminID int64) ([]string, error) {
	codes := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		plain, err := s.generateCode()
		if err != nil {
			return nil, fmt.Errorf("generate recovery code: %w", err)
		}

		hash, err := s.passwordHasher.Hash(plain)
		if err != nil {
			return nil, fmt.Errorf("hash recovery code: %w", err)
		}

		rc := &domain.RecoveryCode{
			AdminID:  adminID,
			CodeHash: hash,
		}
		if err := s.recoveryCodes.Create(ctx, rc); err != nil {
			return nil, fmt.Errorf("store recovery code: %w", err)
		}

		codes = append(codes, plain)
	}
	return codes, nil
}
