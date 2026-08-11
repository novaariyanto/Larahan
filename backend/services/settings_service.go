package services

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/larahan/larahan/backend/apache"
	"github.com/larahan/larahan/backend/config"
	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/mysql"
	"github.com/larahan/larahan/backend/phpmyadmin"
	"github.com/larahan/larahan/backend/storage"
)

// SettingsService exposes application settings to the frontend.
type SettingsService struct {
	manager    *config.Manager
	store      *storage.Store
	apache     *apache.Manager
	mysql      *mysql.Manager
	phpmyadmin *phpmyadmin.Manager
}

// NewSettingsService creates the Settings Wails service.
func NewSettingsService(manager *config.Manager, store *storage.Store, apacheMgr *apache.Manager, mysqlMgr *mysql.Manager, pmaMgr *phpmyadmin.Manager) *SettingsService {
	return &SettingsService{
		manager:    manager,
		store:      store,
		apache:     apacheMgr,
		mysql:      mysqlMgr,
		phpmyadmin: pmaMgr,
	}
}

// Get returns current settings.
func (s *SettingsService) Get() models.Settings {
	return s.manager.Get()
}

// GetAppInfo returns application metadata.
func (s *SettingsService) GetAppInfo() models.AppInfo {
	return models.AppInfo{
		Name:        config.AppName,
		Version:     config.AppVersion,
		Description: config.AppDescription,
	}
}

// Save persists settings and applies port changes to installed services.
func (s *SettingsService) Save(settings models.Settings) models.Result {
	current := s.manager.Get()

	// Install path & active PHP are managed elsewhere in MVP.
	settings.InstallPath = current.InstallPath
	if settings.InstallPath == "" {
		settings.InstallPath = config.DefaultInstallPath
	}
	settings.ActivePHP = current.ActivePHP
	settings.FirstRun = false

	if settings.ApachePort <= 0 || settings.ApachePort > 65535 {
		return models.FailResult("Apache port tidak valid")
	}
	if settings.MySQLPort <= 0 || settings.MySQLPort > 65535 {
		return models.FailResult("MySQL port tidak valid")
	}

	apacheChanged := settings.ApachePort != current.ApachePort
	mysqlChanged := settings.MySQLPort != current.MySQLPort

	if err := s.manager.Save(settings); err != nil {
		return models.FailResult(err.Error())
	}
	if s.store != nil {
		if err := s.store.SyncSettings(settings); err != nil {
			return models.FailResult("settings disimpan, sync sqlite gagal: " + err.Error())
		}
	}

	notes := []string{"Settings disimpan"}
	if apacheChanged && s.apache != nil {
		if err := s.apache.ApplyPort(settings.ApachePort); err != nil {
			notes = append(notes, "Apache: "+err.Error())
		} else {
			notes = append(notes, "Apache port diterapkan")
		}
	}
	if mysqlChanged && s.mysql != nil {
		if err := s.mysql.ApplyPort(settings.MySQLPort); err != nil {
			notes = append(notes, "MySQL: "+err.Error())
		} else {
			notes = append(notes, "MySQL port diterapkan")
		}
	}
	if mysqlChanged && s.phpmyadmin != nil {
		if err := s.phpmyadmin.ApplyMySQLPort(settings.MySQLPort); err != nil {
			notes = append(notes, "phpMyAdmin: "+err.Error())
		} else if s.phpmyadmin.Info().Installed {
			notes = append(notes, "phpMyAdmin config diperbarui")
		}
	}

	ok := true
	for _, n := range notes {
		if strings.Contains(n, "Apache:") || strings.Contains(n, "MySQL:") || strings.Contains(n, "phpMyAdmin:") {
			ok = false
			break
		}
	}
	msg := strings.Join(notes, " · ")
	if !ok {
		return models.Result{OK: false, Message: msg}
	}
	return models.OKResult(msg)
}

// GetPaths returns the standard Larahan directory layout.
func (s *SettingsService) GetPaths() models.Paths {
	settings := s.manager.Get()
	return config.ResolvePaths(settings.InstallPath)
}

// OpenDirectory opens a Larahan directory in Windows Explorer.
func (s *SettingsService) OpenDirectory(kind string) models.Result {
	paths := s.GetPaths()
	var target string
	switch strings.ToLower(kind) {
	case "root":
		target = paths.Root
	case "apache":
		target = paths.Apache
	case "php":
		target = paths.PHP
	case "mysql":
		target = paths.MySQL
	case "phpmyadmin":
		target = paths.PhpMyAdmin
	case "downloads":
		target = paths.Downloads
	case "logs":
		target = paths.Logs
	case "config":
		target = paths.Config
	default:
		return models.FailResult("direktori tidak dikenali")
	}
	if err := exec.Command("explorer", target).Start(); err != nil {
		return models.FailResult(fmt.Sprintf("gagal membuka folder: %v", err))
	}
	return models.OKResult("Folder dibuka")
}
