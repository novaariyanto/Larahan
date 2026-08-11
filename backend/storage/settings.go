package storage

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/larahan/larahan/backend/models"
)

// SetSetting upserts a single settings key/value.
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(`
INSERT INTO settings (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value
`, key, value)
	if err != nil {
		return fmt.Errorf("set setting %s: %w", key, err)
	}
	return nil
}

// GetSetting returns a setting value, or empty string if missing.
func (s *Store) GetSetting(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return value, nil
}

// SyncSettings writes the JSON settings snapshot into the settings table.
func (s *Store) SyncSettings(settings models.Settings) error {
	pairs := map[string]string{
		"install_path": settings.InstallPath,
		"active_php":   settings.ActivePHP,
		"apache_port":  strconv.Itoa(settings.ApachePort),
		"mysql_port":   strconv.Itoa(settings.MySQLPort),
		"first_run":    strconv.FormatBool(settings.FirstRun),
	}
	for key, value := range pairs {
		if err := s.SetSetting(key, value); err != nil {
			return err
		}
	}
	return nil
}
