package services

import (
	"fmt"
	"os/exec"

	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/phpmyadmin"
)

// PhpMyAdminService exposes phpMyAdmin management to the frontend.
type PhpMyAdminService struct {
	manager *phpmyadmin.Manager
}

// NewPhpMyAdminService creates the phpMyAdmin Wails service.
func NewPhpMyAdminService(manager *phpmyadmin.Manager) *PhpMyAdminService {
	return &PhpMyAdminService{manager: manager}
}

// GetInfo returns phpMyAdmin version, status, and install path.
func (s *PhpMyAdminService) GetInfo() models.ComponentInfo {
	return s.manager.Info()
}

// GetURL returns the local phpMyAdmin URL.
func (s *PhpMyAdminService) GetURL() string {
	return s.manager.URL()
}

// Install starts async phpMyAdmin package install + MySQL config + Apache alias.
func (s *PhpMyAdminService) Install() models.Result {
	if err := s.manager.StartInstall(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("Instalasi phpMyAdmin dimulai")
}

// Uninstall removes phpMyAdmin and clears the Apache alias.
func (s *PhpMyAdminService) Uninstall() models.Result {
	if err := s.manager.Uninstall(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("phpMyAdmin di-uninstall")
}

// Open opens phpMyAdmin in the default browser.
func (s *PhpMyAdminService) Open() models.Result {
	info := s.manager.Info()
	if !info.Installed {
		return models.FailResult("phpMyAdmin belum terinstal")
	}
	url := s.manager.URL()
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
		return models.FailResult(fmt.Sprintf("gagal membuka browser: %v", err))
	}
	return models.OKResult("phpMyAdmin dibuka di browser")
}
