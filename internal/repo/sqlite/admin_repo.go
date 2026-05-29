package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
)

type AdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO admins (username, password_hash, totp_secret, is_totp_enabled, totp_confirmed_at)
		 VALUES (?, ?, ?, ?, ?)`,
		admin.Username, admin.PasswordHash,
		admin.TOTPSecret, admin.IsTOTPEnabled, admin.TOTPConfirmedAt,
	)
	if err != nil {
		return fmt.Errorf("insert admin: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get admin id: %w", err)
	}
	admin.ID = id
	return nil
}

func (r *AdminRepo) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, totp_secret, is_totp_enabled,
		        totp_confirmed_at, created_at, updated_at
		 FROM admins WHERE id = ?`, id,
	)
	return scanAdmin(row)
}

func (r *AdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, totp_secret, is_totp_enabled,
		        totp_confirmed_at, created_at, updated_at
		 FROM admins WHERE username = ?`, username,
	)
	return scanAdmin(row)
}

func (r *AdminRepo) Update(ctx context.Context, admin *domain.Admin) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE admins SET password_hash = ?,
		                  totp_secret = ?,
		                  is_totp_enabled = ?,
		                  totp_confirmed_at = ?,
		                  updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		admin.PasswordHash, admin.TOTPSecret,
		admin.IsTOTPEnabled, admin.TOTPConfirmedAt, admin.ID,
	)
	if err != nil {
		return fmt.Errorf("update admin: %w", err)
	}
	return nil
}

func (r *AdminRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admins`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return count, nil
}

func scanAdmin(row *sql.Row) (*domain.Admin, error) {
	var a domain.Admin
	err := row.Scan(
		&a.ID, &a.Username, &a.PasswordHash,
		&a.TOTPSecret, &a.IsTOTPEnabled,
		&a.TOTPConfirmedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("scan admin: %w", err)
	}
	return &a, nil
}
