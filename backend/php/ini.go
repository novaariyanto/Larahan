package php

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsureIni creates php.ini from the development template when missing,
// and applies minimal defaults useful for local development.
func EnsureIni(phpDir string) error {
	iniPath := filepath.Join(phpDir, "php.ini")
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		src := filepath.Join(phpDir, "php.ini-development")
		if _, err := os.Stat(src); err != nil {
			src = filepath.Join(phpDir, "php.ini-production")
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("php.ini template tidak ditemukan di %s", phpDir)
		}
		if err := os.WriteFile(iniPath, data, 0o644); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(iniPath)
	if err != nil {
		return err
	}
	content := string(data)
	extDir := filepath.Join(phpDir, "ext")
	extSlash := filepath.ToSlash(extDir)

	content = setIniValue(content, "extension_dir", `"`+extSlash+`"`)
	content = setIniValue(content, "date.timezone", `"UTC"`)

	for _, ext := range []string{"curl", "fileinfo", "mbstring", "mysqli", "openssl", "pdo_mysql"} {
		content = enableExtension(content, ext)
	}
	// GD extension DLL name differs across PHP majors.
	if strings.HasPrefix(filepath.Base(phpDir), "7.") || strings.Contains(phpDir, `\7.`) || strings.Contains(phpDir, `/7.`) {
		content = enableExtension(content, "gd2")
	} else {
		content = enableExtension(content, "gd")
	}

	return os.WriteFile(iniPath, []byte(content), 0o644)
}

func setIniValue(content, key, value string) string {
	prefix := key + " ="
	lines := strings.Split(content, "\n")
	found := false
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		naked := strings.TrimPrefix(trim, ";")
		naked = strings.TrimSpace(naked)
		if strings.HasPrefix(naked, key+" ") || strings.HasPrefix(naked, key+"=") {
			lines[i] = prefix + " " + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, prefix+" "+value)
	}
	return strings.Join(lines, "\n")
}

func enableExtension(content, name string) string {
	targets := []string{
		"extension=" + name,
		"extension=php_" + name + ".dll",
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		naked := strings.TrimPrefix(trim, ";")
		naked = strings.TrimSpace(naked)
		for _, t := range targets {
			if strings.EqualFold(naked, t) || strings.EqualFold(strings.ReplaceAll(naked, " ", ""), strings.ReplaceAll(t, " ", "")) {
				lines[i] = t
				return strings.Join(lines, "\n")
			}
		}
	}
	return strings.Join(append(lines, "extension="+name), "\n")
}
