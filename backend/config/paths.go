package config

import (
	"path/filepath"

	"github.com/larahan/larahan/backend/models"
)

const DefaultInstallPath = `C:\Larahan`

// ResolvePaths builds the standard directory layout under installPath.
func ResolvePaths(installPath string) models.Paths {
	if installPath == "" {
		installPath = DefaultInstallPath
	}
	return models.Paths{
		Root:       installPath,
		Apache:     filepath.Join(installPath, "apache"),
		PHP:        filepath.Join(installPath, "php"),
		MySQL:      filepath.Join(installPath, "mysql"),
		PhpMyAdmin: filepath.Join(installPath, "phpmyadmin"),
		Downloads:  filepath.Join(installPath, "downloads"),
		Logs:       filepath.Join(installPath, "logs"),
		Config:     filepath.Join(installPath, "config"),
		Temp:       filepath.Join(installPath, "temp"),
	}
}
