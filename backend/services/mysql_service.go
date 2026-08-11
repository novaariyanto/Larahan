package services

import (
	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/mysql"
)

// MySQLService exposes MySQL management to the frontend.
type MySQLService struct {
	manager *mysql.Manager
}

// NewMySQLService creates the MySQL Wails service.
func NewMySQLService(manager *mysql.Manager) *MySQLService {
	return &MySQLService{manager: manager}
}

// GetInfo returns MySQL version, status, and install path.
func (s *MySQLService) GetInfo() models.ComponentInfo {
	return s.manager.Info()
}

// Install starts async MySQL package install + configure + init.
func (s *MySQLService) Install() models.Result {
	if err := s.manager.StartInstall(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("Instalasi MySQL dimulai")
}

// Uninstall stops and removes MySQL.
func (s *MySQLService) Uninstall() models.Result {
	if err := s.manager.Uninstall(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("MySQL di-uninstall")
}

// Start launches MySQL.
func (s *MySQLService) Start() models.Result {
	if err := s.manager.Start(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("MySQL started")
}

// Stop stops MySQL.
func (s *MySQLService) Stop() models.Result {
	if err := s.manager.Stop(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("MySQL stopped")
}

// Restart restarts MySQL.
func (s *MySQLService) Restart() models.Result {
	if err := s.manager.Restart(); err != nil {
		return models.FailResult(err.Error())
	}
	return models.OKResult("MySQL restarted")
}
