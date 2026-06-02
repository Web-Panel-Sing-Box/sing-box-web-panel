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

type NodeRepo struct {
	db *sql.DB
}

func NewNodeRepo(db *sql.DB) *NodeRepo { return &NodeRepo{db: db} }

const nodeColumns = `id, name, remark, scheme, address, port, base_path, api_token_secret,
	enabled, allow_private_address, status, last_heartbeat_at, latency_ms, panel_version,
	core_version, cpu_pct, ram_pct, uptime_seconds, last_error, created_at, updated_at`

func (r *NodeRepo) Create(ctx context.Context, n *domain.Node) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO nodes (name, remark, scheme, address, port, base_path, api_token_secret,
		                    enabled, allow_private_address, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.Name, n.Remark, n.Scheme, n.Address, n.Port, n.BasePath, n.APITokenSecret,
		n.Enabled, n.AllowPrivateAddress, n.Status,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return repo.ErrExist
		}
		return fmt.Errorf("insert node: %w", err)
	}
	n.ID, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get node id: %w", err)
	}
	return nil
}

func (r *NodeRepo) GetByID(ctx context.Context, id int64) (*domain.Node, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+nodeColumns+` FROM nodes WHERE id = ?`, id)
	return scanNode(row)
}

func (r *NodeRepo) List(ctx context.Context) ([]domain.Node, error) {
	return r.query(ctx, `SELECT `+nodeColumns+` FROM nodes ORDER BY id DESC`)
}

func (r *NodeRepo) ListEnabled(ctx context.Context) ([]domain.Node, error) {
	return r.query(ctx, `SELECT `+nodeColumns+` FROM nodes WHERE enabled = 1 ORDER BY id`)
}

func (r *NodeRepo) query(ctx context.Context, q string, args ...any) ([]domain.Node, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query nodes: %w", err)
	}
	defer rows.Close()

	var out []domain.Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *n)
	}
	return out, rows.Err()
}

func (r *NodeRepo) Update(ctx context.Context, n *domain.Node) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE nodes SET name = ?, remark = ?, scheme = ?, address = ?, port = ?,
		                  base_path = ?, api_token_secret = ?, enabled = ?,
		                  allow_private_address = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		n.Name, n.Remark, n.Scheme, n.Address, n.Port, n.BasePath, n.APITokenSecret,
		n.Enabled, n.AllowPrivateAddress, n.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return repo.ErrExist
		}
		return fmt.Errorf("update node: %w", err)
	}
	nRows, _ := res.RowsAffected()
	if nRows == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *NodeRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM nodes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete node: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *NodeRepo) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE nodes SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, enabled, id)
	if err != nil {
		return fmt.Errorf("set node enabled: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *NodeRepo) SetStatus(ctx context.Context, id int64, status domain.NodeStatus, heartbeatAt *time.Time, latencyMS int64, panelVersion, coreVersion string, cpuPct, ramPct float64, uptimeSeconds int64, lastErr string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE nodes SET status = ?, last_heartbeat_at = ?, latency_ms = ?, panel_version = ?,
		                  core_version = ?, cpu_pct = ?, ram_pct = ?, uptime_seconds = ?,
		                  last_error = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		status, heartbeatAt, latencyMS, panelVersion, coreVersion, cpuPct, ramPct, uptimeSeconds, lastErr, id,
	)
	if err != nil {
		return fmt.Errorf("set node status: %w", err)
	}
	return nil
}

func scanNode(s rowScanner) (*domain.Node, error) {
	var n domain.Node
	var heartbeat sql.NullTime
	err := s.Scan(
		&n.ID, &n.Name, &n.Remark, &n.Scheme, &n.Address, &n.Port, &n.BasePath, &n.APITokenSecret,
		&n.Enabled, &n.AllowPrivateAddress, &n.Status, &heartbeat, &n.LatencyMS, &n.PanelVersion,
		&n.CoreVersion, &n.CPUPct, &n.RAMPct, &n.UptimeSeconds, &n.LastError, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("scan node: %w", err)
	}
	n.LastHeartbeatAt = ptrFromNullTime(heartbeat)
	return &n, nil
}
