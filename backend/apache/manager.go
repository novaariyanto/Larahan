package apache

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

const defaultVersion = "2.4"

// Manager owns Apache install/lifecycle state.
type Manager struct {
	installPath string
	pipeline    *installer.Pipeline
	store       *storage.Store
	settings    *config.Manager
	beforeStart func() error
	mu          sync.Mutex
	busy        bool
}

// NewManager creates an Apache manager.
func NewManager(installPath string, pipeline *installer.Pipeline, store *storage.Store, settings *config.Manager) *Manager {
	return &Manager{
		installPath: installPath,
		pipeline:    pipeline,
		store:       store,
		settings:    settings,
	}
}

// Info returns Apache status.
func (m *Manager) Info() models.ComponentInfo {
	path := m.root()
	if !dirHasApache(path) {
		return models.ComponentInfo{
			Type:      models.TypeApache,
			Status:    models.StatusNotInstalled,
			Installed: false,
		}
	}

	status := models.StatusStopped
	if process.IsRunning(httpdImage) {
		status = models.StatusRunning
	}

	version := DetectVersion(m.httpdExe())
	return models.ComponentInfo{
		Type:      models.TypeApache,
		Version:   version,
		Status:    status,
		Path:      path,
		Installed: true,
	}
}

// StartInstall downloads, extracts, then auto-configures Apache.
func (m *Manager) StartInstall() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy {
		return fmt.Errorf("instalasi Apache sedang berjalan")
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

		_ = m.pipeline.Install(context.Background(), models.TypeApache, defaultVersion, func(dest string) error {
			return Configure(dest, m.port())
		})
	}()
	return nil
}

// Uninstall stops Apache then removes files and metadata.
func (m *Manager) Uninstall() error {
	_ = m.Stop()
	path := m.root()
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	_ = os.MkdirAll(path, 0o755)
	if m.store != nil {
		_ = m.store.DeletePackage(string(models.TypeApache), defaultVersion)
	}
	return nil
}

func dirHasApache(path string) bool {
	candidates := []string{
		filepath.Join(path, "bin", "httpd.exe"),
		filepath.Join(path, "httpd.exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return true
		}
	}
	return false
}
