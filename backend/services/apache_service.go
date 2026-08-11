package services

import (
	"github.com/larahan/larahan/backend/apache"
	"github.com/larahan/larahan/backend/models"
)

// ApacheService exposes Apache management to the frontend.
type ApacheService struct {
	manager *apache.Manager
}

// NewApacheService creates the Apache Wails service.
func NewApacheService(manager *apache.Manager) *ApacheService {
	return &ApacheService{manager: manager}
}

// GetInfo returns Apache version, status, and install path.
func (s *ApacheService) GetInfo() models.ComponentInfo {
	return s.manager.Info()
}

// Install starts async Apache package install + auto configure.
func (s *ApacheService) Install() models.Result {
	if err := s.manager.StartInstall(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("Instalasi Apache dimulai")
}

// Uninstall stops and removes Apache.
func (s *ApacheService) Uninstall() models.Result {
	if err := s.manager.Uninstall(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("Apache di-uninstall")
}

// Start launches Apache.
func (s *ApacheService) Start() models.Result {
	if err := s.manager.Start(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("Apache started")
}

// Stop stops Apache.
func (s *ApacheService) Stop() models.Result {
	if err := s.manager.Stop(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("Apache stopped")
}

// Restart restarts Apache.
func (s *ApacheService) Restart() models.Result {
	if err := s.manager.Restart(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("Apache restarted")
}
