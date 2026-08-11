package mysql

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Configure writes my.ini and ensures data/logs directories exist.
func Configure(mysqlRoot string, port int) error {
	if err := normalizeLayout(mysqlRoot); err != nil {
		return err
	}
	if port <= 0 {
		port = 3306
	}

	rootSlash := filepath.ToSlash(mysqlRoot)
	dataDir := filepath.Join(mysqlRoot, "data")
	tmpDir := filepath.Join(mysqlRoot, "tmp")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}

	ini := fmt.Sprintf(`[mysqld]
basedir=%s
datadir=%s
port=%d
bind-address=127.0.0.1
tmpdir=%s
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci
default-storage-engine=INNODB
max_allowed_packet=64M

[client]
port=%d
default-character-set=utf8mb4

[mysql]
default-character-set=utf8mb4
`, rootSlash, filepath.ToSlash(dataDir), port, filepath.ToSlash(tmpDir), port)

	return os.WriteFile(filepath.Join(mysqlRoot, "my.ini"), []byte(ini), 0o644)
}

// NeedsInit reports whether the data directory still needs initialization.
func NeedsInit(mysqlRoot string) bool {
	marker := filepath.Join(mysqlRoot, "data", "mysql")
	info, err := os.Stat(marker)
	return err != nil || !info.IsDir()
}

// normalizeLayout lifts nested mysql-*-winx64 folders into the install root.
func normalizeLayout(mysqlRoot string) error {
	if dirHasMySQL(mysqlRoot) {
		return nil
	}
	entries, err := os.ReadDir(mysqlRoot)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		candidate := filepath.Join(mysqlRoot, e.Name())
		if !dirHasMySQL(candidate) {
			continue
		}
		children, err := os.ReadDir(candidate)
		if err != nil {
			return err
		}
		for _, child := range children {
			from := filepath.Join(candidate, child.Name())
			to := filepath.Join(mysqlRoot, child.Name())
			if _, err := os.Stat(to); err == nil {
				continue
			}
			if err := os.Rename(from, to); err != nil {
				return err
			}
		}
		_ = os.Remove(candidate)
		break
	}
	if !dirHasMySQL(mysqlRoot) {
		return fmt.Errorf("mysqld.exe tidak ditemukan di %s", mysqlRoot)
	}
	return nil
}

// DetectVersion parses mysqld --version output.
func DetectVersion(mysqldExe string) string {
	out, err := runMysqld(mysqldExe, "--version")
	if err != nil {
		return defaultVersion
	}
	re := regexp.MustCompile(`Ver\s+(\d+\.\d+(?:\.\d+)?)`)
	m := re.FindStringSubmatch(out)
	if len(m) >= 2 {
		return m[1]
	}
	// Alternate: "mysql  Ver 8.4.11 for Win64"
	re2 := regexp.MustCompile(`(?i)mysql.*?(\d+\.\d+\.\d+)`)
	m2 := re2.FindStringSubmatch(out)
	if len(m2) >= 2 {
		return m2[1]
	}
	if strings.Contains(out, "8.4") {
		return "8.4"
	}
	return defaultVersion
}
