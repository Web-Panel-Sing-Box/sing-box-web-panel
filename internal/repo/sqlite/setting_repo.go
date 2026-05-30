package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"sing-box-web-panel/internal/repo"
)

type SettingRepo struct {
	db *sql.DB
}

func NewSettingRepo(db *sql.DB) *SettingRepo { return &SettingRepo{db: db} }

func (r *SettingRepo) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repo.ErrNotFound
		}
		return "", fmt.Errorf("get setting: %w", err)
	}
	return value, nil
}

func (r *SettingRepo) All(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		out[k] = v
	}
	return out, rows.Err()
}

func (r *SettingRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
		key, value)
	if err != nil {
		return fmt.Errorf("set setting: %w", err)
	}
	return nil
}
