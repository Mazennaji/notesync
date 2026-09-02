CREATE TABLE IF NOT EXISTS configuration (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_path       TEXT NOT NULL,
    notion_parent_id TEXT,
    sync_mode        TEXT NOT NULL DEFAULT 'manual',
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS note (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    local_path     TEXT NOT NULL UNIQUE,
    title          TEXT,
    notion_page_id TEXT,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sync_state (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id          INTEGER NOT NULL UNIQUE,
    local_hash       TEXT,
    remote_hash      TEXT,
    last_synced_hash TEXT,
    sync_status      TEXT NOT NULL DEFAULT 'pending',
    local_deleted    INTEGER NOT NULL DEFAULT 0,
    remote_deleted   INTEGER NOT NULL DEFAULT 0,
    last_synced_at   DATETIME,
    FOREIGN KEY (note_id) REFERENCES note(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS conflict (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id     INTEGER NOT NULL,
    local_hash  TEXT,
    remote_hash TEXT,
    status      TEXT NOT NULL DEFAULT 'unresolved',
    resolution  TEXT,
    detected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    FOREIGN KEY (note_id) REFERENCES note(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sync_history (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id       INTEGER,
    operation     TEXT NOT NULL,
    direction     TEXT,
    status        TEXT NOT NULL,
    local_hash    TEXT,
    remote_hash   TEXT,
    error_message TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES note(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_note_notion_page ON note(notion_page_id);
CREATE INDEX IF NOT EXISTS idx_sync_state_status ON sync_state(sync_status);
CREATE INDEX IF NOT EXISTS idx_conflict_status ON conflict(status);