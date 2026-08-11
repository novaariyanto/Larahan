package services

import (
	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/php"
)

// PHPService exposes PHP version management to the frontend.
type PHPService struct {
	manager *php.Manager
}

// NewPHPService creates the PHP Wails service.
func NewPHPService(manager *php.Manager) *PHPService {
	return &PHPService{manager: manager}
}

// ListVersions returns supported PHP versions and install state.
func (s *PHPService) ListVersions() []models.PHPVersionInfo {
	return s.manager.ListVersions()
}

// GetActive returns the currently active PHP version.
func (s *PHPService) GetActive() string {
	return s.manager.Active()
}

// Install starts async download+extract for a PHP version.
func (s *PHPService) Install(version string) models.Result {
	if err := s.manager.StartInstall(version); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("Instalasi PHP " + version + " dimulai")
}

// Delete removes an installed PHP version.
func (s *PHPService) Delete(version string) models.Result {
	if err := s.manager.Delete(version); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("PHP " + version + " dihapus")
}

// Switch activates a PHP version and updates Apache automatically.
func (s *PHPService) Switch(version string) models.Result {
	if err := s.manager.Switch(version); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("PHP aktif diganti ke " + version)
}
