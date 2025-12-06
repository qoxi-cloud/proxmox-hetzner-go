// Package config provides configuration structures and utilities for the
// Proxmox VE installer on Hetzner dedicated servers.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads configuration from a YAML file at the specified path.
// It starts with DefaultConfig() values and overlays file contents on top.
// Missing fields in the file retain their default values.
// Returns an error if the file cannot be read or contains invalid YAML.
func LoadFromFile(path string) (*Config, error) {
	// Start with default configuration
	cfg := DefaultConfig()

	// Read the file
	data, err := os.ReadFile(path) //nolint:gosec // path is provided by caller
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s: %w", path, err)
		}

		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Parse YAML and overlay onto defaults
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", path, err)
	}

	return cfg, nil
}

// SaveToFile saves the configuration to a YAML file at the specified path.
// Sensitive fields (RootPassword, SSHPublicKey, AuthKey) are excluded from the output.
// Parent directories are created automatically with 0750 permissions.
// The file is written with 0600 permissions for security.
// The original Config instance is not modified.
func (c *Config) SaveToFile(path string) error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	// Create a safe copy to avoid modifying the original
	safeCopy := *c
	safeCopy.System.RootPassword = ""
	safeCopy.System.SSHPublicKey = ""
	safeCopy.Tailscale.AuthKey = ""

	// Marshal to YAML
	data, err := yaml.Marshal(&safeCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	return nil
}
