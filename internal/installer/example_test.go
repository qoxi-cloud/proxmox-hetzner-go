package installer_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qoxi-cloud/proxmox-hetzner-go/internal/installer"
)

// Constants for example tests to avoid code duplication.
const (
	exampleTempDirPattern  = "installer-example-*"
	exampleLogFileName     = "install.log"
	msgFailedCreateTempDir = "failed to create temp dir"
	msgFailedCreateLogger  = "failed to create logger"
)

// ExampleNewLoggerWithPath demonstrates creating a Logger with a custom path.
// This is useful for testing or custom deployment scenarios where the default
// log paths are not suitable.
func ExampleNewLoggerWithPath() {
	// Create a temporary directory for the example
	tmpDir, err := os.MkdirTemp("", exampleTempDirPattern)
	if err != nil {
		fmt.Println(msgFailedCreateTempDir)
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	logPath := filepath.Join(tmpDir, exampleLogFileName)

	// Create logger with custom path (verbose=false)
	logger, err := installer.NewLoggerWithPath(logPath, false)
	if err != nil {
		fmt.Println(msgFailedCreateLogger)
		return
	}
	defer func() { _ = logger.Close() }()

	// Write log messages
	logger.Log("Installation started")
	logger.Log("Processing step %d of %d", 1, 10)

	// Get the log path
	fmt.Println("Logger created successfully")
	fmt.Println("Log path matches:", logger.LogPath() == logPath)

	// Output:
	// Logger created successfully
	// Log path matches: true
}

// ExampleLogger_Log demonstrates writing formatted log messages.
func ExampleLogger_Log() {
	// Create a temporary directory for the example
	tmpDir, err := os.MkdirTemp("", exampleTempDirPattern)
	if err != nil {
		fmt.Println(msgFailedCreateTempDir)
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	logPath := filepath.Join(tmpDir, exampleLogFileName)

	logger, err := installer.NewLoggerWithPath(logPath, false)
	if err != nil {
		fmt.Println(msgFailedCreateLogger)
		return
	}
	defer func() { _ = logger.Close() }()

	// Log simple messages
	logger.Log("Starting pre-flight checks")

	// Log formatted messages with arguments
	logger.Log("Detected %d disks", 2)
	logger.Log("Network interface: %s", "eth0")
	logger.Log("Installation complete in %d seconds", 120)

	fmt.Println("Log messages written successfully")

	// Output:
	// Log messages written successfully
}

// ExampleLogger_Close demonstrates proper resource cleanup.
func ExampleLogger_Close() {
	// Create a temporary directory for the example
	tmpDir, err := os.MkdirTemp("", exampleTempDirPattern)
	if err != nil {
		fmt.Println(msgFailedCreateTempDir)
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	logPath := filepath.Join(tmpDir, exampleLogFileName)

	logger, err := installer.NewLoggerWithPath(logPath, false)
	if err != nil {
		fmt.Println(msgFailedCreateLogger)
		return
	}

	// Write a message
	logger.Log("Installation complete")

	// Close flushes data and releases the file handle
	err = logger.Close()
	if err != nil {
		fmt.Println("failed to close logger")
		return
	}

	// Close is idempotent - safe to call multiple times
	err = logger.Close()
	fmt.Println("First close successful")
	fmt.Println("Second close returns nil:", err == nil)

	// Log after Close should be a safe no-op
	logger.Log("after close")
	fmt.Println("Log after close did not panic")

	// Output:
	// First close successful
	// Second close returns nil: true
	// Log after close did not panic
}

// ExampleLogger_LogPath demonstrates retrieving the log file path.
func ExampleLogger_LogPath() {
	// Create a temporary directory for the example
	tmpDir, err := os.MkdirTemp("", exampleTempDirPattern)
	if err != nil {
		fmt.Println(msgFailedCreateTempDir)
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	logPath := filepath.Join(tmpDir, exampleLogFileName)

	logger, err := installer.NewLoggerWithPath(logPath, false)
	if err != nil {
		fmt.Println(msgFailedCreateLogger)
		return
	}
	defer func() { _ = logger.Close() }()

	// Get the current log path
	currentPath := logger.LogPath()
	fmt.Println("Path is not empty:", currentPath != "")
	fmt.Println("Path ends with install.log:", filepath.Base(currentPath) == exampleLogFileName)

	// Output:
	// Path is not empty: true
	// Path ends with install.log: true
}
