package php

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

// SupportedVersions is the MVP PHP version set.
var SupportedVersions = []string{"7.4", "8.1", "8.2", "8.3", "8.4", "8.5"}

// ApacheBridge applies PHP settings to Apache without importing the apache package.
type ApacheBridge interface {
	IsApacheInstalled() bool
	IsApacheRunning() bool
	ApplyPHPModule(phpPath, version string) error
	RestartIfRunning() error
}

// Manager owns PHP install/list/switch state.
type Manager struct {
	installPath string
	active      string
	pipeline    *installer.Pipeline
	store       *storage.Store
	settings    *config.Manager
	apache      ApacheBridge
	mu          sync.Mutex
	busy        bool
}

// NewManager creates a PHP manager.
func NewManager(installPath string, pipeline *installer.Pipeline, store *storage.Store, settings *config.Manager, active string) *Manager {
	return &Manager{
		installPath: installPath,
		pipeline:    pipeline,
		store:       store,
		settings:    settings,
		active:      active,
	}
}

// SetApacheBridge wires Apache integration for Switch Version.
func (m *Manager) SetApacheBridge(bridge ApacheBridge) {
	m.apache = bridge
}

// ListVersions returns supported versions with install/active flags.
func (m *Manager) ListVersions() []models.PHPVersionInfo {
	out := make([]models.PHPVersionInfo, 0, len(SupportedVersions))
	for _, v := range SupportedVersions {
		path := filepath.Join(m.installPath, "php", v)
		installed := dirHasPHP(path)
		out = append(out, models.PHPVersionInfo{
			Version:   v,
			Installed: installed,
			Active:    v == m.active && installed,
			Path:      ternary(installed, path, ""),
		})
	}
	return out
}

// Active returns the active PHP version string.
func (m *Manager) Active() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

// Info returns summary info for the active PHP.
// PHP is not a standalone daemon: Ready when installed, Running when Apache is up.
func (m *Manager) Info() models.ComponentInfo {
	m.mu.Lock()
	active := m.active
	m.mu.Unlock()

	if active == "" {
		for _, v := range SupportedVersions {
			path := filepath.Join(m.installPath, "php", v)
			if dirHasPHP(path) {
				return models.ComponentInfo{
					Type:      models.TypePHP,
					Version:   v,
					Status:    models.StatusReady,
					Path:      path,
					Installed: true,
				}
			}
		}
		return models.ComponentInfo{
			Type:      models.TypePHP,
			Status:    models.StatusNotInstalled,
			Installed: false,
		}
	}
	path := filepath.Join(m.installPath, "php", active)
	if !dirHasPHP(path) {
		return models.ComponentInfo{
			Type:      models.TypePHP,
			Version:   active,
			Status:    models.StatusNotInstalled,
			Installed: false,
		}
	}
	status := models.StatusReady
	if m.apache != nil && m.apache.IsApacheRunning() {
		status = models.StatusRunning
	}
	return models.ComponentInfo{
		Type:      models.TypePHP,
		Version:   active,
		Status:    status,
		Path:      path,
		Installed: true,
	}
}

// EnsureApacheHandler reapplies the active PHP module to Apache config.
func (m *Manager) EnsureApacheHandler() error {
	active := m.Active()
	if active == "" || m.apache == nil || !m.apache.IsApacheInstalled() {
		return nil
	}
	path := filepath.Join(m.installPath, "php", active)
	if !dirHasPHP(path) {
		return nil
	}
	if err := EnsureIni(path); err != nil {
		return err
	}
	return m.apache.ApplyPHPModule(path, active)
}

// StartInstall launches async install with php.ini preparation.
func (m *Manager) StartInstall(version string) error {
	if !isSupported(version) {
		return fmt.Errorf("versi PHP tidak didukung: %s", version)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy {
		return fmt.Errorf("instalasi PHP sedang berjalan")
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

		err := m.pipeline.Install(context.Background(), models.TypePHP, version, func(dest string) error {
			return EnsureIni(dest)
		})
		if err != nil {
			return
		}

		m.mu.Lock()
		shouldActivate := m.active == ""
		m.mu.Unlock()
		if shouldActivate {
			_ = m.Switch(version)
		}
	}()
	return nil
}

// Delete removes an installed PHP version directory and metadata.
func (m *Manager) Delete(version string) error {
	m.mu.Lock()
	if m.active == version {
		m.mu.Unlock()
		return fmt.Errorf("tidak bisa menghapus versi aktif; switch dulu")
	}
	m.mu.Unlock()

	path := filepath.Join(m.installPath, "php", version)
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	if m.store != nil {
		_ = m.store.DeletePackage(string(models.TypePHP), version)
	}
	return nil
}

// Switch activates a PHP version, updates Apache, and restarts Apache if running.
func (m *Manager) Switch(version string) error {
	if !isSupported(version) {
		return fmt.Errorf("versi PHP tidak didukung: %s", version)
	}
	path := filepath.Join(m.installPath, "php", version)
	if !dirHasPHP(path) {
		return fmt.Errorf("PHP %s belum terinstal", version)
	}
	if err := EnsureIni(path); err != nil {
		return err
	}

	m.mu.Lock()
	m.active = version
	m.mu.Unlock()

	if err := m.persistActive(version); err != nil {
		return err
	}

	if m.apache == nil || !m.apache.IsApacheInstalled() {
		return nil
	}
	if err := m.apache.ApplyPHPModule(path, version); err != nil {
		return err
	}
	return m.apache.RestartIfRunning()
}

func (m *Manager) persistActive(version string) error {
	if m.settings == nil {
		return nil
	}
	s := m.settings.Get()
	s.ActivePHP = version
	s.FirstRun = false
	if err := m.settings.Save(s); err != nil {
		return err
	}
	if m.store != nil {
		return m.store.SyncSettings(s)
	}
	return nil
}

func isSupported(version string) bool {
	for _, v := range SupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

func dirHasPHP(path string) bool {
	_, err := os.Stat(filepath.Join(path, "php.exe"))
	return err == nil
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
