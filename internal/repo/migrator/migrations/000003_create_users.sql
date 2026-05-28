CREATE TABLE IF NOT EXISTS users (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    name                  TEXT    NOT NULL,
    uuid                  TEXT    NOT NULL UNIQUE,
    inbound_id            INTEGER NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    used_down             INTEGER NOT NULL DEFAULT 0,
    used_up               INTEGER NOT NULL DEFAULT 0,
    total_quota           INTEGER NOT NULL DEFAULT 0,
    expiry                DATETIME,
    status                TEXT    NOT NULL DEFAULT 'active',
    subscription          TEXT    NOT NULL DEFAULT '',
    start_after_first_use BOOLEAN NOT NULL DEFAULT FALSE,
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_uuid       ON users(uuid);
CREATE INDEX IF NOT EXISTS idx_users_inbound_id ON users(inbound_id);
CREATE INDEX IF NOT EXISTS idx_users_status     ON users(status);
