package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbFileName = "larahan.db"

// Store wraps the SQLite connection used for application metadata.
type Store struct {
	db *sql.DB
}

// Open creates (if needed) and opens the SQLite database under configDir.
func Open(configDir string) (*Store, error) {
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	dbPath := filepath.Join(configDir, dbFileName)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return store, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// DB exposes the underlying *sql.DB for advanced use.
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS installed_packages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    version TEXT NOT NULL,
    path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'installed',
    installed_at TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pkg_type_version
    ON installed_packages(type, version);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS download_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    version TEXT NOT NULL,
    filename TEXT NOT NULL,
    checksum TEXT,
    source_url TEXT,
    local_path TEXT NOT NULL,
    status TEXT NOT NULL,
    downloaded_at TEXT NOT NULL
);
`
	_, err := s.db.Exec(schema)
	return err
}
