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
	errMsgUnexpectedError             = "newLoggerWithPaths() returned unexpected error: %v"
	errMsgNewLoggerWithPathUnexpected = "NewLoggerWithPath() returned unexpected error: %v"
	errMsgExpectedFileSet             = "Expected logger.file to be set"
	errMsgExpectedVerboseFalse        = "Expected logger.verbose to be false"
	errMsgExpectedVerboseTrue         = "Expected logger.verbose to be true"
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

// Close method test constants.
const (
	errMsgCloseUnexpected = "Close() returned unexpected error: %v"
)

// LogPath method test constants.
const (
	errMsgLogPathExpectedEmpty    = "LogPath() expected empty string, got %q"
	errMsgLogPathExpectedPath     = "LogPath() expected %q, got %q"
	errMsgLogPathUnexpectedPrefix = "LogPath() expected path with prefix %q, got %q"
)

// Directory and skip message constants.
const (
	errMsgCreateUnwritableDir    = "Failed to create unwritable directory: %v"
	errMsgSkipCannotCreateLogger = "Skipping test: cannot create logger with default paths: %v"
)

// Error handling test constants.
const (
	errMsgExpectedLoggerNil     = "Expected logger to be nil when error is returned"
	errMsgFailedToOpenLogFile   = "failed to open log file"
	errMsgExpectedErrorMsgStart = "Expected error message to start with %q, got %q"
	testMsgBeforeClose          = "message before close"
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
		t.Fatalf(errMsgCreateUnwritableDir, err)
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
		t.Error(errMsgExpectedLoggerNil)
	}

	// Verify error message contains expected text
	if !strings.HasPrefix(err.Error(), errMsgFailedToOpenLogFile) {
		t.Errorf(errMsgExpectedErrorMsgStart, errMsgFailedToOpenLogFile, err.Error())
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
		t.Error(errMsgExpectedLoggerNil)
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
			t.Fatalf(errMsgCreateUnwritableDir, err)
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
		t.Skipf(errMsgSkipCannotCreateLogger, err)
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
		t.Skipf(errMsgSkipCannotCreateLogger, err)
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

// TestCloseValidLogger verifies that Close returns nil for a valid logger.
func TestCloseValidLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, testLogFileName)

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	// Write something to verify sync works
	logger.Log("test message before close")

	// Close should succeed
	err = logger.Close()
	if err != nil {
		t.Errorf(errMsgCloseUnexpected, err)
	}

	// Verify file exists and contains the message
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	if !strings.Contains(string(content), "test message before close") {
		t.Error("Expected log message to be flushed to file before close")
	}
}

// TestCloseIdempotent verifies that calling Close twice is safe and idempotent.
func TestCloseIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, testLogFileName)

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	// First close should succeed
	err = logger.Close()
	if err != nil {
		t.Errorf("First Close() returned unexpected error: %v", err)
	}

	// Second close should also succeed (return nil)
	err = logger.Close()
	if err != nil {
		t.Errorf("Second Close() should return nil, got: %v", err)
	}

	// Third close for good measure
	err = logger.Close()
	if err != nil {
		t.Errorf("Third Close() should return nil, got: %v", err)
	}
}

// TestCloseNilLogger verifies that Close is safe when called on nil Logger.
func TestCloseNilLogger(t *testing.T) {
	var logger *Logger

	// Verify Close doesn't panic on nil Logger
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close panicked on nil Logger: %v", r)
		}
	}()

	err := logger.Close()
	if err != nil {
		t.Errorf("Close() on nil Logger should return nil, got: %v", err)
	}
}

// TestCloseNilFile verifies that Close is safe when file is nil.
func TestCloseNilFile(t *testing.T) {
	// Create logger with nil file (zero value)
	logger := &Logger{}

	err := logger.Close()
	if err != nil {
		t.Errorf("Close() with nil file should return nil, got: %v", err)
	}
}

// TestCloseFlushesData verifies that Close syncs data before closing.
func TestCloseFlushesData(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, testLogFileName)

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	// Write multiple messages
	messages := []string{
		"First message to flush",
		"Second message to flush",
		"Third message to flush",
	}
	for _, msg := range messages {
		logger.Log("%s", msg)
	}

	// Close should sync and close
	err = logger.Close()
	if err != nil {
		t.Fatalf(errMsgCloseUnexpected, err)
	}

	// Read file and verify all messages are present
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	for _, msg := range messages {
		if !strings.Contains(string(content), msg) {
			t.Errorf("Expected message %q to be flushed to file", msg)
		}
	}
}

// TestCloseLogAfterClose verifies that Log becomes a no-op after Close.
func TestCloseLogAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, testLogFileName)

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	// Write initial message
	logger.Log(testMsgBeforeClose)

	// Close the logger
	err = logger.Close()
	if err != nil {
		t.Fatalf(errMsgCloseUnexpected, err)
	}

	// Log after close should not panic and should be a no-op
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Log after Close panicked: %v", r)
		}
	}()

	logger.Log("message after close - should be ignored")

	// Verify only the first message is in the file
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	if !strings.Contains(string(content), testMsgBeforeClose) {
		t.Error("Expected message before close to be present")
	}

	if strings.Contains(string(content), "message after close") {
		t.Error("Message after close should not be written to file")
	}
}

// TestCloseConcurrent verifies that concurrent Close calls are thread-safe.
func TestCloseConcurrent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, testLogFileName)

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	// Write a message first
	logger.Log("message before concurrent close")

	const goroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	// Launch multiple goroutines that all try to close
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := logger.Close(); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for unexpected errors
	// At most one goroutine should successfully close, others should get nil
	for err := range errors {
		t.Errorf("Unexpected error during concurrent Close: %v", err)
	}
}

// TestCloseAndLogConcurrent verifies thread safety with concurrent Log and Close.
func TestCloseAndLogConcurrent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, testLogFileName)

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	const goroutines = 20
	var wg sync.WaitGroup

	// Launch goroutines that log
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Log("Concurrent log %d", id)
		}(i)
	}

	// Launch goroutines that try to close
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Close() //nolint:errcheck // best-effort cleanup in concurrent test
		}()
	}

	// Test passes if no panic or race condition occurs
	wg.Wait()
}

// TestCloseWithZeroValueLogger verifies Close on zero-value Logger.
func TestCloseWithZeroValueLogger(t *testing.T) {
	var logger Logger

	err := logger.Close()
	if err != nil {
		t.Errorf("Close() on zero-value Logger should return nil, got: %v", err)
	}
}

// TestLogPathReturnsCorrectPath verifies that LogPath returns the correct path after NewLogger.
func TestLogPathReturnsCorrectPath(t *testing.T) {
	logger, logPath := createTestLogger(t, false)

	result := logger.LogPath()
	if result != logPath {
		t.Errorf(errMsgLogPathExpectedPath, logPath, result)
	}
}

// TestLogPathNilLogger verifies that LogPath returns empty string on nil Logger.
func TestLogPathNilLogger(t *testing.T) {
	var logger *Logger

	// Verify LogPath doesn't panic on nil Logger
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogPath panicked on nil Logger: %v", r)
		}
	}()

	result := logger.LogPath()
	if result != "" {
		t.Errorf(errMsgLogPathExpectedEmpty, result)
	}
}

// TestLogPathNilFile verifies that LogPath returns empty string when file is nil.
func TestLogPathNilFile(t *testing.T) {
	// Create logger with nil file (zero value)
	logger := &Logger{}

	result := logger.LogPath()
	if result != "" {
		t.Errorf(errMsgLogPathExpectedEmpty, result)
	}
}

// TestLogPathAfterClose verifies that LogPath returns empty string after Close is called.
func TestLogPathAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, testLogFileName)

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	// Verify path is returned before close
	pathBefore := logger.LogPath()
	if pathBefore != logPath {
		t.Errorf("Before Close: "+errMsgLogPathExpectedPath, logPath, pathBefore)
	}

	// Close the logger
	err = logger.Close()
	if err != nil {
		t.Fatalf(errMsgCloseUnexpected, err)
	}

	// Verify empty string is returned after close
	pathAfter := logger.LogPath()
	if pathAfter != "" {
		t.Errorf("After Close: "+errMsgLogPathExpectedEmpty, pathAfter)
	}
}

// TestLogPathConcurrent verifies that concurrent LogPath calls are thread-safe.
func TestLogPathConcurrent(t *testing.T) {
	t.Parallel()

	logger, logPath := createTestLogger(t, false)

	const goroutines = 50
	var wg sync.WaitGroup
	results := make(chan string, goroutines)

	// Launch multiple goroutines that all call LogPath
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- logger.LogPath()
		}()
	}

	wg.Wait()
	close(results)

	// Verify all results are consistent
	for result := range results {
		if result != logPath {
			t.Errorf(errMsgLogPathExpectedPath, logPath, result)
		}
	}
}

// TestLogPathWithPrimaryPath verifies that LogPath returns /var/log path when available.
// This test may skip if the primary path is not writable (e.g., in CI environments or macOS).
func TestLogPathWithPrimaryPath(t *testing.T) {
	logger, err := NewLogger(false)
	if err != nil {
		t.Skipf(errMsgSkipCannotCreateLogger, err)
	}

	t.Cleanup(func() {
		logger.Close() //nolint:errcheck // best-effort cleanup in tests
	})

	result := logger.LogPath()

	// Verify we got one of the expected paths
	if result != defaultLogPath && result != fallbackLogPath {
		t.Errorf("LogPath() expected %q or %q, got %q", defaultLogPath, fallbackLogPath, result)
	}

	// If /var/log is writable (running as root on Linux), it should be the primary path
	if result == defaultLogPath {
		if !strings.HasPrefix(result, "/var/log/") {
			t.Errorf(errMsgLogPathUnexpectedPrefix, "/var/log/", result)
		}
	}
}

// TestLogPathWithFallbackPath verifies that LogPath returns /tmp path as fallback.
func TestLogPathWithFallbackPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an unwritable directory to force fallback
	unwritableDir := filepath.Join(tmpDir, "unwritable")
	//nolint:gosec // G301: intentionally testing unwritable directories
	if err := os.Mkdir(unwritableDir, 0o555); err != nil {
		t.Fatalf(errMsgCreateUnwritableDir, err)
	}

	t.Cleanup(func() {
		os.Chmod(unwritableDir, 0o700) //nolint:errcheck,gosec // G302: directories need execute bit for cleanup
	})

	primaryPath := filepath.Join(unwritableDir, "primary.log")
	fallbackPath := filepath.Join(tmpDir, "fallback.log")

	logger, err := newLoggerWithPaths(false, []string{primaryPath, fallbackPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	t.Cleanup(func() {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	result := logger.LogPath()
	if result != fallbackPath {
		t.Errorf(errMsgLogPathExpectedPath, fallbackPath, result)
	}
}

// TestLogPathZeroValueLogger verifies LogPath on zero-value Logger.
func TestLogPathZeroValueLogger(t *testing.T) {
	var logger Logger

	result := logger.LogPath()
	if result != "" {
		t.Errorf(errMsgLogPathExpectedEmpty, result)
	}
}

// TestLogPathAndLogConcurrent verifies thread safety with concurrent LogPath and Log calls.
func TestLogPathAndLogConcurrent(t *testing.T) {
	t.Parallel()

	logger, logPath := createTestLogger(t, false)

	const goroutines = 30
	var wg sync.WaitGroup

	// Launch goroutines that call LogPath
	// Note: t.Errorf is safe to call from multiple goroutines
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := logger.LogPath()
			if result != logPath {
				t.Errorf(errMsgLogPathExpectedPath, logPath, result)
			}
		}()
	}

	// Launch goroutines that call Log
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Log("Concurrent log %d", id)
		}(i)
	}

	// Test passes if no panic or race condition occurs
	wg.Wait()
}

// TestLogPathAndCloseConcurrent verifies thread safety with concurrent LogPath and Close calls.
func TestLogPathAndCloseConcurrent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, testLogFileName)

	logger, err := newLoggerWithPaths(false, []string{logPath})
	if err != nil {
		t.Fatalf(errMsgUnexpectedError, err)
	}

	const goroutines = 20
	var wg sync.WaitGroup

	// Launch goroutines that call LogPath
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Result can be either the path or empty string (after close)
			_ = logger.LogPath()
		}()
	}

	// Launch goroutines that call Close
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Close() //nolint:errcheck // best-effort cleanup in concurrent test
		}()
	}

	// Test passes if no panic or race condition occurs
	wg.Wait()
}

// ============================================================================
// NewLoggerWithPath Tests
// ============================================================================

// TestNewLoggerWithPathCreatesFile verifies that NewLoggerWithPath creates a file at the specified path.
func TestNewLoggerWithPathCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "custom.log")

	logger, err := NewLoggerWithPath(logPath, false)
	if err != nil {
		t.Fatalf(errMsgNewLoggerWithPathUnexpected, err)
	}

	t.Cleanup(func() {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	// Verify file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Expected log file to be created at specified path")
	}

	// Verify logger is properly initialized
	if logger.file == nil {
		t.Error(errMsgExpectedFileSet)
	}
}

// TestNewLoggerWithPathInvalidPath verifies that NewLoggerWithPath returns error for invalid paths.
func TestNewLoggerWithPathInvalidPath(t *testing.T) {
	// Use a path in a non-existent directory
	invalidPath := "/nonexistent/directory/path/test.log"

	logger, err := NewLoggerWithPath(invalidPath, false)

	if err == nil {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
		t.Fatal("Expected error for invalid path, got nil")
	}

	if logger != nil {
		t.Error(errMsgExpectedLoggerNil)
	}

	// Verify error message contains the path
	if !strings.HasPrefix(err.Error(), errMsgFailedToOpenLogFile) {
		t.Errorf(errMsgExpectedErrorMsgStart, errMsgFailedToOpenLogFile, err.Error())
	}

	if !strings.Contains(err.Error(), invalidPath) {
		t.Errorf("Expected error message to contain path %q, got %q", invalidPath, err.Error())
	}
}

// TestNewLoggerWithPathUnwritablePath verifies that NewLoggerWithPath returns error for unwritable paths.
func TestNewLoggerWithPathUnwritablePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an unwritable directory
	unwritableDir := filepath.Join(tmpDir, "unwritable")
	//nolint:gosec // G301: intentionally testing unwritable directories
	if err := os.Mkdir(unwritableDir, 0o555); err != nil {
		t.Fatalf(errMsgCreateUnwritableDir, err)
	}

	t.Cleanup(func() {
		os.Chmod(unwritableDir, 0o700) //nolint:errcheck,gosec // G302: directories need execute bit for cleanup
	})

	unwritablePath := filepath.Join(unwritableDir, "test.log")

	logger, err := NewLoggerWithPath(unwritablePath, false)

	if err == nil {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
		t.Fatal("Expected error for unwritable path, got nil")
	}

	if logger != nil {
		t.Error(errMsgExpectedLoggerNil)
	}

	// Verify error message format
	if !strings.HasPrefix(err.Error(), errMsgFailedToOpenLogFile) {
		t.Errorf(errMsgExpectedErrorMsgStart, errMsgFailedToOpenLogFile, err.Error())
	}
}

// TestNewLoggerWithPathVerboseFlagTrue verifies that verbose flag is set correctly when true.
func TestNewLoggerWithPathVerboseFlagTrue(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "verbose-true.log")

	logger, err := NewLoggerWithPath(logPath, true)
	if err != nil {
		t.Fatalf(errMsgNewLoggerWithPathUnexpected, err)
	}

	t.Cleanup(func() {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	if !logger.verbose {
		t.Error(errMsgExpectedVerboseTrue)
	}
}

// TestNewLoggerWithPathVerboseFlagFalse verifies that verbose flag is set correctly when false.
func TestNewLoggerWithPathVerboseFlagFalse(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "verbose-false.log")

	logger, err := NewLoggerWithPath(logPath, false)
	if err != nil {
		t.Fatalf(errMsgNewLoggerWithPathUnexpected, err)
	}

	t.Cleanup(func() {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	if logger.verbose {
		t.Error(errMsgExpectedVerboseFalse)
	}
}

// TestNewLoggerWithPathFilePermissions verifies that created files have 0644 permissions.
func TestNewLoggerWithPathFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "permissions.log")

	logger, err := NewLoggerWithPath(logPath, false)
	if err != nil {
		t.Fatalf(errMsgNewLoggerWithPathUnexpected, err)
	}

	t.Cleanup(func() {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	// Check file permissions (0644 = rw-r--r--)
	// Note: On some systems, umask may affect the actual permissions
	perm := info.Mode().Perm()
	expectedPerm := os.FileMode(0o644)

	// Verify no extra permission bits beyond 0o644 are set.
	// The umask might restrict permissions further (e.g., 0o600), which is acceptable.
	if perm&0o644 != perm {
		t.Errorf("Expected file permissions 0644 or more restrictive, got %04o", perm)
	}

	// On systems without restrictive umask, permissions should be exactly 0644
	// We log this for debugging but don't fail the test due to umask variations
	if perm != expectedPerm {
		t.Logf("Note: File permissions are %04o (expected %04o, may be affected by umask)", perm, expectedPerm)
	}
}

// TestNewLoggerWithPathIntegrationWithLog verifies that Log works correctly with NewLoggerWithPath.
func TestNewLoggerWithPathIntegrationWithLog(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "integration-log.log")

	logger, err := NewLoggerWithPath(logPath, false)
	if err != nil {
		t.Fatalf(errMsgNewLoggerWithPathUnexpected, err)
	}

	t.Cleanup(func() {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	// Write log messages
	testMessages := []string{
		"First test message",
		"Second test message with value: 42",
		"Third test message",
	}

	for _, msg := range testMessages {
		logger.Log("%s", msg)
	}

	// Sync file to ensure content is flushed
	if err := logger.file.Sync(); err != nil {
		t.Fatalf(errMsgSyncLogFileFailed, err)
	}

	// Read file and verify messages
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	for _, msg := range testMessages {
		if !strings.Contains(string(content), msg) {
			t.Errorf(errMsgMessageNotFound, msg)
		}
	}

	// Verify timestamp format
	if !rfc3339Pattern.MatchString(string(content)) {
		t.Error(errMsgTimestampNotMatched)
	}
}

// TestNewLoggerWithPathIntegrationWithLogPath verifies that LogPath returns the correct path.
func TestNewLoggerWithPathIntegrationWithLogPath(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "integration-logpath.log")

	logger, err := NewLoggerWithPath(logPath, false)
	if err != nil {
		t.Fatalf(errMsgNewLoggerWithPathUnexpected, err)
	}

	t.Cleanup(func() {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
	})

	result := logger.LogPath()
	if result != logPath {
		t.Errorf(errMsgLogPathExpectedPath, logPath, result)
	}
}

// TestNewLoggerWithPathIntegrationWithClose verifies that Close works correctly.
func TestNewLoggerWithPathIntegrationWithClose(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "integration-close.log")

	logger, err := NewLoggerWithPath(logPath, false)
	if err != nil {
		t.Fatalf(errMsgNewLoggerWithPathUnexpected, err)
	}

	// Write a message before closing
	logger.Log(testMsgBeforeClose)

	// Close should succeed
	err = logger.Close()
	if err != nil {
		t.Errorf(errMsgCloseUnexpected, err)
	}

	// Verify LogPath returns empty string after close
	if path := logger.LogPath(); path != "" {
		t.Errorf("After Close: "+errMsgLogPathExpectedEmpty, path)
	}

	// Verify message was flushed to file
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	if !strings.Contains(string(content), testMsgBeforeClose) {
		t.Error("Expected message to be flushed to file before close")
	}

	// Verify Close is idempotent
	err = logger.Close()
	if err != nil {
		t.Errorf("Second Close() should return nil, got: %v", err)
	}
}

// TestNewLoggerWithPathAppendMode verifies that NewLoggerWithPath opens files in append mode.
func TestNewLoggerWithPathAppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "append-mode.log")

	// Pre-create file with content
	initialContent := "existing content\n"
	if err := os.WriteFile(logPath, []byte(initialContent), 0o644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	logger, err := NewLoggerWithPath(logPath, false)
	if err != nil {
		t.Fatalf(errMsgNewLoggerWithPathUnexpected, err)
	}

	// Write new content
	logger.Log("new message")

	// Close to flush
	err = logger.Close()
	if err != nil {
		t.Fatalf(errMsgCloseUnexpected, err)
	}

	// Read file and verify both contents exist
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errMsgLogFileReadFailed, err)
	}

	if !strings.Contains(string(content), "existing content") {
		t.Error("Expected initial content to be preserved")
	}

	if !strings.Contains(string(content), "new message") {
		t.Error("Expected new message to be appended")
	}
}

// TestNewLoggerWithPathEmptyPath verifies behavior with empty path string.
func TestNewLoggerWithPathEmptyPath(t *testing.T) {
	logger, err := NewLoggerWithPath("", false)

	if err == nil {
		if logger != nil {
			logger.Close() //nolint:errcheck // best-effort cleanup in tests
		}
		t.Fatal("Expected error for empty path, got nil")
	}

	if logger != nil {
		t.Error(errMsgExpectedLoggerNil)
	}
}
