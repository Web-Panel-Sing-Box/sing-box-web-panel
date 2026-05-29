package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
)

type RecoveryRepo struct {
	db *sql.DB
}

func NewRecoveryRepo(db *sql.DB) *RecoveryRepo {
	return &RecoveryRepo{db: db}
}

func (r *RecoveryRepo) Create(ctx context.Context, code *domain.RecoveryCode) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO admin_recovery_codes (admin_id, code_hash) VALUES (?, ?)`,
		code.AdminID, code.CodeHash,
	)
	if err != nil {
		return fmt.Errorf("insert recovery code: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get recovery code id: %w", err)
	}
	code.ID = id
	return nil
}

func (r *RecoveryRepo) FindUnusedByAdminID(ctx context.Context, adminID int64) ([]domain.RecoveryCode, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, admin_id, code_hash, is_used, used_at, created_at
		 FROM admin_recovery_codes
		 WHERE admin_id = ? AND is_used = 0
		 ORDER BY created_at`, adminID,
	)
	if err != nil {
		return nil, fmt.Errorf("query recovery codes: %w", err)
	}
	defer rows.Close()

	var codes []domain.RecoveryCode
	for rows.Next() {
		var c domain.RecoveryCode
		if err := rows.Scan(&c.ID, &c.AdminID, &c.CodeHash, &c.IsUsed, &c.UsedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan recovery code: %w", err)
		}
		codes = append(codes, c)
	}

	return codes, rows.Err()
}

func (r *RecoveryRepo) GetByAdminIDAndHash(ctx context.Context, adminID int64, codeHash string) (*domain.RecoveryCode, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, admin_id, code_hash, is_used, used_at, created_at
		 FROM admin_recovery_codes
		 WHERE admin_id = ? AND code_hash = ? AND is_used = 0`, adminID, codeHash,
	)

	var c domain.RecoveryCode
	err := row.Scan(&c.ID, &c.AdminID, &c.CodeHash, &c.IsUsed, &c.UsedAt, &c.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("scan recovery code: %w", err)
	}
	return &c, nil
}

func (r *RecoveryRepo) MarkUsed(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE admin_recovery_codes SET is_used = 1, used_at = CURRENT_TIMESTAMP WHERE id = ? AND is_used = 0`, id,
	)
	if err != nil {
		return fmt.Errorf("mark recovery code used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *RecoveryRepo) DeleteByAdminID(ctx context.Context, adminID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_recovery_codes WHERE admin_id = ?`, adminID)
	if err != nil {
		return fmt.Errorf("delete recovery codes: %w", err)
	}
	return nil
}
