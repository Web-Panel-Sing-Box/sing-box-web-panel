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

type ScheduledTaskRepo struct {
	db *sql.DB
}

func NewScheduledTaskRepo(db *sql.DB) *ScheduledTaskRepo { return &ScheduledTaskRepo{db: db} }

const scheduledTaskColumns = `id, name, cron_expr, action, params_json, enabled, last_run_at, next_run_at, created_at, updated_at`

func (r *ScheduledTaskRepo) Create(ctx context.Context, t *domain.ScheduledTask) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO scheduled_tasks (name, cron_expr, action, params_json, enabled, next_run_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.Name, t.CronExpr, t.Action, t.ParamsJSON, t.Enabled, t.NextRunAt,
	)
	if err != nil {
		return fmt.Errorf("insert scheduled task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get task id: %w", err)
	}
	t.ID = id
	return nil
}

func (r *ScheduledTaskRepo) GetByID(ctx context.Context, id int64) (*domain.ScheduledTask, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+scheduledTaskColumns+` FROM scheduled_tasks WHERE id = ?`, id)
	return scanScheduledTask(row)
}

func (r *ScheduledTaskRepo) List(ctx context.Context) ([]domain.ScheduledTask, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+scheduledTaskColumns+` FROM scheduled_tasks ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list scheduled tasks: %w", err)
	}
	defer rows.Close()

	var out []domain.ScheduledTask
	for rows.Next() {
		t, err := scanScheduledTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *ScheduledTaskRepo) ListEnabled(ctx context.Context) ([]domain.ScheduledTask, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+scheduledTaskColumns+` FROM scheduled_tasks WHERE enabled = 1 ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list enabled tasks: %w", err)
	}
	defer rows.Close()

	var out []domain.ScheduledTask
	for rows.Next() {
		t, err := scanScheduledTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *ScheduledTaskRepo) Update(ctx context.Context, t *domain.ScheduledTask) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET name = ?, cron_expr = ?, action = ?, params_json = ?,
		 enabled = ?, last_run_at = ?, next_run_at = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		t.Name, t.CronExpr, t.Action, t.ParamsJSON, t.Enabled, t.LastRunAt, t.NextRunAt, t.ID,
	)
	if err != nil {
		return fmt.Errorf("update scheduled task: %w", err)
	}
	return nil
}

func (r *ScheduledTaskRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM scheduled_tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete scheduled task: %w", err)
	}
	return nil
}

func (r *ScheduledTaskRepo) SetLastRun(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_run_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		at, id,
	)
	if err != nil {
		return fmt.Errorf("set last run: %w", err)
	}
	return nil
}

func (r *ScheduledTaskRepo) SetNextRun(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET next_run_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		at, id,
	)
	if err != nil {
		return fmt.Errorf("set next run: %w", err)
	}
	return nil
}

func scanScheduledTask(s rowScanner) (*domain.ScheduledTask, error) {
	var t domain.ScheduledTask
	err := s.Scan(&t.ID, &t.Name, &t.CronExpr, &t.Action, &t.ParamsJSON,
		&t.Enabled, &t.LastRunAt, &t.NextRunAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("scan scheduled task: %w", err)
	}
	return &t, nil
}
