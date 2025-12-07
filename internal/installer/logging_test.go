package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// Test file name constants to avoid duplication.
const (
	testFirstLogFile  = "first.log"
	testSecondLogFile = "second.log"
	testLogFileName   = "test.log"
)

// Error message constants for test assertions.
const (
	errMsgUnexpectedError      = "newLoggerWithPaths() returned unexpected error: %v"
	errMsgExpectedFileSet      = "Expected logger.file to be set"
	errMsgExpectedVerboseFalse = "Expected logger.verbose to be false"
	errMsgExpectedVerboseTrue  = "Expected logger.verbose to be true"
)

// Log method test constants.
const (
	errMsgLogFileReadFailed   = "Failed to read log file: %v"
	errMsgTimestampNotMatched = "Log entry does not contain valid RFC3339 timestamp"
	errMsgMessageNotFound     = "Expected message %q not found in log content"
	errMsgSyncLogFileFailed   = "Failed to sync log file: %v"
	testLogMessage            = "Test log message"
	testFormatMessage         = "Value: %d, Name: %s"
)

// rfc3339Pattern matches RFC3339 timestamps in log entries.
// Example: [2024-01-15T10:30:45Z] or [2024-01-15T10:30:45+02:00].
var rfc3339Pattern = regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2})\]`)

// createTestLogger creates a Logger for testing with automatic cleanup.
// It returns the logger and the path to the log file.
func createTestLogger(t *testing.T, verbose bool) (logger *Logger, logPath string) {
	t.Helper()

	tmpDir := t.TempDir()
	logPath = filepath.Join(tmpDir, testLogFileName)

	var err error

	logger, err = newLoggerWithPaths(verbose, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil && logger.file != nil {
			logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	return logger, logPath
}

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

	// Verify Log is a no-op and doesn't panic when file is nil
	logger.Log("test message should not panic")
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
		t.Error(errMsgExpectedFileSet)
	}
	if logger.verbose {
		t.Error(errMsgExpectedVerboseFalse)
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
		os.Chmod(unwritableDir, 0o700) //nolint:errcheck,gosec // G302: directories need execute bit for cleanup
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
		t.Error(errMsgExpectedVerboseTrue)
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
		os.Chmod(unwritableDir1, 0o700) //nolint:errcheck,gosec // G302: directories need execute bit for cleanup
		os.Chmod(unwritableDir2, 0o700) //nolint:errcheck,gosec // G302: directories need execute bit for cleanup
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
	logger, _ := createTestLogger(t, true)

	if !logger.verbose {
		t.Error(errMsgExpectedVerboseTrue)
	}
}

// TestNewLoggerWithPathsVerboseFlagFalse verifies verbose flag is set correctly when false.
func TestNewLoggerWithPathsVerboseFlagFalse(t *testing.T) {
	logger, _ := createTestLogger(t, false)

	if logger.verbose {
		t.Error(errMsgExpectedVerboseFalse)
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
	_, logPath := createTestLogger(t, false)

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
	logger, logPath := createTestLogger(t, false)

	if logger.file == nil {
		t.Error(errMsgExpectedFileSet)
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
			os.Chmod(dir, 0o700) //nolint:errcheck,gosec // G302: directories need execute bit for cleanup
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

// TestNewLoggerWithDefaultPaths tests the public NewLogger constructor with verbose=false.
// This test may skip if neither default log path is writable (e.g., in CI environments).
func TestNewLoggerWithDefaultPaths(t *testing.T) {
	logger, err := NewLogger(false)
	if err != nil {
		t.Skipf("Skipping test: cannot create logger with default paths: %v", err)
	}

	t.Cleanup(func() {
		logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
	})

	if logger.file == nil {
		t.Error(errMsgExpectedFileSet)
	}

	if logger.verbose {
		t.Error(errMsgExpectedVerboseFalse)
	}
}

// TestNewLoggerVerboseTrue tests the public NewLogger constructor with verbose=true.
// This test may skip if neither default log path is writable (e.g., in CI environments).
func TestNewLoggerVerboseTrue(t *testing.T) {
	logger, err := NewLogger(true)
	if err != nil {
		t.Skipf("Skipping test: cannot create logger with default paths: %v", err)
	}

	t.Cleanup(func() {
		logger.file.Close() //nolint:errcheck // best-effort cleanup in tests
	})

	if logger.file == nil {
		t.Error(errMsgExpectedFileSet)
	}

	if !logger.verbose {
		t.Error(errMsgExpectedVerboseTrue)
	}
}

// TestLogWritesToFile verifies that Log writes messages to the log file.
func TestLogWritesToFile(t *testing.T) {
	logger, logPath := createTestLogger(t, false)

	logger.Log(testLogMessage)

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	if !strings.Contains(string(content), testLogMessage) {
		t.Errorf(errMsgMessageNotFound, testLogMessage)
	}
}

// TestLogTimestampFormat verifies that Log uses RFC3339 (ISO 8601) timestamp format.
func TestLogTimestampFormat(t *testing.T) {
	logger, logPath := createTestLogger(t, false)

	beforeLog := time.Now()
	logger.Log(testLogMessage)
	afterLog := time.Now()

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	logLine := string(content)

	// Verify timestamp format matches RFC3339 pattern
	if !rfc3339Pattern.MatchString(logLine) {
		t.Errorf("%s: got %q", errMsgTimestampNotMatched, logLine)
	}

	// Extract and parse the timestamp to verify it's within expected range
	timestampMatch := regexp.MustCompile(`\[([^\]]+)\]`).FindStringSubmatch(logLine)
	if len(timestampMatch) < 2 {
		t.Fatalf("Could not extract timestamp from log line: %q", logLine)
	}

	parsedTime, err := time.Parse(time.RFC3339, timestampMatch[1])
	if err != nil {
		t.Fatalf("Failed to parse timestamp %q: %v", timestampMatch[1], err)
	}

	// Verify timestamp is within the expected time window
	if parsedTime.Before(beforeLog.Add(-time.Second)) || parsedTime.After(afterLog.Add(time.Second)) {
		t.Errorf("Timestamp %v is outside expected range [%v, %v]", parsedTime, beforeLog, afterLog)
	}
}

// TestLogWithFormatArgs verifies that Log correctly formats messages with arguments.
func TestLogWithFormatArgs(t *testing.T) {
	logger, logPath := createTestLogger(t, false)

	logger.Log(testFormatMessage, 42, "test")

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	expectedMsg := "Value: 42, Name: test"
	if !strings.Contains(string(content), expectedMsg) {
		t.Errorf(errMsgMessageNotFound, expectedMsg)
	}
}

// TestLogVerboseModeWritesToStdout verifies that verbose mode outputs to stdout.
func TestLogVerboseModeWritesToStdout(t *testing.T) {
	logger, _ := createTestLogger(t, true)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	os.Stdout = w

	logger.Log(testLogMessage)

	// Close writer and restore stdout
	w.Close() //nolint:errcheck // best-effort cleanup in tests
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	capturedOutput := buf.String()

	// Verify message was written to stdout
	if !strings.Contains(capturedOutput, testLogMessage) {
		t.Errorf("Expected stdout to contain %q, got %q", testLogMessage, capturedOutput)
	}

	// Verify timestamp format in stdout
	if !rfc3339Pattern.MatchString(capturedOutput) {
		t.Errorf("%s in stdout: got %q", errMsgTimestampNotMatched, capturedOutput)
	}
}

// TestLogNonVerboseModeNoStdout verifies that non-verbose mode does not output to stdout.
func TestLogNonVerboseModeNoStdout(t *testing.T) {
	logger, _ := createTestLogger(t, false)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	os.Stdout = w

	logger.Log(testLogMessage)

	// Close writer and restore stdout
	w.Close() //nolint:errcheck // best-effort cleanup in tests
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	capturedOutput := buf.String()

	// Verify nothing was written to stdout
	if capturedOutput != "" {
		t.Errorf("Expected no stdout output in non-verbose mode, got %q", capturedOutput)
	}
}

// TestLogConcurrentCalls verifies that concurrent Log calls are thread-safe.
func TestLogConcurrentCalls(t *testing.T) {
	t.Parallel()

	logger, logPath := createTestLogger(t, false)

	const goroutines = 50
	var wg sync.WaitGroup

	// Launch multiple goroutines that all call Log
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Log("Goroutine %d logging", id)
		}(i)
	}

	wg.Wait()

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	// Verify we have exactly the expected number of log entries
	if len(lines) != goroutines {
		t.Errorf("Expected %d log lines, got %d", goroutines, len(lines))
	}

	// Verify each line has valid format (no interleaving)
	for i, line := range lines {
		// Check timestamp format
		if !rfc3339Pattern.MatchString(line) {
			t.Errorf("Line %d has invalid format (possible interleaving): %q", i+1, line)
		}

		// Check that line contains "Goroutine" message
		if !strings.Contains(line, "Goroutine") {
			t.Errorf("Line %d missing expected message: %q", i+1, line)
		}

		// Verify line ends with newline indicator (properly terminated)
		if !strings.Contains(line, "logging") {
			t.Errorf("Line %d appears to be truncated or interleaved: %q", i+1, line)
		}
	}
}

// TestLogMultipleMessages verifies that multiple Log calls append correctly.
func TestLogMultipleMessages(t *testing.T) {
	logger, logPath := createTestLogger(t, false)

	messages := []string{
		"First message",
		"Second message",
		"Third message",
	}

	for _, msg := range messages {
		logger.Log("%s", msg)
	}

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	// Verify all messages are present
	for _, msg := range messages {
		if !strings.Contains(string(content), msg) {
			t.Errorf(errMsgMessageNotFound, msg)
		}
	}

	// Verify order by checking line count
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(messages) {
		t.Errorf("Expected %d log lines, got %d", len(messages), len(lines))
	}
}

// TestLogEmptyMessage verifies that Log handles empty messages correctly.
func TestLogEmptyMessage(t *testing.T) {
	logger, logPath := createTestLogger(t, false)

	logger.Log("")

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	// Verify timestamp is still written even for empty message
	if !rfc3339Pattern.MatchString(string(content)) {
		t.Error("Expected timestamp to be written even for empty message")
	}

	// Verify format is "[timestamp] \n"
	if !strings.Contains(string(content), "] \n") {
		t.Error("Expected empty message format to be '[timestamp] \\n'")
	}
}

// TestLogSpecialCharacters verifies that Log handles special characters correctly.
func TestLogSpecialCharacters(t *testing.T) {
	logger, logPath := createTestLogger(t, false)

	tests := []struct {
		name    string
		message string
	}{
		{"unicode", "Message with unicode: \u4e2d\u6587"},
		{"newlines", "Message with\nnewlines"},
		{"tabs", "Message\twith\ttabs"},
		{"special chars", "Special: !@#$%^&*()"},
		{"quotes", `Message with "quotes" and 'apostrophes'`},
	}

	for _, tt := range tests {
		logger.Log("%s", tt.message)
	}

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	for _, tt := range tests {
		if !strings.Contains(string(content), tt.message) {
			t.Errorf("Test %q: "+errMsgMessageNotFound, tt.name, tt.message)
		}
	}
}

// TestLogLineFormat verifies the exact format of log lines.
func TestLogLineFormat(t *testing.T) {
	logger, logPath := createTestLogger(t, false)

	testMsg := "Test format verification"
	logger.Log("%s", testMsg)

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	// Verify exact format: [TIMESTAMP] MESSAGE\n
	// Example: [2024-01-15T10:30:45Z] Test format verification\n
	formatPattern := regexp.MustCompile(
		`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2})\] ` +
			regexp.QuoteMeta(testMsg) + `\n$`)

	if !formatPattern.Match(content) {
		t.Errorf("Log line does not match expected format.\nExpected pattern: [TIMESTAMP] %s\\n\nGot: %q",
			testMsg, string(content))
	}
}

// TestLogWithNilLogger verifies that Log is a no-op when called on nil Logger.
func TestLogWithNilLogger(t *testing.T) {
	var logger *Logger

	// Verify Log doesn't panic when called on nil Logger
	// Test passes if no panic occurs
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Log panicked on nil Logger: %v", r)
		}
	}()

	logger.Log("test message should not panic")
	logger.Log("formatted message: %d", 42)
}
