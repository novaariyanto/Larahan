package storage

import (
	"fmt"
	"time"
)

// DownloadRecord is a row in download_history.
type DownloadRecord struct {
	ID           int64  `json:"id"`
	Type         string `json:"type"`
	Version      string `json:"version"`
	Filename     string `json:"filename"`
	Checksum     string `json:"checksum"`
	SourceURL    string `json:"source_url"`
	LocalPath    string `json:"local_path"`
	Status       string `json:"status"`
	DownloadedAt string `json:"downloaded_at"`
}

// AddDownloadHistory inserts a download history row.
func (s *Store) AddDownloadHistory(rec DownloadRecord) error {
	if rec.DownloadedAt == "" {
		rec.DownloadedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := s.db.Exec(`
INSERT INTO download_history
    (type, version, filename, checksum, source_url, local_path, status, downloaded_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, rec.Type, rec.Version, rec.Filename, rec.Checksum, rec.SourceURL, rec.LocalPath, rec.Status, rec.DownloadedAt)
	if err != nil {
		return fmt.Errorf("add download history: %w", err)
	}
	return nil
}

// ListDownloadHistory returns recent download rows (newest first).
func (s *Store) ListDownloadHistory(limit int) ([]DownloadRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
SELECT id, type, version, filename, COALESCE(checksum, ''), COALESCE(source_url, ''),
       local_path, status, downloaded_at
FROM download_history
ORDER BY id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list download history: %w", err)
	}
	defer rows.Close()

	out := make([]DownloadRecord, 0)
	for rows.Next() {
		var r DownloadRecord
		if err := rows.Scan(
			&r.ID, &r.Type, &r.Version, &r.Filename, &r.Checksum,
			&r.SourceURL, &r.LocalPath, &r.Status, &r.DownloadedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
