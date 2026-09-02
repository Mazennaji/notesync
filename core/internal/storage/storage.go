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
	sqlBytes, err := migrationsFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	if _, err := s.DB.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
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
