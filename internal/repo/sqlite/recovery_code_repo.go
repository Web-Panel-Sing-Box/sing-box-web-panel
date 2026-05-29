package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
)

type RecoveryCodeRepo struct {
	db *sql.DB
}

func NewRecoveryCodeRepo(db *sql.DB) *RecoveryCodeRepo {
	return &RecoveryCodeRepo{db: db}
}

func (r *RecoveryCodeRepo) Create(ctx context.Context, code *domain.RecoveryCode) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO admin_recovery_codes (admin_id, code_hash)
		 VALUES (?, ?)`,
		code.AdminID, code.CodeHash,
	)
	if err != nil {
		return fmt.Errorf("create recovery code: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	code.ID = id
	return nil
}

func (r *RecoveryCodeRepo) FindByAdminID(ctx context.Context, adminID int64) ([]domain.RecoveryCode, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, admin_id, code_hash, is_used, used_at, created_at
		 FROM admin_recovery_codes
		 WHERE admin_id = ?
		 ORDER BY created_at DESC`, adminID,
	)
	if err != nil {
		return nil, fmt.Errorf("find recovery codes: %w", err)
	}
	defer rows.Close()

	var codes []domain.RecoveryCode
	for rows.Next() {
		var c domain.RecoveryCode
		var usedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.AdminID, &c.CodeHash, &c.IsUsed, &usedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan recovery code: %w", err)
		}
		if usedAt.Valid {
			c.UsedAt = &usedAt.Time
		}
		codes = append(codes, c)
	}

	return codes, rows.Err()
}

func (r *RecoveryCodeRepo) FindUnusedByAdminID(ctx context.Context, adminID int64) ([]domain.RecoveryCode, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, admin_id, code_hash, is_used, used_at, created_at
		 FROM admin_recovery_codes
		 WHERE admin_id = ? AND is_used = 0
		 ORDER BY created_at DESC`, adminID,
	)
	if err != nil {
		return nil, fmt.Errorf("find unused recovery codes: %w", err)
	}
	defer rows.Close()

	var codes []domain.RecoveryCode
	for rows.Next() {
		var c domain.RecoveryCode
		var usedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.AdminID, &c.CodeHash, &c.IsUsed, &usedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan recovery code: %w", err)
		}
		if usedAt.Valid {
			c.UsedAt = &usedAt.Time
		}
		codes = append(codes, c)
	}

	return codes, rows.Err()
}

func (r *RecoveryCodeRepo) GetByAdminIDAndHash(ctx context.Context, adminID int64, codeHash string) (*domain.RecoveryCode, error) {
	c := &domain.RecoveryCode{}
	var usedAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, admin_id, code_hash, is_used, used_at, created_at
		 FROM admin_recovery_codes
		 WHERE admin_id = ? AND code_hash = ? AND is_used = 0`, adminID, codeHash,
	).Scan(&c.ID, &c.AdminID, &c.CodeHash, &c.IsUsed, &usedAt, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get recovery code: %w", err)
	}
	if usedAt.Valid {
		c.UsedAt = &usedAt.Time
	}

	return c, nil
}

func (r *RecoveryCodeRepo) MarkUsed(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE admin_recovery_codes
		 SET is_used = 1, used_at = ?
		 WHERE id = ? AND is_used = 0`, now, id,
	)
	if err != nil {
		return fmt.Errorf("mark recovery code used: %w", err)
	}
	return nil
}

func (r *RecoveryCodeRepo) DeleteByAdminID(ctx context.Context, adminID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM admin_recovery_codes WHERE admin_id = ?`, adminID,
	)
	if err != nil {
		return fmt.Errorf("delete recovery codes: %w", err)
	}
	return nil
}
