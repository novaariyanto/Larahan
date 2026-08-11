package mysql

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/larahan/larahan/backend/process"
)

const mysqldImage = "mysqld.exe"

func (m *Manager) root() string {
	return filepath.Join(m.installPath, "mysql")
}

func (m *Manager) mysqldExe() string {
	return filepath.Join(m.root(), "bin", "mysqld.exe")
}

func (m *Manager) mysqladminExe() string {
	return filepath.Join(m.root(), "bin", "mysqladmin.exe")
}

func (m *Manager) defaultsFile() string {
	return filepath.Join(m.root(), "my.ini")
}

// Start initializes (if needed) and launches mysqld.
func (m *Manager) Start() error {
	if !dirHasMySQL(m.root()) {
		return fmt.Errorf("MySQL belum terinstal")
	}
	if process.IsRunning(mysqldImage) {
		return nil
	}

	port := m.port()
	if err := Configure(m.root(), port); err != nil {
		return fmt.Errorf("konfigurasi MySQL: %w", err)
	}

	if NeedsInit(m.root()) {
		if err := m.initialize(); err != nil {
			return err
		}
	}

	args := []string{"--defaults-file=" + m.defaultsFile()}
	if err := process.StartDetached(m.mysqldExe(), args, filepath.Join(m.root(), "bin")); err != nil {
		return fmt.Errorf("gagal start MySQL: %w", err)
	}

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if process.IsRunning(mysqldImage) && m.ping() {
			return nil
		}
		time.Sleep(400 * time.Millisecond)
	}
	if process.IsRunning(mysqldImage) {
		return nil
	}
	return fmt.Errorf("MySQL tidak merespons setelah start (port %d)", port)
}

// Stop shuts down MySQL gracefully, then force-kills if needed.
func (m *Manager) Stop() error {
	if !process.IsRunning(mysqldImage) {
		return nil
	}

	_, _ = process.RunCapture(m.mysqladminExe(), []string{
		"--defaults-file=" + m.defaultsFile(),
		"-uroot",
		"shutdown",
	}, filepath.Join(m.root(), "bin"))

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if !process.IsRunning(mysqldImage) {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return process.StopImage(mysqldImage)
}

// Restart stops then starts MySQL.
func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return err
	}
	return m.Start()
}

// ApplyPort rewrites my.ini and restarts MySQL when it is running.
func (m *Manager) ApplyPort(port int) error {
	if !dirHasMySQL(m.root()) {
		return nil
	}
	if port <= 0 {
		port = 3306
	}
	if err := Configure(m.root(), port); err != nil {
		return err
	}
	if process.IsRunning(mysqldImage) {
		return m.Restart()
	}
	return nil
}

func (m *Manager) initialize() error {
	args := []string{
		"--defaults-file=" + m.defaultsFile(),
		"--initialize-insecure",
		"--console",
	}
	out, err := process.RunCapture(m.mysqldExe(), args, filepath.Join(m.root(), "bin"))
	if err != nil {
		return fmt.Errorf("inisialisasi datadir gagal: %v (%s)", err, out)
	}
	return nil
}

func (m *Manager) ping() bool {
	out, err := process.RunCapture(m.mysqladminExe(), []string{
		"--defaults-file=" + m.defaultsFile(),
		"-uroot",
		"ping",
	}, filepath.Join(m.root(), "bin"))
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(out), "mysqld is alive")
}

func (m *Manager) port() int {
	if m.settings != nil {
		if p := m.settings.Get().MySQLPort; p > 0 {
			return p
		}
	}
	return 3306
}

func runMysqld(exe string, args ...string) (string, error) {
	return process.RunCapture(exe, args, filepath.Dir(exe))
}
