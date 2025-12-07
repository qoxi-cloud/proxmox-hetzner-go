package installer

import (
	"os"
	"sync"
	"testing"
)

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
