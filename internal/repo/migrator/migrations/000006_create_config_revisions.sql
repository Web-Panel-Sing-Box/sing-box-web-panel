CREATE TABLE IF NOT EXISTS config_revisions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    config_json TEXT    NOT NULL,
    checksum    TEXT    NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_config_revisions_created ON config_revisions(created_at);
