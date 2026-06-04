package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
)

type InboundRepo struct {
	db *sql.DB
}

func NewInboundRepo(db *sql.DB) *InboundRepo { return &InboundRepo{db: db} }

const inboundColumns = `id, node_id, remote_id, remote_version, remark, protocol, port, transmission, tls, sni, dest, enabled, settings_json, last_synced_at, created_at, updated_at`

func (r *InboundRepo) Create(ctx context.Context, ib *domain.Inbound) error {
	settings, err := json.Marshal(ib.Settings)
	if err != nil {
		return fmt.Errorf("marshal inbound settings: %w", err)
	}

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO inbounds (node_id, remote_id, remote_version, remark, protocol, port, transmission, tls, sni, dest, enabled, settings_json, last_synced_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		nullableInt64(ib.NodeID), ib.RemoteID, ib.RemoteVersion, ib.Remark, ib.Protocol, ib.Port, ib.Transmission,
		ib.TLS, ib.SNI, ib.Dest, ib.Enabled, string(settings), ib.LastSyncedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return repo.ErrExist
		}
		return fmt.Errorf("insert inbound: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get inbound id: %w", err)
	}
	ib.ID = id
	return nil
}

func (r *InboundRepo) GetByID(ctx context.Context, id int64) (*domain.Inbound, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+inboundColumns+` FROM inbounds WHERE id = ?`, id)
	return scanInbound(row)
}

func (r *InboundRepo) GetByRemote(ctx context.Context, nodeID int64, remoteID string) (*domain.Inbound, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+inboundColumns+` FROM inbounds WHERE node_id = ? AND remote_id = ?`, nodeID, remoteID)
	return scanInbound(row)
}

func (r *InboundRepo) List(ctx context.Context) ([]domain.Inbound, error) {
	return r.query(ctx, `SELECT `+inboundColumns+` FROM inbounds ORDER BY id DESC`)
}

func (r *InboundRepo) ListEnabled(ctx context.Context) ([]domain.Inbound, error) {
	return r.query(ctx, `SELECT `+inboundColumns+` FROM inbounds WHERE enabled = 1 AND node_id IS NULL ORDER BY id`)
}

func (r *InboundRepo) query(ctx context.Context, q string, args ...any) ([]domain.Inbound, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query inbounds: %w", err)
	}
	defer rows.Close()

	var out []domain.Inbound
	for rows.Next() {
		ib, err := scanInbound(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *ib)
	}
	return out, rows.Err()
}

func (r *InboundRepo) Update(ctx context.Context, ib *domain.Inbound) error {
	settings, err := json.Marshal(ib.Settings)
	if err != nil {
		return fmt.Errorf("marshal inbound settings: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE inbounds SET node_id = ?, remote_id = ?, remote_version = ?,
		                     remark = ?, protocol = ?, port = ?, transmission = ?, tls = ?,
		                     sni = ?, dest = ?, enabled = ?, settings_json = ?,
		                     last_synced_at = ?,
		                     updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		nullableInt64(ib.NodeID), ib.RemoteID, ib.RemoteVersion, ib.Remark, ib.Protocol, ib.Port,
		ib.Transmission, ib.TLS, ib.SNI, ib.Dest, ib.Enabled, string(settings), ib.LastSyncedAt, ib.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return repo.ErrExist
		}
		return fmt.Errorf("update inbound: %w", err)
	}
	return nil
}

func (r *InboundRepo) UpsertRemote(ctx context.Context, nodeID int64, remoteID string, ib *domain.Inbound) (int64, error) {
	if remoteID == "" {
		return 0, fmt.Errorf("remote inbound id is required")
	}
	settings, err := json.Marshal(ib.Settings)
	if err != nil {
		return 0, fmt.Errorf("marshal inbound settings: %w", err)
	}
	now := "CURRENT_TIMESTAMP"
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO inbounds (node_id, remote_id, remote_version, remark, protocol, port, transmission, tls, sni, dest, enabled, settings_json, last_synced_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+now+`)
		 ON CONFLICT(node_id, remote_id) WHERE node_id IS NOT NULL AND remote_id <> ''
		 DO UPDATE SET remote_version = excluded.remote_version,
		               remark = excluded.remark,
		               protocol = excluded.protocol,
		               port = excluded.port,
		               transmission = excluded.transmission,
		               tls = excluded.tls,
		               sni = excluded.sni,
		               dest = excluded.dest,
		               enabled = excluded.enabled,
		               settings_json = excluded.settings_json,
		               last_synced_at = CURRENT_TIMESTAMP,
		               updated_at = CURRENT_TIMESTAMP`,
		nodeID, remoteID, ib.RemoteVersion, ib.Remark, ib.Protocol, ib.Port, ib.Transmission,
		ib.TLS, ib.SNI, ib.Dest, ib.Enabled, string(settings),
	)
	if err != nil {
		return 0, fmt.Errorf("upsert remote inbound: %w", err)
	}
	row := r.db.QueryRowContext(ctx, `SELECT id FROM inbounds WHERE node_id = ? AND remote_id = ?`, nodeID, remoteID)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("read remote inbound id: %w", err)
	}
	return id, nil
}

func (r *InboundRepo) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE inbounds SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, enabled, id)
	if err != nil {
		return fmt.Errorf("set inbound enabled: %w", err)
	}
	return nil
}

func (r *InboundRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM inbounds WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete inbound: %w", err)
	}
	return nil
}

func scanInbound(s rowScanner) (*domain.Inbound, error) {
	var ib domain.Inbound
	var settingsJSON string
	var nodeID sql.NullInt64
	var lastSyncedAt sql.NullTime
	err := s.Scan(
		&ib.ID, &nodeID, &ib.RemoteID, &ib.RemoteVersion, &ib.Remark, &ib.Protocol, &ib.Port,
		&ib.Transmission, &ib.TLS, &ib.SNI, &ib.Dest, &ib.Enabled, &settingsJSON,
		&lastSyncedAt, &ib.CreatedAt, &ib.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("scan inbound: %w", err)
	}
	if settingsJSON != "" {
		if err := json.Unmarshal([]byte(settingsJSON), &ib.Settings); err != nil {
			return nil, fmt.Errorf("unmarshal inbound settings: %w", err)
		}
	}
	ib.NodeID = ptrFromNullInt64(nodeID)
	ib.LastSyncedAt = ptrFromNullTime(lastSyncedAt)
	return &ib, nil
}
