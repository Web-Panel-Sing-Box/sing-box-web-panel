CREATE TABLE IF NOT EXISTS admin_recovery_codes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id   INTEGER  NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    code_hash  TEXT     NOT NULL,
    is_used    BOOLEAN  NOT NULL DEFAULT 0,
    used_at    DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_recovery_codes_admin ON admin_recovery_codes(admin_id);
