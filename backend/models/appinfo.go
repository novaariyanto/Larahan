package models

// AppInfo describes the Larahan application itself.
type AppInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}
