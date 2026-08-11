package phpmyadmin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/larahan/larahan/backend/config"
	"github.com/larahan/larahan/backend/installer"
	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/storage"
)

const defaultVersion = "5.2"

// ApacheBridge applies phpMyAdmin Alias without importing the apache package.
type ApacheBridge interface {
	IsApacheInstalled() bool
	WritePhpMyAdminAlias(pmaPath string) error
	RemovePhpMyAdminAlias() error
	RestartIfRunning() error
}

// Manager owns phpMyAdmin install/config state.
type Manager struct {
	installPath string
	pipeline    *installer.Pipeline
	store       *storage.Store
	settings    *config.Manager
	apache      ApacheBridge
	mu          sync.Mutex
	busy        bool
}

// NewManager creates a phpMyAdmin manager.
func NewManager(installPath string, pipeline *installer.Pipeline, store *storage.Store, settings *config.Manager) *Manager {
	return &Manager{
		installPath: installPath,
		pipeline:    pipeline,
		store:       store,
		settings:    settings,
	}
}

// SetApacheBridge wires Apache Alias integration.
func (m *Manager) SetApacheBridge(bridge ApacheBridge) {
	m.apache = bridge
}

// Info returns phpMyAdmin status.
func (m *Manager) Info() models.ComponentInfo {
	path := m.root()
	if !dirHasPhpMyAdmin(path) {
		return models.ComponentInfo{
			Type:      models.TypePhpMyAdmin,
			Status:    models.StatusNotInstalled,
			Installed: false,
		}
	}
	return models.ComponentInfo{
		Type:      models.TypePhpMyAdmin,
		Version:   DetectVersion(path),
		Status:    models.StatusReady,
		Path:      path,
		Installed: true,
	}
}

// URL builds the local phpMyAdmin URL using the configured Apache port.
func (m *Manager) URL() string {
	port := 80
	if m.settings != nil {
		if p := m.settings.Get().ApachePort; p > 0 {
			port = p
		}
	}
	if port == 80 {
		return "http://127.0.0.1/phpmyadmin/"
	}
	return fmt.Sprintf("http://127.0.0.1:%d/phpmyadmin/", port)
}

// StartInstall downloads, extracts, writes config.inc.php, and configures Apache Alias.
func (m *Manager) StartInstall() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy {
		return fmt.Errorf("instalasi phpMyAdmin sedang berjalan")
	}
	if m.pipeline == nil {
		return fmt.Errorf("install pipeline belum siap")
	}
	if m.apache == nil || !m.apache.IsApacheInstalled() {
		return fmt.Errorf("Apache harus terinstal sebelum phpMyAdmin")
	}
	m.busy = true

	go func() {
		defer func() {
			m.mu.Lock()
			m.busy = false
			m.mu.Unlock()
		}()

		_ = m.pipeline.Install(context.Background(), models.TypePhpMyAdmin, defaultVersion, func(dest string) error {
			if err := WriteConfig(dest, m.mysqlPort()); err != nil {
				return err
			}
			if err := m.apache.WritePhpMyAdminAlias(dest); err != nil {
				return err
			}
			return m.apache.RestartIfRunning()
		})
	}()
	return nil
}

// Uninstall removes phpMyAdmin files, Apache alias, and metadata.
func (m *Manager) Uninstall() error {
	if m.apache != nil {
		_ = m.apache.RemovePhpMyAdminAlias()
		_ = m.apache.RestartIfRunning()
	}
	path := m.root()
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	_ = os.MkdirAll(path, 0o755)
	if m.store != nil {
		_ = m.store.DeletePackage(string(models.TypePhpMyAdmin), defaultVersion)
	}
	return nil
}

// ApplyMySQLPort regenerates config.inc.php when MySQL port changes.
func (m *Manager) ApplyMySQLPort(port int) error {
	path := m.root()
	if !dirHasPhpMyAdmin(path) {
		return nil
	}
	return WriteConfig(path, port)
}

func (m *Manager) root() string {
	return filepath.Join(m.installPath, "phpmyadmin")
}

func (m *Manager) mysqlPort() int {
	if m.settings != nil {
		if p := m.settings.Get().MySQLPort; p > 0 {
			return p
		}
	}
	return 3306
}

func dirHasPhpMyAdmin(path string) bool {
	candidates := []string{
		filepath.Join(path, "index.php"),
		filepath.Join(path, "libraries", "classes", "Version.php"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return true
		}
	}
	return false
}
