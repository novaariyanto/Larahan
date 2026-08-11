package models

// ComponentStatus describes the runtime state of a stack component.
type ComponentStatus string

const (
	StatusNotInstalled ComponentStatus = "not_installed"
	StatusStopped      ComponentStatus = "stopped"
	StatusRunning      ComponentStatus = "running"
	StatusReady        ComponentStatus = "ready"
	StatusError        ComponentStatus = "error"
	StatusInstalling   ComponentStatus = "installing"
)

// ComponentType identifies a managed package kind.
type ComponentType string

const (
	TypeApache      ComponentType = "apache"
	TypePHP         ComponentType = "php"
	TypeMySQL       ComponentType = "mysql"
	TypePhpMyAdmin  ComponentType = "phpmyadmin"
)

// ComponentInfo is shared across Dashboard / Apache / PHP / MySQL views.
type ComponentInfo struct {
	Type      ComponentType   `json:"type"`
	Version   string          `json:"version"`
	Status    ComponentStatus `json:"status"`
	Path      string          `json:"path"`
	Installed bool            `json:"installed"`
}

// PHPVersionInfo describes one supported PHP version entry.
type PHPVersionInfo struct {
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
	Active    bool   `json:"active"`
	Path      string `json:"path"`
}

// DashboardSummary is the Dashboard overview payload.
type DashboardSummary struct {
	Apache     ComponentInfo `json:"apache"`
	MySQL      ComponentInfo `json:"mysql"`
	ActivePHP  string        `json:"active_php"`
	PHP        ComponentInfo `json:"php"`
	PhpMyAdmin ComponentInfo `json:"phpmyadmin"`
}

// Settings holds runtime configuration persisted as JSON.
type Settings struct {
	InstallPath string `json:"install_path"`
	ActivePHP   string `json:"active_php"`
	ApachePort  int    `json:"apache_port"`
	MySQLPort   int    `json:"mysql_port"`
	FirstRun    bool   `json:"first_run"`
}

// Paths lists standard Larahan directories under the install root.
type Paths struct {
	Root       string `json:"root"`
	Apache     string `json:"apache"`
	PHP        string `json:"php"`
	MySQL      string `json:"mysql"`
	PhpMyAdmin string `json:"phpmyadmin"`
	Downloads  string `json:"downloads"`
	Logs       string `json:"logs"`
	Config     string `json:"config"`
	Temp       string `json:"temp"`
}

// Result is a generic operation response for the frontend.
type Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// OKResult returns a successful Result.
func OKResult(message string) Result {
	return Result{OK: true, Message: message}
}

// FailResult returns a failed Result.
func FailResult(message string) Result {
	return Result{OK: false, Message: message}
}
