package php

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/process"
)

var extNameRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// iniExtLine matches optional comment + extension/zend_extension = [php_]name[.dll]
var iniExtLineRe = regexp.MustCompile(`(?i)^\s*(;+)?\s*(zend_)?extension\s*=\s*"?(?:php_)?([a-z0-9_]+)(?:\.dll)?"?\s*$`)

type libraryDef struct {
	Name    string
	Display string
	DLLs    []string
	Zend    bool
}

var knownLibraries = []libraryDef{
	{Name: "curl", Display: "cURL", DLLs: []string{"php_curl.dll"}},
	{Name: "openssl", Display: "OpenSSL", DLLs: []string{"php_openssl.dll"}},
	{Name: "gd", Display: "GD", DLLs: []string{"php_gd.dll", "php_gd2.dll"}},
	{Name: "mbstring", Display: "Mbstring", DLLs: []string{"php_mbstring.dll"}},
	{Name: "pdo", Display: "PDO", DLLs: []string{"php_pdo.dll"}},
	{Name: "pdo_mysql", Display: "PDO MySQL", DLLs: []string{"php_pdo_mysql.dll"}},
	{Name: "mysqli", Display: "MySQLi", DLLs: []string{"php_mysqli.dll"}},
	{Name: "zip", Display: "ZIP", DLLs: []string{"php_zip.dll"}},
	{Name: "fileinfo", Display: "Fileinfo", DLLs: []string{"php_fileinfo.dll"}},
	{Name: "intl", Display: "Intl", DLLs: []string{"php_intl.dll"}},
	{Name: "soap", Display: "SOAP", DLLs: []string{"php_soap.dll"}},
	{Name: "exif", Display: "Exif", DLLs: []string{"php_exif.dll"}},
	{Name: "ftp", Display: "FTP", DLLs: []string{"php_ftp.dll"}},
	{Name: "bz2", Display: "Bzip2", DLLs: []string{"php_bz2.dll"}},
	{Name: "sodium", Display: "Sodium", DLLs: []string{"php_sodium.dll"}},
	{Name: "sqlite3", Display: "SQLite3", DLLs: []string{"php_sqlite3.dll"}},
	{Name: "pdo_sqlite", Display: "PDO SQLite", DLLs: []string{"php_pdo_sqlite.dll"}},
	{Name: "xsl", Display: "XSL", DLLs: []string{"php_xsl.dll"}},
	{Name: "ldap", Display: "LDAP", DLLs: []string{"php_ldap.dll"}},
	{Name: "gmp", Display: "GMP", DLLs: []string{"php_gmp.dll"}},
	{Name: "tidy", Display: "Tidy", DLLs: []string{"php_tidy.dll"}},
	{Name: "gettext", Display: "Gettext", DLLs: []string{"php_gettext.dll"}},
	{Name: "opcache", Display: "OPcache", DLLs: []string{"php_opcache.dll"}, Zend: true},
}

// LibrarySummary returns PHP libraries for the active version.
func (m *Manager) LibrarySummary() models.LibrarySummary {
	active := m.Active()
	if active == "" {
		return models.LibrarySummary{
			Message: "Tidak ada versi PHP aktif. Install dan Switch PHP terlebih dahulu.",
		}
	}
	phpDir := filepath.Join(m.installPath, "php", active)
	if !dirHasPHP(phpDir) {
		return models.LibrarySummary{
			PHPVersion: active,
			Message:    "PHP " + active + " belum terinstal.",
		}
	}

	phpVer := detectPHPVersion(phpDir)
	loaded := loadedModules(phpDir)
	iniEnabled := parseIniExtensions(filepath.Join(phpDir, "php.ini"))
	dlls := scanExtensionDLLs(phpDir)

	seen := map[string]struct{}{}
	out := make([]models.PHPExtension, 0, len(knownLibraries)+len(dlls))

	for _, def := range knownLibraries {
		seen[def.Name] = struct{}{}
		out = append(out, m.buildExtension(phpDir, def, dlls, iniEnabled, loaded))
	}

	extra := make([]string, 0)
	extraSeen := map[string]struct{}{}
	for name := range dlls {
		c := canonicalExtName(name)
		if _, ok := seen[c]; ok {
			continue
		}
		if _, ok := extraSeen[c]; ok {
			continue
		}
		extraSeen[c] = struct{}{}
		extra = append(extra, c)
	}
	sort.Strings(extra)
	for _, name := range extra {
		def := libraryDef{Name: name, Display: displayName(name), DLLs: []string{"php_" + name + ".dll"}}
		out = append(out, m.buildExtension(phpDir, def, dlls, iniEnabled, loaded))
	}

	apacheRunning := m.apache != nil && m.apache.IsApacheRunning()
	return models.LibrarySummary{
		Installed:     true,
		PHPVersion:    phpVer,
		PHPPath:       phpDir,
		ApacheRunning: apacheRunning,
		Extensions:    out,
	}
}

// SetExtension enables or disables a PHP extension in php.ini of the active version.
func (m *Manager) SetExtension(name string, enabled bool) error {
	name = strings.TrimSpace(strings.ToLower(name))
	if !extNameRe.MatchString(name) {
		return fmt.Errorf("nama extension tidak valid")
	}

	summary := m.LibrarySummary()
	var target *models.PHPExtension
	for i := range summary.Extensions {
		if strings.EqualFold(summary.Extensions[i].Name, name) {
			target = &summary.Extensions[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("extension %s tidak dikenali", name)
	}
	if target.Builtin {
		return fmt.Errorf("%s adalah extension bawaan PHP dan tidak dapat diubah", target.DisplayName)
	}
	if !target.Toggleable {
		return fmt.Errorf("%s tidak dapat diubah", target.DisplayName)
	}
	if enabled && target.Status == models.ExtNotInstalled {
		return fmt.Errorf("%s belum terinstal (file %s tidak ditemukan). Tidak dapat diaktifkan sebelum DLL tersedia", target.DisplayName, target.DLL)
	}

	m.mu.Lock()
	phpDir, err := m.safePHPDir(m.active)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	iniPath := filepath.Join(phpDir, "php.ini")
	if err := ensureIniUnderPHPDir(phpDir, iniPath); err != nil {
		m.mu.Unlock()
		return err
	}

	data, err := os.ReadFile(iniPath)
	if err != nil {
		m.mu.Unlock()
		return fmt.Errorf("tidak bisa membaca php.ini: %w", err)
	}
	if err := assertWritable(iniPath); err != nil {
		m.mu.Unlock()
		return err
	}

	updated, err := setExtensionLine(string(data), name, target.DLL, target.Zend, enabled)
	if err != nil {
		m.mu.Unlock()
		return err
	}

	if err := writeIniAtomic(iniPath, []byte(updated)); err != nil {
		m.mu.Unlock()
		return err
	}

	if err := verifyIniChange(iniPath, name, target.DLL, enabled); err != nil {
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock()

	if m.apache != nil && m.apache.IsApacheRunning() {
		if err := m.apache.RestartIfRunning(); err != nil {
			return fmt.Errorf("php.ini diperbarui, tetapi restart Apache gagal: %w", err)
		}
	}
	return nil
}

func (m *Manager) safePHPDir(version string) (string, error) {
	if !isSupported(version) {
		return "", fmt.Errorf("versi PHP tidak didukung")
	}
	phpDir := filepath.Clean(filepath.Join(m.installPath, "php", version))
	root := filepath.Clean(filepath.Join(m.installPath, "php"))
	rel, err := filepath.Rel(root, phpDir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path PHP tidak valid")
	}
	if !dirHasPHP(phpDir) {
		return "", fmt.Errorf("PHP %s belum terinstal", version)
	}
	return phpDir, nil
}

func (m *Manager) buildExtension(
	phpDir string,
	def libraryDef,
	dlls map[string]string,
	iniEnabled map[string]bool,
	loaded map[string]struct{},
) models.PHPExtension {
	dll := findDLL(def, dlls, phpDir)
	_, isLoaded := loaded[def.Name]
	if def.Name == "opcache" {
		_, isLoaded = loaded["opcache"]
		if !isLoaded {
			_, isLoaded = loaded["zend opcache"]
		}
	}
	iniOn := iniEnabled[def.Name] || iniEnabled[iniAlias(def.Name, dll)]

	ext := models.PHPExtension{
		Name:        def.Name,
		DisplayName: def.Display,
		DLL:         dll,
		Zend:        def.Zend,
		Toggleable:  true,
	}

	if dll == "" && !isLoaded {
		ext.Status = models.ExtNotInstalled
		ext.Toggleable = false
		return ext
	}
	if dll == "" && isLoaded {
		ext.Status = models.ExtEnabled
		ext.Builtin = true
		ext.Toggleable = false
		ext.Version = extensionVersion(phpDir, def.Name)
		return ext
	}
	if isLoaded || iniOn {
		ext.Status = models.ExtEnabled
		ext.Version = extensionVersion(phpDir, def.Name)
		return ext
	}
	ext.Status = models.ExtDisabled
	return ext
}

func findDLL(def libraryDef, dlls map[string]string, phpDir string) string {
	for _, name := range def.DLLs {
		if path, ok := dlls[strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(name, "php_"), ".dll"))]; ok {
			return filepath.Base(path)
		}
		if _, err := os.Stat(filepath.Join(phpDir, "ext", name)); err == nil {
			return name
		}
	}
	if path, ok := dlls[def.Name]; ok {
		return filepath.Base(path)
	}
	if def.Name == "gd" {
		if path, ok := dlls["gd2"]; ok {
			return filepath.Base(path)
		}
	}
	return ""
}

func scanExtensionDLLs(phpDir string) map[string]string {
	out := map[string]string{}
	entries, err := os.ReadDir(filepath.Join(phpDir, "ext"))
	if err != nil {
		return out
	}
	for _, e := range entries {
		name := strings.ToLower(e.Name())
		if !strings.HasPrefix(name, "php_") || !strings.HasSuffix(name, ".dll") {
			continue
		}
		key := strings.TrimSuffix(strings.TrimPrefix(name, "php_"), ".dll")
		out[key] = e.Name()
		out[canonicalExtName(key)] = e.Name()
	}
	return out
}

func parseIniExtensions(iniPath string) map[string]bool {
	out := map[string]bool{}
	data, err := os.ReadFile(iniPath)
	if err != nil {
		return out
	}
	for _, line := range strings.Split(string(data), "\n") {
		m := iniExtLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		commented := strings.TrimSpace(m[1]) != ""
		name := canonicalExtName(m[3])
		if !commented {
			out[name] = true
		}
	}
	return out
}

func setExtensionLine(content, name, dll string, zend, enabled bool) (string, error) {
	aliases := iniAliases(name, dll)
	directive := "extension"
	if zend {
		directive = "zend_extension"
	}
	value := name
	if dll != "" {
		value = strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(dll), "php_"), ".dll")
		if name == "gd" && strings.Contains(strings.ToLower(dll), "gd2") {
			value = "gd2"
		} else {
			value = name
		}
	}

	lines := strings.Split(content, "\n")
	found := false
	for i, line := range lines {
		m := iniExtLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		lineName := canonicalExtName(m[3])
		match := false
		for _, a := range aliases {
			if lineName == a || strings.EqualFold(m[3], a) {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		found = true
		prefix := ""
		if !enabled {
			prefix = "; "
		}
		lines[i] = fmt.Sprintf("%s%s=%s", prefix, directive, value)
	}
	if !found {
		if !enabled {
			return content, nil
		}
		lines = append(lines, fmt.Sprintf("%s=%s", directive, value))
	}
	return strings.Join(lines, "\n"), nil
}

func verifyIniChange(iniPath, name, dll string, enabled bool) error {
	parsed := parseIniExtensions(iniPath)
	on := parsed[canonicalExtName(name)]
	if dll != "" {
		key := strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(dll), "php_"), ".dll")
		on = on || parsed[canonicalExtName(key)]
	}
	if enabled && !on {
		return fmt.Errorf("validasi gagal: %s belum aktif di php.ini", name)
	}
	if !enabled && on {
		return fmt.Errorf("validasi gagal: %s masih aktif di php.ini", name)
	}
	return nil
}

func writeIniAtomic(iniPath string, data []byte) error {
	dir := filepath.Dir(iniPath)
	backup := iniPath + ".bak"
	if current, err := os.ReadFile(iniPath); err == nil {
		_ = os.WriteFile(backup, current, 0o644)
	}
	tmp := filepath.Join(dir, "php.ini.larahan.tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("gagal menulis php.ini (izin ditolak?): %w", err)
	}
	if err := os.Rename(tmp, iniPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("gagal menerapkan php.ini: %w", err)
	}
	return nil
}

func assertWritable(iniPath string) error {
	f, err := os.OpenFile(iniPath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("php.ini tidak dapat ditulis (periksa izin file): %w", err)
	}
	_ = f.Close()
	return nil
}

func ensureIniUnderPHPDir(phpDir, iniPath string) error {
	absIni, err := filepath.Abs(filepath.Clean(iniPath))
	if err != nil {
		return err
	}
	absDir, err := filepath.Abs(filepath.Clean(phpDir))
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absDir, absIni)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path php.ini di luar direktori PHP")
	}
	if filepath.Base(absIni) != "php.ini" {
		return fmt.Errorf("hanya php.ini yang boleh diubah")
	}
	return nil
}

func loadedModules(phpDir string) map[string]struct{} {
	out := map[string]struct{}{}
	exe := filepath.Join(phpDir, "php.exe")
	text, err := process.RunCapture(exe, []string{"-m"}, phpDir)
	if err != nil {
		return out
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" || strings.HasPrefix(line, "[") {
			continue
		}
		out[canonicalExtName(line)] = struct{}{}
	}
	return out
}

func detectPHPVersion(phpDir string) string {
	exe := filepath.Join(phpDir, "php.exe")
	text, err := process.RunCapture(exe, []string{"-r", "echo PHP_VERSION;"}, phpDir)
	if err != nil {
		return filepath.Base(phpDir)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return filepath.Base(phpDir)
	}
	return text
}

func extensionVersion(phpDir, name string) string {
	exe := filepath.Join(phpDir, "php.exe")
	script := fmt.Sprintf("echo extension_loaded('%s') ? (phpversion('%s') ?: '') : '';", name, name)
	if name == "opcache" {
		script = "echo (extension_loaded('Zend OPcache') || extension_loaded('opcache')) ? (phpversion('Zend OPcache') ?: phpversion('opcache') ?: '') : '';"
	}
	text, err := process.RunCapture(exe, []string{"-r", script}, phpDir)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(text)
}

func canonicalExtName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.TrimPrefix(n, "php_")
	n = strings.TrimSuffix(n, ".dll")
	switch n {
	case "gd2":
		return "gd"
	case "zend opcache", "zend_opcache":
		return "opcache"
	default:
		return n
	}
}

func iniAlias(name, dll string) string {
	if name == "gd" && strings.Contains(strings.ToLower(dll), "gd2") {
		return "gd2"
	}
	return name
}

func iniAliases(name, dll string) []string {
	out := []string{name, canonicalExtName(name)}
	if dll != "" {
		key := strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(dll), "php_"), ".dll")
		out = append(out, key, canonicalExtName(key))
	}
	if name == "gd" {
		out = append(out, "gd2")
	}
	return out
}

func displayName(name string) string {
	if name == "" {
		return name
	}
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}
