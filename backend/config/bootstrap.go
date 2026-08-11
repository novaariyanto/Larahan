package config

import (
	"fmt"
	"os"

	"github.com/larahan/larahan/backend/models"
)

// EnsureDirectories creates the standard Larahan folder layout if missing.
func EnsureDirectories(paths models.Paths) error {
	dirs := []string{
		paths.Root,
		paths.Apache,
		paths.PHP,
		paths.MySQL,
		paths.PhpMyAdmin,
		paths.Downloads,
		paths.Logs,
		paths.Config,
		paths.Temp,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	return nil
}
