package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qoxi-cloud/proxmox-hetzner-go/pkg/version"
)

func TestVersionCommand(t *testing.T) {
	// Verify version package returns expected values
	assert.Equal(t, "dev", version.Version)
	assert.Equal(t, "none", version.Commit)
	assert.Equal(t, "unknown", version.Date)
}

func TestVersionFull(t *testing.T) {
	full := version.Full()
	assert.Contains(t, full, "dev")
	assert.Contains(t, full, "none")
	assert.Contains(t, full, "unknown")
}

func TestRootCmdExists(t *testing.T) {
	require.NotNil(t, rootCmd)
	assert.Equal(t, "pve-install", rootCmd.Use)
}

func TestVersionCmdExists(t *testing.T) {
	require.NotNil(t, versionCmd)
	assert.Equal(t, "version", versionCmd.Use)
}

func TestFlagsExist(t *testing.T) {
	// Verify config flag exists
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)

	// Verify save-config flag exists
	saveConfigFlag := rootCmd.PersistentFlags().Lookup("save-config")
	require.NotNil(t, saveConfigFlag)
	assert.Equal(t, "s", saveConfigFlag.Shorthand)

	// Verify verbose flag exists
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)
}
