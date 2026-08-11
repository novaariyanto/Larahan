package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/larahan/larahan/backend/models"
)

// InstalledPackage is a row in installed_packages.
type InstalledPackage struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Version     string `json:"version"`
	Path        string `json:"path"`
	Status      string `json:"status"`
	InstalledAt string `json:"installed_at"`
}

// UpsertPackage inserts or updates a package record by type+version.
func (s *Store) UpsertPackage(pkgType, version, path, status string) error {
	if status == "" {
		status = "installed"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
INSERT INTO installed_packages (type, version, path, status, installed_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(type, version) DO UPDATE SET
    path = excluded.path,
    status = excluded.status
`, pkgType, version, path, status, now)
	if err != nil {
		return fmt.Errorf("upsert package: %w", err)
	}
	return nil
}

// DeletePackage removes a package metadata row.
func (s *Store) DeletePackage(pkgType, version string) error {
	_, err := s.db.Exec(`DELETE FROM installed_packages WHERE type = ? AND version = ?`, pkgType, version)
	if err != nil {
		return fmt.Errorf("delete package: %w", err)
	}
	return nil
}

// ListPackages returns packages, optionally filtered by type (empty = all).
func (s *Store) ListPackages(pkgType string) ([]InstalledPackage, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if pkgType == "" {
		rows, err = s.db.Query(`
SELECT id, type, version, path, status, installed_at
FROM installed_packages
ORDER BY type, version`)
	} else {
		rows, err = s.db.Query(`
SELECT id, type, version, path, status, installed_at
FROM installed_packages
WHERE type = ?
ORDER BY version`, pkgType)
	}
	if err != nil {
		return nil, fmt.Errorf("list packages: %w", err)
	}
	defer rows.Close()

	out := make([]InstalledPackage, 0)
	for rows.Next() {
		var p InstalledPackage
		if err := rows.Scan(&p.ID, &p.Type, &p.Version, &p.Path, &p.Status, &p.InstalledAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetPackage returns one package or nil if not found.
func (s *Store) GetPackage(pkgType, version string) (*InstalledPackage, error) {
	row := s.db.QueryRow(`
SELECT id, type, version, path, status, installed_at
FROM installed_packages
WHERE type = ? AND version = ?`, pkgType, version)

	var p InstalledPackage
	err := row.Scan(&p.ID, &p.Type, &p.Version, &p.Path, &p.Status, &p.InstalledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get package: %w", err)
	}
	return &p, nil
}

// HasType reports whether any package of the given type is installed.
func (s *Store) HasType(pkgType models.ComponentType) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(1) FROM installed_packages WHERE type = ?`, string(pkgType)).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
