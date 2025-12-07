package installer

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Test file name constants to avoid duplication.
const (
	testFirstLogFile  = "first.log"
	testSecondLogFile = "second.log"
)

// Error message constant for newLoggerWithPaths errors.
const errMsgUnexpectedError = "newLoggerWithPaths() returned unexpected error: %v"

// TestLoggerZeroValue verifies that Logger can be instantiated with zero values.
// This ensures the struct has no unexported initialization requirements.
func TestLoggerZeroValue(t *testing.T) {
	var logger Logger

	// Verify zero value fields
	if logger.file != nil {
		t.Error("Logger zero value: file should be nil")
	}

	if logger.verbose {
		t.Error("Logger zero value: verbose should be false")
	}

	// Verify mutex is usable (zero value is valid)
	// We test that Lock/Unlock don't panic on zero-value mutex
	func() {
		logger.mu.Lock()
		defer logger.mu.Unlock()
	}()
}

// TestLoggerStructFields verifies that Logger struct fields can be set directly.
func TestLoggerStructFields(t *testing.T) {
	// Create a temporary file for testing - t.TempDir() auto-cleans
	tmpFile, err := os.CreateTemp(t.TempDir(), "logger-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	t.Cleanup(func() {
		if err := tmpFile.Close(); err != nil {
			t.Logf("Warning: failed to close temp file: %v", err)
		}
	})

	logger := Logger{
		file:    tmpFile,
		verbose: true,
	}

	if logger.file != tmpFile {
		t.Error("Logger file field not set correctly")
	}

	if !logger.verbose {
		t.Error("Logger verbose field not set correctly")
	}
}

// TestLoggerMutexThreadSafety verifies that Logger's mutex provides thread safety.
func TestLoggerMutexThreadSafety(t *testing.T) {
	t.Parallel()

	var logger Logger
	var wg sync.WaitGroup
	const goroutines = 10

	// Shared state protected by logger.mu
	sharedCounter := 0

	// Launch multiple goroutines that all try to use the mutex
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			logger.mu.Lock()
			sharedCounter++
			logger.mu.Unlock()
		}()
	}

	wg.Wait()

	if sharedCounter != goroutines {
		t.Fatalf("expected sharedCounter to be %d, got %d", goroutines, sharedCounter)
	}
}

// TestLoggerPointerInstantiation verifies Logger can be created as a pointer.
func TestLoggerPointerInstantiation(t *testing.T) {
	logger := &Logger{
		verbose: true,
	}

	if !logger.verbose {
		t.Error("Logger pointer verbose field not set correctly")
	}
}

// TestNewLoggerWithPathsFirstPathWritable verifies that the first writable path is used.
func TestNewLoggerWithPathsFirstPathWritable(t *testing.T) {
	tmpDir := t.TempDir()
	firstPath := filepath.Join(tmpDir, testFirstLogFile)
	secondPath := filepath.Join(tmpDir, testSecondLogFile)

	logger, err := newLoggerWithPaths(false, []string{firstPath, secondPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	// Verify first path was used
	if _, err := os.Stat(firstPath); os.IsNotExist(err) {
		t.Error("Expected first path to be created")
	}

	// Verify second path was NOT created
	if _, err := os.Stat(secondPath); !os.IsNotExist(err) {
		t.Error("Expected second path to NOT be created")
	}

	// Verify logger fields
	if logger.file == nil {
		t.Error("Expected logger.file to be set")
	}
	if logger.verbose {
		t.Error("Expected logger.verbose to be false")
	}
}

// TestNewLoggerWithPathsFallbackToSecondPath verifies fallback when first path is not writable.
func TestNewLoggerWithPathsFallbackToSecondPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory that cannot be written to (no file can be created inside)
	unwritableDir := filepath.Join(tmpDir, "unwritable")
	//nolint:gosec // G301: intentionally testing unwritable directories
	if err := os.Mkdir(unwritableDir, 0o555); err != nil {
		t.Fatalf("Failed to create unwritable directory: %v", err)
	}

	firstPath := filepath.Join(unwritableDir, testFirstLogFile)
	secondPath := filepath.Join(tmpDir, testSecondLogFile)

	logger, err := newLoggerWithPaths(true, []string{firstPath, secondPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
		// Restore permissions for cleanup
		os.Chmod(unwritableDir, 0o755) //nolint:errcheck,gosec // best-effort cleanup in tests
	})

	// Verify first path was NOT created (not writable)
	if _, err := os.Stat(firstPath); !os.IsNotExist(err) {
		t.Error("Expected first path to NOT be created (directory is unwritable)")
	}

	// Verify second path was created (fallback)
	if _, err := os.Stat(secondPath); os.IsNotExist(err) {
		t.Error("Expected second path to be created as fallback")
	}

	// Verify verbose flag
	if !logger.verbose {
		t.Error("Expected logger.verbose to be true")
	}
}

// TestNewLoggerWithPathsAllPathsUnwritable verifies error when all paths are not writable.
func TestNewLoggerWithPathsAllPathsUnwritable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create unwritable directories
	unwritableDir1 := filepath.Join(tmpDir, "unwritable1")
	unwritableDir2 := filepath.Join(tmpDir, "unwritable2")

	//nolint:gosec // G301: intentionally testing unwritable directories
	if err := os.Mkdir(unwritableDir1, 0o555); err != nil {
		t.Fatalf("Failed to create unwritable directory 1: %v", err)
	}
	//nolint:gosec // G301: intentionally testing unwritable directories
	if err := os.Mkdir(unwritableDir2, 0o555); err != nil {
		t.Fatalf("Failed to create unwritable directory 2: %v", err)
	}

	t.Cleanup(func() {
		os.Chmod(unwritableDir1, 0o755) //nolint:errcheck,gosec // best-effort cleanup in tests
		os.Chmod(unwritableDir2, 0o755) //nolint:errcheck,gosec // best-effort cleanup in tests
	})

	firstPath := filepath.Join(unwritableDir1, testFirstLogFile)
	secondPath := filepath.Join(unwritableDir2, testSecondLogFile)

	logger, err := newLoggerWithPaths(false, []string{firstPath, secondPath})

	if err == nil {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
		t.Fatal("Expected error when all paths are unwritable, got nil")
	}

	if logger != nil {
		t.Error("Expected logger to be nil when error is returned")
	}

	// Verify error message contains expected text
	expectedMsg := "failed to open log file"
	if !strings.HasPrefix(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to start with %q, got %q", expectedMsg, err.Error())
	}
}

// TestNewLoggerWithPathsEmptyPaths verifies error when no paths are provided.
func TestNewLoggerWithPathsEmptyPaths(t *testing.T) {
	logger, err := newLoggerWithPaths(false, []string{})

	if err == nil {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
		t.Fatal("Expected error when no paths provided, got nil")
	}

	if logger != nil {
		t.Error("Expected logger to be nil when error is returned")
	}

	expectedMsg := "failed to open log file: no paths provided"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestNewLoggerWithPathsVerboseFlagTrue verifies verbose flag is set correctly when true.
func TestNewLoggerWithPathsVerboseFlagTrue(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := newLoggerWithPaths(true, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	if !logger.verbose {
		t.Error("Expected logger.verbose to be true")
	}
}

// TestNewLoggerWithPathsVerboseFlagFalse verifies verbose flag is set correctly when false.
func TestNewLoggerWithPathsVerboseFlagFalse(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	if logger.verbose {
		t.Error("Expected logger.verbose to be false")
	}
}

// TestNewLoggerWithPathsFileAppendMode verifies that files are opened in append mode.
func TestNewLoggerWithPathsFileAppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "append.log")

	// Pre-create file with content
	initialContent := "existing content\n"
	if err := os.WriteFile(logPath, []byte(initialContent), 0o644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	// Write something to the file
	newContent := "new content\n"
	if _, err := logger.file.WriteString(newContent); err != nil {
		logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		t.Fatalf("Failed to write to log file: %v", err)
	}

	logger.file.Close() //nolint:errcheck // best-effort cleanup in tests

	// Read file and verify both contents exist
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	expectedContent := initialContent + newContent
	if string(content) != expectedContent {
		t.Errorf("Expected file content %q, got %q", expectedContent, string(content))
	}
}

// TestNewLoggerWithPathsFilePermissions verifies that created files have correct permissions.
func TestNewLoggerWithPathsFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "perms.log")

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	// Check file permissions (0600 = rw-------)
	// Note: On some systems, umask may affect the actual permissions
	perm := info.Mode().Perm()
	// We check that the file is readable/writable by owner only
	if perm&0o600 != perm {
		t.Errorf("Expected file permissions 0600 or more restrictive, got %04o", perm)
	}
}

// TestNewLoggerWithPathsSinglePath verifies behavior with only one path.
func TestNewLoggerWithPathsSinglePath(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "single.log")

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	if logger.file == nil {
		t.Error("Expected logger.file to be set")
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

// TestNewLoggerWithPathsMultipleFallbacks verifies fallback through multiple unwritable paths.
func TestNewLoggerWithPathsMultipleFallbacks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple unwritable directories
	unwritable1 := filepath.Join(tmpDir, "unwritable1")
	unwritable2 := filepath.Join(tmpDir, "unwritable2")
	unwritable3 := filepath.Join(tmpDir, "unwritable3")

	for _, dir := range []string{unwritable1, unwritable2, unwritable3} {
		//nolint:gosec // G301: intentionally testing unwritable directories
		if err := os.Mkdir(dir, 0o555); err != nil {
			t.Fatalf("Failed to create unwritable directory: %v", err)
		}
	}

	t.Cleanup(func() {
		for _, dir := range []string{unwritable1, unwritable2, unwritable3} {
			os.Chmod(dir, 0o755) //nolint:errcheck,gosec // best-effort cleanup in tests
		}
	})

	paths := []string{
		filepath.Join(unwritable1, testFirstLogFile),
		filepath.Join(unwritable2, testSecondLogFile),
		filepath.Join(unwritable3, "third.log"),
		filepath.Join(tmpDir, "final.log"), // This one should work
	}

	logger, err := newLoggerWithPaths(false, paths)
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	// Verify only the last path was created
	finalPath := filepath.Join(tmpDir, "final.log")
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		t.Error("Expected final path to be created")
	}

	// Verify others were not created
	for i, path := range paths[:3] {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected path %d (%s) to NOT be created", i+1, path)
		}
	}
}

// TestNewLoggerConstants verifies that the default log path constants are defined correctly.
func TestNewLoggerConstants(t *testing.T) {
	if defaultLogPath != "/var/log/proxmox-install.log" {
		t.Errorf("Expected defaultLogPath to be '/var/log/proxmox-install.log', got %q", defaultLogPath)
	}

	if fallbackLogPath != "/tmp/proxmox-install.log" {
		t.Errorf("Expected fallbackLogPath to be '/tmp/proxmox-install.log', got %q", fallbackLogPath)
	}
}
