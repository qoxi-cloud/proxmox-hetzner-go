package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTempDir(t *testing.T) {
	dir := TempDir(t)

	// Verify directory exists
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify we can write to the directory
	testFile := filepath.Join(dir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(testFile)
	require.NoError(t, err)
}

func TestTempDirCleanup(t *testing.T) {
	var tempDir string

	// Run in a subtest to trigger cleanup
	t.Run("create temp dir", func(t *testing.T) {
		tempDir = TempDir(t)
		require.DirExists(t, tempDir)
	})

	// After subtest completes, directory should be cleaned up
	_, err := os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err), "temp directory should be removed after test cleanup")
}

func TestSkipIfShort(t *testing.T) {
	// This test verifies SkipIfShort doesn't panic
	// When running with -short, this test will be skipped
	if testing.Short() {
		t.Run("should skip", func(t *testing.T) {
			SkipIfShort(t)
			t.Error("this should not be reached in short mode")
		})
	} else {
		// When not in short mode, verify the function doesn't skip
		t.Run("should not skip", func(t *testing.T) {
			SkipIfShort(t)
			// If we reach here, the test wasn't skipped (correct behavior)
		})
	}
}
