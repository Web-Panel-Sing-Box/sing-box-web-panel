CREATE TABLE IF NOT EXISTS inbounds (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    remark        TEXT     NOT NULL,
    protocol      TEXT     NOT NULL,
    port          INTEGER  NOT NULL UNIQUE,
    transmission  TEXT     NOT NULL DEFAULT 'tcp',
    tls           TEXT     NOT NULL DEFAULT 'none',
    sni           TEXT     NOT NULL DEFAULT '',
    dest          TEXT     NOT NULL DEFAULT '',
    enabled       BOOLEAN  NOT NULL DEFAULT 1,
    settings_json TEXT     NOT NULL DEFAULT '{}',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
