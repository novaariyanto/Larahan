package downloader

import (
	"fmt"
	"os"
	"path/filepath"
)

// Cache resolves local package files under the downloads directory.
type Cache struct {
	dir string
}

// NewCache creates a cache rooted at downloadsDir.
func NewCache(downloadsDir string) *Cache {
	return &Cache{dir: downloadsDir}
}

// Path returns the expected local path for a package filename.
func (c *Cache) Path(filename string) string {
	return filepath.Join(c.dir, filename)
}

// HasValid reports whether a cached file exists and (optionally) matches sizeHint.
func (c *Cache) HasValid(filename string, sizeHint int64) (string, bool, error) {
	path := c.Path(filename)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, false, nil
		}
		return path, false, err
	}
	if info.IsDir() {
		return path, false, fmt.Errorf("cache path is a directory: %s", path)
	}
	if sizeHint > 0 && info.Size() != sizeHint {
		return path, false, nil
	}
	if info.Size() == 0 {
		return path, false, nil
	}
	return path, true, nil
}

// EnsureDir creates the downloads directory if needed.
func (c *Cache) EnsureDir() error {
	return os.MkdirAll(c.dir, 0o755)
}
