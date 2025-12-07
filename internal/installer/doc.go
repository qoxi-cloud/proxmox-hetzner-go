// Package installer provides installation orchestration for Proxmox VE
// on Hetzner dedicated servers.
//
// The installer package coordinates all installation steps including pre-flight checks,
// hardware detection, Proxmox installation, network configuration, and system optimization.
// It uses the exec package for command execution and provides thread-safe logging throughout
// the installation process.
//
// # Logger
//
// The Logger provides thread-safe logging to file with optional stdout output.
// It automatically handles log file location with fallback:
//
//	logger, err := installer.NewLogger(verbose)
//	if err != nil {
//	    return err
//	}
//	defer logger.Close()
//
//	logger.Log("Starting installation...")
//	logger.Log("Processing step %d of %d", current, total)
//
// Log files are written to /var/log/proxmox-install.log by default,
// with automatic fallback to /tmp/proxmox-install.log if /var/log
// is not writable.
//
// For testing or custom deployments, use NewLoggerWithPath to specify
// a custom log file location:
//
//	logger, err := installer.NewLoggerWithPath("/custom/path/install.log", verbose)
//	if err != nil {
//	    return err
//	}
//	defer logger.Close()
//
// # Log Format
//
// Each log entry follows this format:
//
//	[2024-01-15T10:30:45Z] Message text here
//
// Timestamps use ISO 8601 format (RFC3339) for consistent parsing and
// timezone-independent logging. All timestamps are in UTC.
//
// # Log File Paths
//
// The default log path selection follows this priority:
//
//  1. /var/log/proxmox-install.log (primary, requires root on Linux)
//  2. /tmp/proxmox-install.log (fallback, generally writable)
//
// The fallback mechanism ensures logging works in environments where /var/log
// may not be accessible (e.g., development on macOS or non-root execution).
//
// To determine the actual log file path being used:
//
//	logPath := logger.LogPath()
//	fmt.Printf("Logs are being written to: %s\n", logPath)
//
// # Thread Safety
//
// All Logger methods are safe for concurrent use. The Logger uses mutex locking
// to ensure thread-safe access to the underlying file handle. This allows
// multiple goroutines to log messages simultaneously without data races or
// interleaved output.
//
// # Resource Management
//
// Always call Close when the Logger is no longer needed to ensure all buffered
// data is flushed to disk and the file handle is released:
//
//	logger, err := installer.NewLogger(false)
//	if err != nil {
//	    return err
//	}
//	defer logger.Close()
//
// After Close is called, subsequent Log calls become no-ops (they do not panic).
// Close is idempotent - calling it multiple times is safe and returns nil after
// the first successful close.
package installer
