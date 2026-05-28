CREATE TABLE IF NOT EXISTS traffic_ledger (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id        INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    upload_bytes   INTEGER  NOT NULL DEFAULT 0,
    download_bytes INTEGER  NOT NULL DEFAULT 0,
    recorded_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_traffic_user_time ON traffic_ledger(user_id, recorded_at);
