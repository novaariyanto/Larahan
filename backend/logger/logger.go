package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger writes simple timestamped lines to a file and stdout.
type Logger struct {
	mu     sync.Mutex
	file   *os.File
	stdLog *log.Logger
}

// New creates a logger that appends to logDir/larahan.log.
func New(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	path := filepath.Join(logDir, "larahan.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:   f,
		stdLog: log.New(os.Stdout, "", 0),
	}, nil
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...any) {
	l.write("INFO", format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...any) {
	l.write("ERROR", format, args...)
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *Logger) write(level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s [%s] %s\n", time.Now().Format(time.RFC3339), level, msg)
	l.stdLog.Print(line)
	if l.file != nil {
		_, _ = l.file.WriteString(line)
	}
}
