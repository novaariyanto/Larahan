package services

import (
	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/php"
)

// LibraryService exposes PHP extension management to the frontend.
type LibraryService struct {
	php *php.Manager
}

// NewLibraryService creates the Library Wails service.
func NewLibraryService(phpMgr *php.Manager) *LibraryService {
	return &LibraryService{php: phpMgr}
}

// GetSummary returns PHP version and extension list for the active PHP.
func (s *LibraryService) GetSummary() models.LibrarySummary {
	return s.php.LibrarySummary()
}

// Enable turns on a PHP extension in php.ini.
func (s *LibraryService) Enable(name string) models.Result {
	return s.set(name, true)
}

// Disable turns off a PHP extension in php.ini.
func (s *LibraryService) Disable(name string) models.Result {
	return s.set(name, false)
}

func (s *LibraryService) set(name string, enabled bool) models.Result {
	wasRunning := s.php.LibrarySummary().ApacheRunning
	if err := s.php.SetExtension(name, enabled); err != nil {
		return models.FailResult(err.Error())
	}
	action := "dinonaktifkan"
	if enabled {
		action = "diaktifkan"
	}
	msg := name + " " + action
	if wasRunning {
		msg += ". Apache di-restart agar perubahan berlaku."
	} else {
		msg += ". Start Apache agar perubahan berlaku di web."
	}
	return models.OKResult(msg)
}
