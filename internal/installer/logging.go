// Package installer provides installation orchestration for Proxmox VE on Hetzner dedicated servers.
//
// The installer package coordinates all installation steps including pre-flight checks,
// hardware detection, Proxmox installation, network configuration, and system optimization.
// It uses the exec package for command execution and provides thread-safe logging throughout
// the installation process.
package installer

import (
	"os"
	"sync"
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
