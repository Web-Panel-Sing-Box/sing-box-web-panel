package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
)

type ConfigRevisionRepo struct {
	db *sql.DB
}

func NewConfigRevisionRepo(db *sql.DB) *ConfigRevisionRepo { return &ConfigRevisionRepo{db: db} }

func (r *ConfigRevisionRepo) Create(ctx context.Context, rev *domain.ConfigRevision) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO config_revisions (sha256, ok, error) VALUES (?, ?, ?)`,
		rev.SHA256, rev.OK, rev.Error)
	if err != nil {
		return fmt.Errorf("insert config revision: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get config revision id: %w", err)
	}
	rev.ID = id
	return nil
}

func (r *ConfigRevisionRepo) Latest(ctx context.Context) (*domain.ConfigRevision, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, sha256, ok, error, applied_at FROM config_revisions ORDER BY id DESC LIMIT 1`)
	var rev domain.ConfigRevision
	err := row.Scan(&rev.ID, &rev.SHA256, &rev.OK, &rev.Error, &rev.AppliedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("scan config revision: %w", err)
	}
	return &rev, nil
}

func (r *ConfigRevisionRepo) List(ctx context.Context, limit int) ([]domain.ConfigRevision, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, sha256, ok, error, applied_at FROM config_revisions ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list config revisions: %w", err)
	}
	defer rows.Close()

	var out []domain.ConfigRevision
	for rows.Next() {
		var rev domain.ConfigRevision
		if err := rows.Scan(&rev.ID, &rev.SHA256, &rev.OK, &rev.Error, &rev.AppliedAt); err != nil {
			return nil, fmt.Errorf("scan config revision: %w", err)
		}
		out = append(out, rev)
	}
	return out, rows.Err()
}
