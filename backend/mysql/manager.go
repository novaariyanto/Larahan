package mysql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/larahan/larahan/backend/config"
	"github.com/larahan/larahan/backend/installer"
	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/process"
	"github.com/larahan/larahan/backend/storage"
)

const defaultVersion = "8.4"

// Manager owns MySQL install/lifecycle state.
type Manager struct {
	installPath string
	pipeline    *installer.Pipeline
	store       *storage.Store
	settings    *config.Manager
	mu          sync.Mutex
	busy        bool
}

// NewManager creates a MySQL manager.
func NewManager(installPath string, pipeline *installer.Pipeline, store *storage.Store, settings *config.Manager) *Manager {
	return &Manager{
		installPath: installPath,
		pipeline:    pipeline,
		store:       store,
		settings:    settings,
	}
}

// Info returns MySQL status.
func (m *Manager) Info() models.ComponentInfo {
	path := m.root()
	if !dirHasMySQL(path) {
		return models.ComponentInfo{
			Type:      models.TypeMySQL,
			Status:    models.StatusNotInstalled,
			Installed: false,
		}
	}

	status := models.StatusStopped
	if process.IsRunning(mysqldImage) {
		status = models.StatusRunning
	}

	version := DetectVersion(m.mysqldExe())
	return models.ComponentInfo{
		Type:      models.TypeMySQL,
		Version:   version,
		Status:    status,
		Path:      path,
		Installed: true,
	}
}

// StartInstall downloads, extracts, configures MySQL, and initializes datadir.
func (m *Manager) StartInstall() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy {
		return fmt.Errorf("instalasi MySQL sedang berjalan")
	}
	if m.pipeline == nil {
		return fmt.Errorf("install pipeline belum siap")
	}
	m.busy = true

	go func() {
		defer func() {
			m.mu.Lock()
			m.busy = false
			m.mu.Unlock()
		}()

		_ = m.pipeline.Install(context.Background(), models.TypeMySQL, defaultVersion, func(dest string) error {
			if err := Configure(dest, m.port()); err != nil {
				return err
			}
			if NeedsInit(dest) {
				args := []string{
					"--defaults-file=" + filepath.Join(dest, "my.ini"),
					"--initialize-insecure",
					"--console",
				}
				out, err := process.RunCapture(filepath.Join(dest, "bin", "mysqld.exe"), args, filepath.Join(dest, "bin"))
				if err != nil {
					return fmt.Errorf("inisialisasi datadir gagal: %v (%s)", err, out)
				}
			}
			return nil
		})
	}()
	return nil
}

// Uninstall stops MySQL then removes files and metadata.
func (m *Manager) Uninstall() error {
	_ = m.Stop()
	path := m.root()
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	_ = os.MkdirAll(path, 0o755)
	if m.store != nil {
		_ = m.store.DeletePackage(string(models.TypeMySQL), defaultVersion)
	}
	return nil
}

func dirHasMySQL(path string) bool {
	candidates := []string{
		filepath.Join(path, "bin", "mysqld.exe"),
		filepath.Join(path, "mysqld.exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return true
		}
	}
	return false
}
