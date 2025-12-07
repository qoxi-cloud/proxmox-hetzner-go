package installer_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qoxi-cloud/proxmox-hetzner-go/internal/config"
	"github.com/qoxi-cloud/proxmox-hetzner-go/internal/installer"
)

// These integration tests verify that the Logger correctly integrates with
// Config.Verbose, allowing the CLI verbose flag to control Logger output.
//
// Example integration in main.go or installer orchestration:
//
//	cfg := config.DefaultConfig()
//	cfg.Verbose = verbose // from CLI flag
//
//	logger, err := installer.NewLogger(cfg.Verbose)
//	if err != nil {
//	    return fmt.Errorf("failed to create logger: %w", err)
//	}
//	defer logger.Close()
//
//	logger.Log("Log file: %s", logger.LogPath())

// Test message and error constants to avoid string literal duplication.
// These constants are used across multiple test functions.
const (
	testIntegrationMessage = "Integration test message"

	// Error format strings used in t.Fatalf calls.
	errNewLoggerWithPath = "NewLoggerWithPath() returned unexpected error: %v"
	errCreatePipe        = "Failed to create pipe: %v"
	errReadPipe          = "Failed to read from pipe: %v"
	errReadLogFile       = "Failed to read log file: %v"
)

// captureStdout captures stdout output during the execution of the provided function.
// It returns the captured output as a string. This helper reduces code duplication
// in tests that need to verify verbose logging output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(errCreatePipe, err)
	}

	os.Stdout = w

	fn()

	//nolint:errcheck // best-effort in tests - pipe close for capture
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf(errReadPipe, err)
	}

	return buf.String()
}

// TestLoggerWithConfigVerboseTrue verifies that Logger produces stdout output
// when created with Config.Verbose=true.
//
// This test demonstrates the integration pattern between Config and Logger:
// the Config.Verbose field (set from CLI --verbose flag) controls whether
// log messages are echoed to stdout in addition to being written to the log file.
func TestLoggerWithConfigVerboseTrue(t *testing.T) {
	// Create config with Verbose enabled (simulating --verbose CLI flag)
	cfg := config.DefaultConfig()
	cfg.Verbose = true

	// Create logger using config's verbose setting with custom path for testing
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "verbose-integration.log")

	logger, err := installer.NewLoggerWithPath(logPath, cfg.Verbose)
	if err != nil {
		t.Fatalf(errNewLoggerWithPath, err)
	}

	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup in tests
		logger.Close()
	})

	// Capture stdout and log a message - with Verbose=true, this should appear on stdout
	capturedOutput := captureStdout(t, func() {
		logger.Log(testIntegrationMessage)
	})

	// Verify message was written to stdout (verbose mode)
	if !strings.Contains(capturedOutput, testIntegrationMessage) {
		t.Errorf("With Config.Verbose=true, expected stdout to contain %q, got %q",
			testIntegrationMessage, capturedOutput)
	}

	// Also verify message was written to log file
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errReadLogFile, err)
	}

	if !strings.Contains(string(content), testIntegrationMessage) {
		t.Errorf("Expected log file to contain %q", testIntegrationMessage)
	}
}

// TestLoggerWithConfigVerboseFalse verifies that Logger does NOT produce stdout
// output when created with Config.Verbose=false (the default).
//
// This is the default behavior - logs are written only to the log file,
// keeping the terminal clean during normal operation.
func TestLoggerWithConfigVerboseFalse(t *testing.T) {
	// Create config with Verbose disabled (default behavior)
	cfg := config.DefaultConfig()
	// cfg.Verbose is already false by default, but explicitly set for clarity
	cfg.Verbose = false

	// Create logger using config's verbose setting with custom path for testing
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "quiet-integration.log")

	logger, err := installer.NewLoggerWithPath(logPath, cfg.Verbose)
	if err != nil {
		t.Fatalf(errNewLoggerWithPath, err)
	}

	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup in tests
		logger.Close()
	})

	// Capture stdout and log a message - with Verbose=false, this should NOT appear on stdout
	capturedOutput := captureStdout(t, func() {
		logger.Log(testIntegrationMessage)
	})

	// Verify NO output was written to stdout (quiet mode)
	if capturedOutput != "" {
		t.Errorf("With Config.Verbose=false, expected no stdout output, got %q", capturedOutput)
	}

	// Verify message WAS written to log file (logging still works)
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errReadLogFile, err)
	}

	if !strings.Contains(string(content), testIntegrationMessage) {
		t.Errorf("Expected log file to contain %q even with Verbose=false", testIntegrationMessage)
	}
}

// TestLoggerConfigVerboseDefaultIsFalse verifies that Config.Verbose defaults to false.
//
// This ensures that by default (without --verbose flag), the Logger operates
// in quiet mode - logs are written to file but not echoed to stdout.
func TestLoggerConfigVerboseDefaultIsFalse(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.Verbose {
		t.Error("Expected Config.Verbose to default to false")
	}
}

// TestLoggerIntegrationWithNewLogger demonstrates integration using NewLogger
// (with default paths) when running in environments where log paths are writable.
//
// This test may skip if default log paths are not writable (e.g., macOS development
// without root, or CI environments with restricted permissions).
func TestLoggerIntegrationWithNewLogger(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Verbose = true

	// Try to create logger with default paths
	logger, err := installer.NewLogger(cfg.Verbose)
	if err != nil {
		t.Skipf("Skipping: cannot create logger with default paths: %v", err)
	}

	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup in tests
		logger.Close()
	})

	// Verify logger was created with correct verbose setting
	// We can verify this by checking that LogPath returns a valid path
	logPath := logger.LogPath()
	if logPath == "" {
		t.Error("Expected LogPath() to return a valid path")
	}

	// Verify Verbose=true behavior by capturing stdout and checking log output
	testMessage := "Integration test with NewLogger"
	capturedOutput := captureStdout(t, func() {
		logger.Log("%s", testMessage)
	})

	// With Verbose=true, the message should appear on stdout
	if !strings.Contains(capturedOutput, testMessage) {
		t.Errorf("With Verbose=true, expected stdout to contain %q, got %q",
			testMessage, capturedOutput)
	}

	t.Logf("Logger successfully created with Config.Verbose=%v, log path: %s",
		cfg.Verbose, logPath)
}

// TestLoggerMultipleMessagesWithConfigVerbose verifies that multiple log messages
// respect the verbose setting consistently.
func TestLoggerMultipleMessagesWithConfigVerbose(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Verbose = true

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "multi-message.log")

	logger, err := installer.NewLoggerWithPath(logPath, cfg.Verbose)
	if err != nil {
		t.Fatalf(errNewLoggerWithPath, err)
	}

	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup in tests
		logger.Close()
	})

	// Log multiple messages
	messages := []string{
		"Starting installation",
		"Detecting hardware",
		"Configuring network",
		"Installation complete",
	}

	// Capture stdout while logging all messages
	capturedOutput := captureStdout(t, func() {
		for _, msg := range messages {
			logger.Log("%s", msg)
		}
	})

	// Verify all messages appear in stdout
	for _, msg := range messages {
		if !strings.Contains(capturedOutput, msg) {
			t.Errorf("Expected stdout to contain %q with Verbose=true", msg)
		}
	}

	// Verify all messages appear in log file
	//nolint:gosec // G304: test file path from t.TempDir()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(errReadLogFile, err)
	}

	for _, msg := range messages {
		if !strings.Contains(string(content), msg) {
			t.Errorf("Expected log file to contain %q", msg)
		}
	}
}

// TestLoggerWithConfigVerboseDisplaysLogPath demonstrates a common pattern:
// when verbose mode is enabled, the log path is typically displayed to the user.
func TestLoggerWithConfigVerboseDisplaysLogPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Verbose = true

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "display-path.log")

	logger, err := installer.NewLoggerWithPath(logPath, cfg.Verbose)
	if err != nil {
		t.Fatalf(errNewLoggerWithPath, err)
	}

	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup in tests
		logger.Close()
	})

	// Common pattern: log the log file path when in verbose mode
	actualLogPath := logger.LogPath()

	// Capture stdout while logging the path
	capturedOutput := captureStdout(t, func() {
		logger.Log("Log file: %s", actualLogPath)
	})

	// Verify the log path is displayed in verbose mode
	if !strings.Contains(capturedOutput, logPath) {
		t.Errorf("Expected stdout to contain log path %q, got %q", logPath, capturedOutput)
	}
}
