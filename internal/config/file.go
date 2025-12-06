// Package config provides configuration structures and utilities for the
// Proxmox VE installer on Hetzner dedicated servers.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Ensure imports are used (these will be used by subsequent tasks).
var (
	_ = fmt.Errorf
	_ = os.ReadFile
	_ = filepath.Dir
	_ = yaml.Marshal
)
