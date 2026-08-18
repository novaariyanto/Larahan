package apache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/larahan/larahan/backend/process"
)

const httpdImage = "httpd.exe"

func (m *Manager) httpdExe() string {
	bin := filepath.Join(m.root(), "bin", "httpd.exe")
	if _, err := os.Stat(bin); err == nil {
		return bin
	}
	return filepath.Join(m.root(), "httpd.exe")
}

func (m *Manager) httpdWorkDir() string {
	return filepath.Dir(m.httpdExe())
}

func (m *Manager) root() string {
	return filepath.Join(m.installPath, "apache")
}

// Start launches Apache httpd in the background.
func (m *Manager) Start() error {
	if !dirHasApache(m.root()) {
		return fmt.Errorf("Apache belum terinstal")
	}
	if process.IsRunning(httpdImage) {
		return nil
	}

	if m.beforeStart != nil {
		if err := m.beforeStart(); err != nil {
			return fmt.Errorf("persiapan sebelum start Apache: %w", err)
		}
	}

	port := m.port()
	if err := Configure(m.root(), port); err != nil {
		return fmt.Errorf("konfigurasi Apache: %w", err)
	}

	if out, err := m.run("-t"); err != nil {
		return fmt.Errorf("httpd config invalid: %s (%v)", out, err)
	}

	if out, err := m.run("-k", "start"); err != nil {
		if err2 := process.StartDetached(m.httpdExe(), []string{"-d", m.root()}, m.httpdWorkDir(), m.phpPath); err2 != nil {
			return fmt.Errorf("gagal start Apache: %v / %v (%s)", err, err2, out)
		}
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if process.IsRunning(httpdImage) {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("Apache tidak merespons setelah start (cek port %d / jalankan sebagai Administrator bila perlu)", port)
}

// Stop gracefully stops Apache, then force-kills if needed.
func (m *Manager) Stop() error {
	if !process.IsRunning(httpdImage) {
		return nil
	}
	_, _ = m.run("-k", "stop")
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if !process.IsRunning(httpdImage) {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return process.StopImage(httpdImage)
}

// Restart stops then starts Apache.
func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return err
	}
	return m.Start()
}

// ApplyPort rewrites Listen and restarts Apache when it is running.
func (m *Manager) ApplyPort(port int) error {
	if !dirHasApache(m.root()) {
		return nil
	}
	if port <= 0 {
		port = 80
	}
	if err := Configure(m.root(), port); err != nil {
		return err
	}
	if process.IsRunning(httpdImage) {
		return m.Restart()
	}
	return nil
}

func (m *Manager) port() int {
	if m.settings != nil {
		if p := m.settings.Get().ApachePort; p > 0 {
			return p
		}
	}
	return 80
}

func (m *Manager) run(args ...string) (string, error) {
	full := append([]string{"-d", m.root()}, args...)
	return process.RunCapture(m.httpdExe(), full, m.httpdWorkDir(), m.phpPath)
}

func runHttpd(exe string, args ...string) (string, error) {
	return process.RunCapture(exe, args, filepath.Dir(exe))
}
