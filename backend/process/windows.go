package process

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// IsRunning reports whether a process with the given image name is alive (e.g. "httpd.exe").
func IsRunning(imageName string) bool {
	pids, err := findPIDs(imageName)
	return err == nil && len(pids) > 0
}

// StartDetached starts a process without attaching a console window.
// extraPath directories are prepended to PATH for the child (PHP cURL/OpenSSL DLLs).
func StartDetached(exe string, args []string, workDir string, extraPath ...string) error {
	cmd := exec.Command(exe, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	applyExtraPath(cmd, extraPath...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		HideWindow:    true,
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

// RunCapture runs a command and returns combined stdout/stderr.
func RunCapture(exe string, args []string, workDir string, extraPath ...string) (string, error) {
	cmd := exec.Command(exe, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	applyExtraPath(cmd, extraPath...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000,
		HideWindow:    true,
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return strings.TrimSpace(buf.String()), err
}

func applyExtraPath(cmd *exec.Cmd, extraDirs ...string) {
	var extras []string
	for _, d := range extraDirs {
		d = strings.TrimSpace(d)
		if d != "" {
			extras = append(extras, d)
		}
	}
	if len(extras) == 0 {
		return
	}
	prefix := strings.Join(extras, string(os.PathListSeparator)) + string(os.PathListSeparator)
	env := os.Environ()
	found := false
	for i, e := range env {
		eq := strings.IndexByte(e, '=')
		if eq <= 0 {
			continue
		}
		if strings.EqualFold(e[:eq], "PATH") {
			env[i] = e[:eq+1] + prefix + e[eq+1:]
			found = true
			break
		}
	}
	if !found {
		env = append(env, "PATH="+prefix+os.Getenv("PATH"))
	}
	cmd.Env = env
}

// StopImage kills all processes matching imageName.
func StopImage(imageName string) error {
	pids, err := findPIDs(imageName)
	if err != nil {
		return err
	}
	if len(pids) == 0 {
		return nil
	}
	var lastErr error
	for _, pid := range pids {
		if err := killPID(pid); err != nil {
			lastErr = err
		}
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !IsRunning(imageName) {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("proses %s masih berjalan", imageName)
}

func findPIDs(imageName string) ([]int, error) {
	out, err := RunCapture("tasklist", []string{"/FI", "IMAGENAME eq " + imageName, "/FO", "CSV", "/NH"}, "")
	if err != nil {
		// tasklist returns exit 1 when no tasks match
		if strings.Contains(strings.ToLower(out), "no tasks") || out == "" {
			return nil, nil
		}
		return nil, err
	}
	if strings.Contains(strings.ToLower(out), "no tasks") {
		return nil, nil
	}

	pids := make([]int, 0)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// "httpd.exe","1234","Session","1","12,345 K"
		parts := splitCSV(line)
		if len(parts) < 2 {
			continue
		}
		pid, err := strconv.Atoi(strings.Trim(parts[1], `"`))
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func killPID(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func splitCSV(line string) []string {
	out := make([]string, 0)
	var cur strings.Builder
	inQuotes := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '"':
			inQuotes = !inQuotes
			cur.WriteByte(ch)
		case ch == ',' && !inQuotes:
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(ch)
		}
	}
	out = append(out, cur.String())
	return out
}
