// Package config provides configuration structures and utilities for the
// Proxmox VE installer on Hetzner dedicated servers.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SaveToFile saves the configuration to a YAML file at the specified path.
// Sensitive fields (RootPassword, SSHPublicKey, AuthKey) are excluded from the output.
// Parent directories are created automatically with 0750 permissions.
// The file is written with 0600 permissions for security.
// The original Config instance is not modified.
func (c *Config) SaveToFile(path string) error {
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
