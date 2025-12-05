// Package testutil provides shared test utilities for the project.
package testutil

import (
	"os"
	"testing"
)

// TempDir creates a temporary directory for tests and returns its path.
// The directory is automatically cleaned up when the test completes.
func TempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "pve-install-test-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("failed to remove temp directory %s: %v", dir, err)
		}
	})

	return dir
}

// SkipIfShort skips the test if the -short flag is set.
// Use this for long-running tests (integration tests, etc.).
func SkipIfShort(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
}
