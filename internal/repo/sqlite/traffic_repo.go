package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

// TrafficRepo persists the daily traffic rollup used for the "today" and
// "this month" dashboard figures.
type TrafficRepo struct {
	db *sql.DB
}

func NewTrafficRepo(db *sql.DB) *TrafficRepo { return &TrafficRepo{db: db} }

// AddDaily increments the rollup counters for the given UTC day (YYYY-MM-DD).
func (r *TrafficRepo) AddDaily(ctx context.Context, day string, up, down int64) error {
	if up == 0 && down == 0 {
		return nil
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO traffic_rollup (day, up, down, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(day) DO UPDATE SET up = up + excluded.up, down = down + excluded.down,
		                                updated_at = CURRENT_TIMESTAMP`,
		day, up, down)
	if err != nil {
		return fmt.Errorf("add daily traffic: %w", err)
	}
	return nil
}

// Day returns the up/down totals for a single UTC day.
func (r *TrafficRepo) Day(ctx context.Context, day string) (up, down int64, err error) {
	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(up, 0), COALESCE(down, 0) FROM traffic_rollup WHERE day = ?`, day).Scan(&up, &down)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, fmt.Errorf("get daily traffic: %w", err)
	}
	return up, down, nil
}

// SumSince returns the up/down totals for all days >= sinceDay (inclusive).
func (r *TrafficRepo) SumSince(ctx context.Context, sinceDay string) (up, down int64, err error) {
	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(up), 0), COALESCE(SUM(down), 0) FROM traffic_rollup WHERE day >= ?`,
		sinceDay).Scan(&up, &down)
	if err != nil {
		return 0, 0, fmt.Errorf("sum traffic since: %w", err)
	}
	return up, down, nil
}
