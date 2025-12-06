package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Test constants for file names to avoid duplication.
const (
	testConfigFileName = "config.yaml"
)

func TestSaveToFileSuccessfulSave(t *testing.T) {
	cfg := DefaultConfig()
	cfg.System.Hostname = "test-save-host"
	cfg.System.Email = "test@example.com"

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Verify content can be read back
	data, err := os.ReadFile(filePath) //nolint:gosec // test file path is controlled
	require.NoError(t, err)

	var restored Config
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "test-save-host", restored.System.Hostname)
	assert.Equal(t, "test@example.com", restored.System.Email)
}

func TestSaveToFileSensitiveFieldsExcluded(t *testing.T) {
	cfg := DefaultConfig()
	cfg.System.RootPassword = "super-secret-password"
	cfg.System.SSHPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG..."
	cfg.Tailscale.AuthKey = "tskey-auth-secret123"

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Read file content
	data, err := os.ReadFile(filePath) //nolint:gosec // test file path is controlled
	require.NoError(t, err)
	content := string(data)

	// Sensitive fields should NOT be in the file
	assert.NotContains(t, content, "super-secret-password")
	assert.NotContains(t, content, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG...")
	assert.NotContains(t, content, "tskey-auth-secret123")
	assert.NotContains(t, content, "root_password")
	assert.NotContains(t, content, "ssh_public_key")
	assert.NotContains(t, content, "auth_key")
}

func TestSaveToFileOriginalConfigUnmodified(t *testing.T) {
	cfg := DefaultConfig()
	cfg.System.RootPassword = "original-password"
	cfg.System.SSHPublicKey = "original-ssh-key"
	cfg.Tailscale.AuthKey = "original-tailscale-key"

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Verify original config is NOT modified
	assert.Equal(t, "original-password", cfg.System.RootPassword)
	assert.Equal(t, "original-ssh-key", cfg.System.SSHPublicKey)
	assert.Equal(t, "original-tailscale-key", cfg.Tailscale.AuthKey)
}

func TestSaveToFileCreatesParentDirectories(t *testing.T) {
	cfg := DefaultConfig()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nested", testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)
}

func TestSaveToFileCreatesNestedDirectories(t *testing.T) {
	cfg := DefaultConfig()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "level1", "level2", "level3", testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Verify all directories were created
	_, err = os.Stat(filepath.Join(tmpDir, "level1", "level2", "level3"))
	require.NoError(t, err)
}

func TestSaveToFileFilePermissions(t *testing.T) {
	cfg := DefaultConfig()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Check file permissions
	info, err := os.Stat(filePath)
	require.NoError(t, err)

	// File should be readable/writable by owner only (0600) for security
	expectedMode := os.FileMode(0o600)
	actualMode := info.Mode().Perm()
	assert.Equal(t, expectedMode, actualMode, "file permissions should be 0600")
}

func TestSaveToFileOverwritesExistingFile(t *testing.T) {
	cfg := DefaultConfig()
	cfg.System.Hostname = "first-hostname"

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, testConfigFileName)

	// First save
	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Modify and save again
	cfg.System.Hostname = "second-hostname"
	err = cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Verify the new content
	data, err := os.ReadFile(filePath) //nolint:gosec // test file path is controlled
	require.NoError(t, err)

	var restored Config
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "second-hostname", restored.System.Hostname)
}

func TestSaveToFileInvalidPath(t *testing.T) {
	cfg := DefaultConfig()

	// Try to save to a path where directory creation should fail
	// Using a null byte in path which is invalid on most systems
	err := cfg.SaveToFile("/\x00/invalid/path/config.yaml")
	require.Error(t, err)
}

func TestSaveToFilePreservesAllNonSensitiveFields(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			Hostname:     "test-hostname",
			DomainSuffix: "example.com",
			Timezone:     "America/New_York",
			Email:        "admin@example.com",
			RootPassword: "secret",
			SSHPublicKey: "ssh-key",
		},
		Network: NetworkConfig{
			InterfaceName: "eth0",
			BridgeMode:    BridgeModeExternal,
			PrivateSubnet: "192.168.1.0/24", // NOSONAR(go:S1313) Class C private range - test data
		},
		Storage: StorageConfig{
			ZFSRaid: ZFSRaid0,
			Disks:   []string{"/dev/sda", "/dev/sdb"},
		},
		Tailscale: TailscaleConfig{
			Enabled: true,
			AuthKey: "tskey-secret",
			SSH:     true,
			WebUI:   true,
		},
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	data, err := os.ReadFile(filePath) //nolint:gosec // test file path is controlled
	require.NoError(t, err)

	var restored Config
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Verify all non-sensitive fields are preserved
	assert.Equal(t, cfg.System.Hostname, restored.System.Hostname)
	assert.Equal(t, cfg.System.DomainSuffix, restored.System.DomainSuffix)
	assert.Equal(t, cfg.System.Timezone, restored.System.Timezone)
	assert.Equal(t, cfg.System.Email, restored.System.Email)

	assert.Equal(t, cfg.Network.InterfaceName, restored.Network.InterfaceName)
	assert.Equal(t, cfg.Network.BridgeMode, restored.Network.BridgeMode)
	assert.Equal(t, cfg.Network.PrivateSubnet, restored.Network.PrivateSubnet)

	assert.Equal(t, cfg.Storage.ZFSRaid, restored.Storage.ZFSRaid)
	assert.Equal(t, cfg.Storage.Disks, restored.Storage.Disks)

	assert.Equal(t, cfg.Tailscale.Enabled, restored.Tailscale.Enabled)
	assert.Equal(t, cfg.Tailscale.SSH, restored.Tailscale.SSH)
	assert.Equal(t, cfg.Tailscale.WebUI, restored.Tailscale.WebUI)

	// Sensitive fields should be empty
	assert.Empty(t, restored.System.RootPassword)
	assert.Empty(t, restored.System.SSHPublicKey)
	assert.Empty(t, restored.Tailscale.AuthKey)
}

func TestSaveToFileValidYAMLOutput(t *testing.T) {
	cfg := DefaultConfig()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	data, err := os.ReadFile(filePath) //nolint:gosec // test file path is controlled
	require.NoError(t, err)

	content := string(data)

	// Verify YAML structure
	assert.Contains(t, content, "system:")
	assert.Contains(t, content, "network:")
	assert.Contains(t, content, "storage:")
	assert.Contains(t, content, "tailscale:")

	// Verify the YAML is parseable
	var parsed map[string]interface{}
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify top-level keys exist
	_, hasSystem := parsed["system"]
	_, hasNetwork := parsed["network"]
	_, hasStorage := parsed["storage"]
	_, hasTailscale := parsed["tailscale"]

	assert.True(t, hasSystem, "YAML should have 'system' key")
	assert.True(t, hasNetwork, "YAML should have 'network' key")
	assert.True(t, hasStorage, "YAML should have 'storage' key")
	assert.True(t, hasTailscale, "YAML should have 'tailscale' key")
}

func TestSaveToFileDirectoryPermissions(t *testing.T) {
	cfg := DefaultConfig()

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "created_dir")
	filePath := filepath.Join(nestedDir, testConfigFileName)

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Check directory permissions
	info, err := os.Stat(nestedDir)
	require.NoError(t, err)

	// Directory should be 0750 for security
	expectedMode := os.FileMode(0o750)
	actualMode := info.Mode().Perm()
	assert.Equal(t, expectedMode, actualMode, "directory permissions should be 0750")
}

func TestSaveToFileEmptyConfig(t *testing.T) {
	cfg := &Config{}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty-config.yaml")

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	// Verify file exists and is valid YAML
	data, err := os.ReadFile(filePath) //nolint:gosec // test file path is controlled
	require.NoError(t, err)

	var restored Config
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)
}

func TestSaveToFileWithSpecialCharactersInPath(t *testing.T) {
	cfg := DefaultConfig()

	tmpDir := t.TempDir()
	// Path with spaces and other valid but unusual characters
	filePath := filepath.Join(tmpDir, "path with spaces", "config-file.yaml")

	err := cfg.SaveToFile(filePath)
	require.NoError(t, err)

	_, err = os.Stat(filePath)
	require.NoError(t, err)
}

func TestSaveToFileNilReceiver(t *testing.T) {
	var cfg *Config

	err := cfg.SaveToFile("/tmp/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}
