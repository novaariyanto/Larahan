package downloader

import "github.com/larahan/larahan/backend/models"

// PackageSpec describes a downloadable component package.
type PackageSpec struct {
	Type     models.ComponentType
	Version  string
	Filename string
	URL      string
	SHA256   string // optional; empty skips hash check
	SizeHint int64  // optional expected size; 0 skips size check
}

// Lookup returns a package spec for type+version, or false if unknown.
func Lookup(componentType models.ComponentType, version string) (PackageSpec, bool) {
	key := string(componentType) + "@" + version
	spec, ok := registry[key]
	return spec, ok
}

// ListVersions returns registered versions for a component type.
func ListVersions(componentType models.ComponentType) []string {
	out := make([]string, 0)
	prefix := string(componentType) + "@"
	for key := range registry {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			out = append(out, key[len(prefix):])
		}
	}
	return out
}

// registry maps "type@version" → package metadata.
// PHP uses Thread Safe x64 builds (needed for Apache module).
// URLs use the official "latest" aliases where available.
var registry = map[string]PackageSpec{
	"php@7.4": {
		Type:     models.TypePHP,
		Version:  "7.4",
		Filename: "php-7.4-Win32-vc15-x64-latest.zip",
		URL:      "https://windows.php.net/downloads/releases/latest/php-7.4-Win32-vc15-x64-latest.zip",
	},
	"php@8.1": {
		Type:     models.TypePHP,
		Version:  "8.1",
		Filename: "php-8.1-Win32-vs16-x64-latest.zip",
		URL:      "https://windows.php.net/downloads/releases/latest/php-8.1-Win32-vs16-x64-latest.zip",
	},
	"php@8.2": {
		Type:     models.TypePHP,
		Version:  "8.2",
		Filename: "php-8.2-Win32-vs16-x64-latest.zip",
		URL:      "https://windows.php.net/downloads/releases/latest/php-8.2-Win32-vs16-x64-latest.zip",
	},
	"php@8.3": {
		Type:     models.TypePHP,
		Version:  "8.3",
		Filename: "php-8.3-Win32-vs16-x64-latest.zip",
		URL:      "https://windows.php.net/downloads/releases/latest/php-8.3-Win32-vs16-x64-latest.zip",
	},
	"php@8.4": {
		Type:     models.TypePHP,
		Version:  "8.4",
		Filename: "php-8.4-Win32-vs16-x64-latest.zip",
		URL:      "https://windows.php.net/downloads/releases/latest/php-8.4-Win32-vs16-x64-latest.zip",
	},
	"php@8.5": {
		Type:     models.TypePHP,
		Version:  "8.5",
		Filename: "php-8.5-Win32-vs16-x64-latest.zip",
		URL:      "https://windows.php.net/downloads/releases/latest/php-8.5-Win32-vs16-x64-latest.zip",
	},
	// Apache / MySQL package URLs are finalized in their dedicated phases.
	// Specs are registered so the download pipeline can resolve them once URLs are set.
	"apache@2.4": {
		Type:     models.TypeApache,
		Version:  "2.4",
		Filename: "httpd-2.4.66-251206-Win64-VS17.zip",
		URL:      "https://www.apachelounge.com/download/VS17/binaries/httpd-2.4.66-251206-Win64-VS17.zip",
	},
	"mysql@8.4": {
		Type:     models.TypeMySQL,
		Version:  "8.4",
		Filename: "mysql-8.4.11-winx64.zip",
		URL:      "https://cdn.mysql.com/Downloads/MySQL-8.4/mysql-8.4.11-winx64.zip",
	},
	"phpmyadmin@5.2": {
		Type:     models.TypePhpMyAdmin,
		Version:  "5.2",
		Filename: "phpMyAdmin-5.2.3-all-languages.zip",
		URL:      "https://files.phpmyadmin.net/phpMyAdmin/5.2.3/phpMyAdmin-5.2.3-all-languages.zip",
	},
}
