CREATE TABLE IF NOT EXISTS inbounds (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    remark        TEXT    NOT NULL,
    protocol      TEXT    NOT NULL,
    port          INTEGER NOT NULL,
    transmission  TEXT    NOT NULL DEFAULT 'tcp',
    tls           TEXT    NOT NULL DEFAULT 'none',
    sni           TEXT    NOT NULL DEFAULT '',
    dest          TEXT    NOT NULL DEFAULT '',
    enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_inbounds_port ON inbounds(port);
