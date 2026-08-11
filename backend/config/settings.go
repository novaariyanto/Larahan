package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/larahan/larahan/backend/models"
)

const settingsFileName = "settings.json"

// Manager reads and writes settings.json under the config directory.
type Manager struct {
	mu       sync.RWMutex
	filePath string
	current  models.Settings
}

// NewManager creates a settings manager pointing at installPath/config/settings.json.
func NewManager(installPath string) *Manager {
	paths := ResolvePaths(installPath)
	return &Manager{
		filePath: filepath.Join(paths.Config, settingsFileName),
		current:  DefaultSettings(installPath),
	}
}

// DefaultSettings returns MVP defaults.
func DefaultSettings(installPath string) models.Settings {
	if installPath == "" {
		installPath = DefaultInstallPath
	}
	return models.Settings{
		InstallPath: installPath,
		ActivePHP:   "",
		ApachePort:  80,
		MySQLPort:   3306,
		FirstRun:    true,
	}
}

// Load reads settings from disk. If missing, persists defaults (first run).
func (m *Manager) Load() (models.Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := m.saveLocked(m.current); err != nil {
				return m.current, err
			}
			return m.current, nil
		}
		return m.current, err
	}

	var loaded models.Settings
	if err := json.Unmarshal(data, &loaded); err != nil {
		return m.current, err
	}
	if loaded.InstallPath == "" {
		loaded.InstallPath = DefaultInstallPath
	}
	if loaded.ApachePort == 0 {
		loaded.ApachePort = 80
	}
	if loaded.MySQLPort == 0 {
		loaded.MySQLPort = 3306
	}
	m.current = loaded
	return m.current, nil
}

// Get returns the in-memory settings snapshot.
func (m *Manager) Get() models.Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// Save persists settings to disk and updates memory.
func (m *Manager) Save(settings models.Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveLocked(settings)
}

// FilePath returns the settings.json absolute path.
func (m *Manager) FilePath() string {
	return m.filePath
}

func (m *Manager) saveLocked(settings models.Settings) error {
	if err := os.MkdirAll(filepath.Dir(m.filePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(m.filePath, data, 0o644); err != nil {
		return err
	}
	m.current = settings
	return nil
}
