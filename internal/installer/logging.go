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
	"time"
)

// Default log file paths in priority order.
const (
	// Primary log file location.
	// This path is typically writable only by root on Linux systems.
	defaultLogPath = "/var/log/proxmox-install.log"

	// Fallback log path used when the primary path is not writable.
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
//	logger, err := NewLogger(true)
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
// The log file is opened with O_CREATE|O_WRONLY|O_APPEND flags and 0600 permissions,
// creating the file if it doesn't exist, appending new entries, and restricting
// access to the current user by default.
//
// Parameters:
//   - verbose: when true, log entries will also be written to stdout
//
// Returns an error if neither log path is writable.
func NewLogger(verbose bool) (*Logger, error) {
	return newLoggerWithPaths(verbose, []string{defaultLogPath, fallbackLogPath})
}

// NewLoggerWithPath creates a Logger with a custom log file path.
//
// This constructor is primarily useful for testing or special deployment scenarios
// where the default log paths (/var/log/proxmox-install.log or /tmp/proxmox-install.log)
// are not suitable. It allows tests to use temporary directories without requiring
// root permissions.
//
// The log file is opened with O_CREATE|O_WRONLY|O_APPEND flags and 0644 permissions,
// creating the file if it doesn't exist and appending new entries.
//
// Parameters:
//   - path: the absolute path where the log file should be created
//   - verbose: when true, log entries will also be written to stdout
//
// Returns an error if the file cannot be opened or created at the specified path.
//
// Example usage:
//
//	// In tests
//	tmpDir := t.TempDir()
//	logPath := filepath.Join(tmpDir, "test.log")
//	logger, err := NewLoggerWithPath(logPath, false)
//	if err != nil {
//	    t.Fatalf("failed to create logger: %v", err)
//	}
//	defer logger.Close()
func NewLoggerWithPath(path string, verbose bool) (*Logger, error) {
	//nolint:gosec // G304: path is user-provided; G302: 0644 allows log readability in test/deployment scenarios
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
	}

	return &Logger{file: file, verbose: verbose}, nil
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

		//nolint:gosec // G304: paths are controlled constants in production
		file, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)

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

// Log writes a formatted message to the log file with an ISO 8601 timestamp.
//
// If verbose mode is enabled, the message is also printed to stdout.
// The format string and args follow fmt.Sprintf conventions.
//
// Example output format:
//
//	[2024-01-15T10:30:45Z] Installation started
//	[2024-01-15T10:30:46Z] Detected network interface: eth0
//
// Log is safe for concurrent use. It is a no-op if the Logger is nil or
// if the underlying file has not been initialized. Errors from writing
// to the file are intentionally ignored to avoid interrupting the
// installation process.
//
//nolint:goprintffuncname // Log is the intended API name per project spec
func (l *Logger) Log(format string, args ...interface{}) {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] %s\n", timestamp, msg)

	// Write to file - errors are intentionally ignored as logging
	// should not interrupt the installation process.
	if _, err := l.file.WriteString(line); err != nil {
		// Intentionally ignored: logging failures should not
		// interrupt the installation process.
		_ = err
	}

	if l.verbose {
		fmt.Print(line)
	}
}

// Close flushes any buffered data and closes the log file.
//
// It should be called when the logger is no longer needed to ensure all
// buffered data is written and the file handle is properly released.
//
// Close is safe for concurrent use. It is a no-op if the Logger is nil or
// if the underlying file has already been closed (idempotent behavior).
//
// Returns an error if the file cannot be synced or closed properly.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return nil
	}

	// Sync to flush any buffered data to disk before closing.
	// We capture both errors to ensure the file is always closed,
	// even if Sync fails, preventing file descriptor leaks.
	syncErr := l.file.Sync()
	closeErr := l.file.Close()

	// Set file to nil for idempotent behavior.
	// This ensures:
	// 1. Subsequent Close calls return nil without error
	// 2. Log method becomes a no-op after Close
	l.file = nil

	// Return sync error first (it's more informative about data loss)
	if syncErr != nil {
		return fmt.Errorf("failed to sync log file: %w", syncErr)
	}

	if closeErr != nil {
		return fmt.Errorf("failed to close log file: %w", closeErr)
	}

	return nil
}

// LogPath returns the path to the current log file.
//
// This is useful for displaying the log location to users in the TUI or CLI,
// allowing them to know where logs are being written during installation.
//
// LogPath is safe for concurrent use. It returns an empty string if the Logger
// is nil or if the underlying file has not been initialized or has been closed.
func (l *Logger) LogPath() string {
	if l == nil {
		return ""
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return ""
	}

	return l.file.Name()
}
