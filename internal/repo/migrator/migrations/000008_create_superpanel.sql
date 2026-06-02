CREATE TABLE IF NOT EXISTS api_tokens (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT     NOT NULL UNIQUE,
    token_hash   TEXT     NOT NULL UNIQUE,
    token_prefix TEXT     NOT NULL,
    scopes       TEXT     NOT NULL DEFAULT 'node',
    enabled      BOOLEAN  NOT NULL DEFAULT 1,
    last_used_at DATETIME,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_api_tokens_enabled ON api_tokens (enabled);

CREATE TABLE IF NOT EXISTS nodes (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    name                  TEXT     NOT NULL UNIQUE,
    remark                TEXT     NOT NULL DEFAULT '',
    scheme                TEXT     NOT NULL DEFAULT 'https',
    address               TEXT     NOT NULL,
    port                  INTEGER  NOT NULL,
    base_path             TEXT     NOT NULL DEFAULT '',
    api_token_secret      TEXT     NOT NULL DEFAULT '',
    enabled               BOOLEAN  NOT NULL DEFAULT 1,
    allow_private_address BOOLEAN  NOT NULL DEFAULT 0,
    status                TEXT     NOT NULL DEFAULT 'unknown',
    last_heartbeat_at     DATETIME,
    latency_ms            INTEGER  NOT NULL DEFAULT 0,
    panel_version         TEXT     NOT NULL DEFAULT '',
    core_version          TEXT     NOT NULL DEFAULT '',
    cpu_pct               REAL     NOT NULL DEFAULT 0,
    ram_pct               REAL     NOT NULL DEFAULT 0,
    uptime_seconds        INTEGER  NOT NULL DEFAULT 0,
    last_error            TEXT     NOT NULL DEFAULT '',
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_nodes_enabled ON nodes (enabled);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes (status);

ALTER TABLE clients RENAME TO clients_superpanel_old;
ALTER TABLE inbounds RENAME TO inbounds_superpanel_old;

CREATE TABLE inbounds (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id         INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    remote_id       TEXT     NOT NULL DEFAULT '',
    remote_version  TEXT     NOT NULL DEFAULT '',
    remark          TEXT     NOT NULL,
    protocol        TEXT     NOT NULL,
    port            INTEGER  NOT NULL,
    transmission    TEXT     NOT NULL DEFAULT 'tcp',
    tls             TEXT     NOT NULL DEFAULT 'none',
    sni             TEXT     NOT NULL DEFAULT '',
    dest            TEXT     NOT NULL DEFAULT '',
    enabled         BOOLEAN  NOT NULL DEFAULT 1,
    settings_json   TEXT     NOT NULL DEFAULT '{}',
    last_synced_at  DATETIME,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE clients (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id               INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    remote_id             TEXT     NOT NULL DEFAULT '',
    inbound_id            INTEGER  NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    name                  TEXT     NOT NULL,
    uuid                  TEXT     NOT NULL DEFAULT '',
    password              TEXT     NOT NULL DEFAULT '',
    used_up               INTEGER  NOT NULL DEFAULT 0,
    used_down             INTEGER  NOT NULL DEFAULT 0,
    total_quota           INTEGER  NOT NULL DEFAULT 0,
    expiry                DATETIME,
    status                TEXT     NOT NULL DEFAULT 'active',
    sub_token             TEXT     NOT NULL UNIQUE,
    start_after_first_use BOOLEAN  NOT NULL DEFAULT 0,
    enabled               BOOLEAN  NOT NULL DEFAULT 1,
    first_used_at         DATETIME,
    last_synced_at        DATETIME,
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO inbounds (
    id, remark, protocol, port, transmission, tls, sni, dest, enabled,
    settings_json, created_at, updated_at
)
SELECT
    id, remark, protocol, port, transmission, tls, sni, dest, enabled,
    settings_json, created_at, updated_at
FROM inbounds_superpanel_old;

INSERT INTO clients (
    id, inbound_id, name, uuid, password, used_up, used_down, total_quota,
    expiry, status, sub_token, start_after_first_use, enabled, first_used_at,
    created_at, updated_at
)
SELECT
    id, inbound_id, name, uuid, password, used_up, used_down, total_quota,
    expiry, status, sub_token, start_after_first_use, enabled, first_used_at,
    created_at, updated_at
FROM clients_superpanel_old;

DROP TABLE clients_superpanel_old;
DROP TABLE inbounds_superpanel_old;

CREATE UNIQUE INDEX IF NOT EXISTS idx_inbounds_local_port ON inbounds (port)
    WHERE node_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_inbounds_remote_id ON inbounds (node_id, remote_id)
    WHERE node_id IS NOT NULL AND remote_id <> '';
CREATE INDEX IF NOT EXISTS idx_inbounds_node ON inbounds (node_id);

CREATE INDEX IF NOT EXISTS idx_clients_inbound ON clients (inbound_id);
CREATE INDEX IF NOT EXISTS idx_clients_sub_token ON clients (sub_token);
CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_local_name ON clients (name)
    WHERE node_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_remote_name ON clients (node_id, name)
    WHERE node_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_remote_id ON clients (node_id, remote_id)
    WHERE node_id IS NOT NULL AND remote_id <> '';
CREATE INDEX IF NOT EXISTS idx_clients_node ON clients (node_id);
