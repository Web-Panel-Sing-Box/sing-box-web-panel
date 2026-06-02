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

type APITokenRepo struct {
	db *sql.DB
}

func NewAPITokenRepo(db *sql.DB) *APITokenRepo { return &APITokenRepo{db: db} }

const apiTokenColumns = `id, name, token_hash, token_prefix, scopes, enabled, last_used_at, created_at, updated_at`

func (r *APITokenRepo) Create(ctx context.Context, t *domain.APIToken) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO api_tokens (name, token_hash, token_prefix, scopes, enabled)
		 VALUES (?, ?, ?, ?, ?)`,
		t.Name, t.TokenHash, t.TokenPrefix, t.Scopes, t.Enabled,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return repo.ErrExist
		}
		return fmt.Errorf("insert api token: %w", err)
	}
	t.ID, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get api token id: %w", err)
	}
	return nil
}

func (r *APITokenRepo) List(ctx context.Context) ([]domain.APIToken, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+apiTokenColumns+` FROM api_tokens ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("query api tokens: %w", err)
	}
	defer rows.Close()

	var out []domain.APIToken
	for rows.Next() {
		t, err := scanAPIToken(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *APITokenRepo) ListEnabled(ctx context.Context) ([]domain.APIToken, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+apiTokenColumns+` FROM api_tokens WHERE enabled = 1`)
	if err != nil {
		return nil, fmt.Errorf("query enabled api tokens: %w", err)
	}
	defer rows.Close()

	var out []domain.APIToken
	for rows.Next() {
		t, err := scanAPIToken(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *APITokenRepo) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE api_tokens SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, enabled, id)
	if err != nil {
		return fmt.Errorf("set api token enabled: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *APITokenRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM api_tokens WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete api token: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *APITokenRepo) Touch(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE api_tokens SET last_used_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, at, id)
	if err != nil {
		return fmt.Errorf("touch api token: %w", err)
	}
	return nil
}

func scanAPIToken(s rowScanner) (*domain.APIToken, error) {
	var t domain.APIToken
	var lastUsed sql.NullTime
	err := s.Scan(&t.ID, &t.Name, &t.TokenHash, &t.TokenPrefix, &t.Scopes, &t.Enabled, &lastUsed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("scan api token: %w", err)
	}
	t.LastUsedAt = ptrFromNullTime(lastUsed)
	return &t, nil
}
