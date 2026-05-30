CREATE TABLE IF NOT EXISTS clients (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    inbound_id            INTEGER  NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    name                  TEXT     NOT NULL UNIQUE,
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
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_clients_inbound ON clients (inbound_id);
CREATE INDEX IF NOT EXISTS idx_clients_sub_token ON clients (sub_token);
