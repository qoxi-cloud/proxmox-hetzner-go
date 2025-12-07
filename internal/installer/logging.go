// Package installer provides installation orchestration for Proxmox VE on Hetzner dedicated servers.
//
// The installer package coordinates all installation steps including pre-flight checks,
// hardware detection, Proxmox installation, network configuration, and system optimization.
// It uses the exec package for command execution and provides thread-safe logging throughout
// the installation process.
package installer

import (
	"fmt"
	"os"
	"sync"
)

// Default log file paths in priority order.
const (
	// defaultLogPath is the primary log file location.
	// This path is typically writable only by root on Linux systems.
	defaultLogPath = "/var/log/proxmox-install.log"

	// fallbackLogPath is used when the primary path is not writable.
	// This path should be writable on most systems.
	fallbackLogPath = "/tmp/proxmox-install.log"
)

// Logger provides thread-safe logging to file with optional stdout output.
//
// Logger writes timestamped log entries to a file and optionally echoes them to stdout
// when verbose mode is enabled. It uses ISO 8601 timestamps for consistent log formatting.
//
// Logger is safe for concurrent use. All methods use mutex locking to ensure
// thread-safe access to the underlying file handle.
//
// Usage:
//
//	logger, err := NewLogger("/var/log/pve-install.log", true)
//	if err != nil {
//	    return err
//	}
//	defer logger.Close()
//
//	logger.Info("Installation started")
//	logger.Error("Something went wrong: %v", err)
type Logger struct {
	// file is the file handle for log output.
	// It is nil if the logger has not been initialized or has been closed.
	file *os.File

	// verbose enables output to stdout in addition to the log file.
	// When true, all log entries are also written to stdout.
	verbose bool

	// mu protects concurrent access to the file handle.
	mu sync.Mutex
}

// NewLogger creates a new Logger instance.
//
// It attempts to open the log file at /var/log/proxmox-install.log first,
// falling back to /tmp/proxmox-install.log if the primary path is not writable.
// This fallback mechanism ensures logging works in environments where /var/log
// may not be accessible (e.g., development on macOS or non-root execution).
//
// The log file is opened with O_CREATE|O_WRONLY|O_APPEND flags and 0644 permissions,
// allowing the file to be created if it doesn't exist and appending new entries.
//
// Parameters:
//   - verbose: when true, log entries will also be written to stdout
//
// Returns an error if neither log path is writable.
func NewLogger(verbose bool) (*Logger, error) {
	return newLoggerWithPaths(verbose, []string{defaultLogPath, fallbackLogPath})
}

// newLoggerWithPaths creates a Logger using the provided paths in order.
//
// This is an internal helper function that allows testing the path fallback logic
// without requiring access to system directories like /var/log.
//
// The function tries each path in order, returning a Logger using the first
// path that can be opened successfully. If all paths fail, it returns an error
// wrapping the last encountered error.
func newLoggerWithPaths(verbose bool, paths []string) (*Logger, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("failed to open log file: no paths provided")
	}

	var file *os.File
	var lastErr error

	for _, path := range paths {
		var err error
		file, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			break
		}
		lastErr = err
	}

	if file == nil {
		return nil, fmt.Errorf("failed to open log file: %w", lastErr)
	}

	return &Logger{file: file, verbose: verbose}, nil
}
