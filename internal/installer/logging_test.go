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

	if logger.verbose != false {
		t.Error("Logger zero value: verbose should be false")
	}

	// Verify mutex is usable (zero value is valid)
	logger.mu.Lock()
	logger.mu.Unlock()
}

// TestLoggerStructFields verifies that Logger struct fields can be set directly.
func TestLoggerStructFields(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "logger-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	logger := Logger{
		file:    tmpFile,
		verbose: true,
	}

	if logger.file != tmpFile {
		t.Error("Logger file field not set correctly")
	}

	if logger.verbose != true {
		t.Error("Logger verbose field not set correctly")
	}
}

// TestLoggerMutexThreadSafety verifies that Logger's mutex provides thread safety.
func TestLoggerMutexThreadSafety(t *testing.T) {
	var logger Logger
	var wg sync.WaitGroup
	const goroutines = 10

	// Launch multiple goroutines that all try to use the mutex
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.mu.Lock()
			// Simulate work
			logger.verbose = !logger.verbose
			logger.mu.Unlock()
		}()
	}

	wg.Wait()
	// Test passes if no race conditions or deadlocks occurred
}

// TestLoggerPointerInstantiation verifies Logger can be created as a pointer.
func TestLoggerPointerInstantiation(t *testing.T) {
	logger := &Logger{
		verbose: true,
	}

	if logger == nil {
		t.Error("Logger pointer should not be nil")
	}

	if logger.verbose != true {
		t.Error("Logger pointer verbose field not set correctly")
	}
}
