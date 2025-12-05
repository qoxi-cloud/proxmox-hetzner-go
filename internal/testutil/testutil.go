// Package testutil provides shared test utilities for the project.
package testutil

import (
	"testing"
)

// TempDir creates a temporary directory for tests and returns its path.
// The directory is automatically cleaned up when the test completes.
func TempDir(t *testing.T) string {
	t.Helper()

	return t.TempDir()
}

// SkipIfShort skips the test if the -short flag is set.
// Use this for long-running tests (integration tests, etc.).
func SkipIfShort(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
}
