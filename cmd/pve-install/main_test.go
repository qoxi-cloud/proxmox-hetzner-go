package main

import (
	"bytes"
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

func TestRootCmdHelpOutput(t *testing.T) {
	// Capture help output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Verify help contains expected content
	assert.Contains(t, output, "pve-install")
	assert.Contains(t, output, "TUI-based installer for Proxmox VE")
	assert.Contains(t, output, "--config")
	assert.Contains(t, output, "--save-config")
	assert.Contains(t, output, "--verbose")
	assert.Contains(t, output, "version")

	// Reset args for other tests
	rootCmd.SetArgs(nil)
}

func TestVersionCmdOutput(t *testing.T) {
	// Capture version output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Verify version output contains expected content
	assert.Contains(t, output, "pve-install")
	assert.Contains(t, output, version.Version)

	// Reset args for other tests
	rootCmd.SetArgs(nil)
}
