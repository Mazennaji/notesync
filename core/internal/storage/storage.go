package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Store struct {
	DB *sql.DB
}

type Note struct {
	ID           int64
	LocalPath    string
	Title        string
	NotionPageID string
}

type State struct {
	LocalHash      string
	RemoteHash     string
	LastSyncedHash string
}

func (s *Store) UnlinkedNotes() ([]Note, error) {
	rows, err := s.DB.Query(
		`SELECT id, local_path, title FROM note
		 WHERE notion_page_id IS NULL OR notion_page_id = ''`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.LocalPath, &n.Title); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *Store) SetNotionPageID(noteID int64, pageID string) error {
	_, err := s.DB.Exec(
		`UPDATE note SET notion_page_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		pageID, noteID,
	)
	return err
}

func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	s := &Store{DB: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	if _, err := s.DB.Exec(mustRead("migrations/001_init.sql")); err != nil {
		return fmt.Errorf("apply 001: %w", err)
	}
	if !s.columnExists("note", "deleted") {
		if _, err := s.DB.Exec(mustRead("migrations/002_deletions.sql")); err != nil {
			return fmt.Errorf("apply 002: %w", err)
		}
	}
	return nil
}

func (s *Store) columnExists(table, col string) bool {
	rows, err := s.DB.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false
		}
		if name == col {
			return true
		}
	}
	return false
}

func mustRead(path string) string {
	b, _ := migrationsFS.ReadFile(path)
	return string(b)
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func (s *Store) UpsertNotes(paths []string) (int, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO note (local_path, title) VALUES (?, ?)
		 ON CONFLICT(local_path) DO NOTHING`,
	)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	added := 0
	for _, p := range paths {
		res, err := stmt.Exec(p, titleFromPath(p))
		if err != nil {
			return 0, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			added++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return added, nil
}

func (s *Store) CountNotes() (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM note`).Scan(&n)
	return n, err
}

func titleFromPath(p string) string {
	base := path.Base(p)
	return strings.TrimSuffix(base, path.Ext(base))
}

func (s *Store) LinkNotionPage(title, pageID string) (bool, error) {
	res, err := s.DB.Exec(
		`UPDATE note SET notion_page_id = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE title = ? AND (notion_page_id IS NULL OR notion_page_id = '')`,
		pageID, title,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *Store) CountLinkedNotes() (int, error) {
	var n int
	err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM note WHERE notion_page_id IS NOT NULL AND notion_page_id != ''`,
	).Scan(&n)
	return n, err
}

func (s *Store) LinkedNotes() ([]Note, error) {
	rows, err := s.DB.Query(
		`SELECT id, local_path, title, notion_page_id FROM note
		 WHERE notion_page_id IS NOT NULL AND notion_page_id != ''`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.LocalPath, &n.Title, &n.NotionPageID); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *Store) LastSyncedHash(noteID int64) (string, error) {
	var h string
	err := s.DB.QueryRow(
		`SELECT COALESCE(last_synced_hash, '') FROM sync_state WHERE note_id = ?`,
		noteID,
	).Scan(&h)
	if err != nil {
		return "", nil
	}
	return h, nil
}

func (s *Store) RecordSync(noteID int64, localHash, remoteHash, syncedHash, status string) error {
	_, err := s.DB.Exec(
		`INSERT INTO sync_state
		   (note_id, local_hash, remote_hash, last_synced_hash, sync_status, last_synced_at)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(note_id) DO UPDATE SET
		   local_hash       = excluded.local_hash,
		   remote_hash      = excluded.remote_hash,
		   last_synced_hash = excluded.last_synced_hash,
		   sync_status      = excluded.sync_status,
		   last_synced_at   = CURRENT_TIMESTAMP`,
		noteID, localHash, remoteHash, syncedHash, status,
	)
	return err
}

func (s *Store) LogHistory(noteID int64, operation, direction, status, localHash, remoteHash, errMsg string) error {
	_, err := s.DB.Exec(
		`INSERT INTO sync_history
		   (note_id, operation, direction, status, local_hash, remote_hash, error_message)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		noteID, operation, direction, status, localHash, remoteHash, nullify(errMsg),
	)
	return err
}

func nullify(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *Store) SyncState(noteID int64) (State, error) {
	var st State
	err := s.DB.QueryRow(
		`SELECT COALESCE(local_hash,''), COALESCE(remote_hash,''), COALESCE(last_synced_hash,'')
		 FROM sync_state WHERE note_id = ?`,
		noteID,
	).Scan(&st.LocalHash, &st.RemoteHash, &st.LastSyncedHash)
	if err != nil {
		return State{}, nil
	}
	return st, nil
}

func (s *Store) ResetSyncState() error {
	_, err := s.DB.Exec(`DELETE FROM sync_state`)
	return err
}

func (s *Store) RecordConflict(noteID int64, localHash, remoteHash string) error {
	_, err := s.DB.Exec(
		`INSERT INTO conflict (note_id, local_hash, remote_hash, status)
		 VALUES (?, ?, ?, 'unresolved')`,
		noteID, localHash, remoteHash,
	)
	return err
}
