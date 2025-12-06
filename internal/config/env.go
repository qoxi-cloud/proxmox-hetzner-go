// Package config provides configuration structures and utilities for the
// Proxmox VE installer on Hetzner dedicated servers.
//
// # Environment Variable Support
//
// This file provides functions for loading configuration from environment variables.
// Environment variables override file configuration values but are overridden by
// TUI user input.
//
// # Configuration Priority (highest to lowest)
//
//  1. User input in TUI
//  2. Environment variables (PVE_* prefix)
//  3. Config file values
//  4. Default values from DefaultConfig()
//
// # Supported Environment Variables
//
// System Configuration:
//   - PVE_HOSTNAME: Server hostname (RFC 1123 compliant)
//   - PVE_DOMAIN_SUFFIX: Domain suffix (e.g., "local")
//   - PVE_TIMEZONE: Timezone (e.g., "Europe/Kyiv")
//   - PVE_EMAIL: Admin email address
//   - PVE_ROOT_PASSWORD: Root password (sensitive)
//   - PVE_SSH_PUBLIC_KEY: SSH public key (sensitive)
//
// Network Configuration:
//   - INTERFACE_NAME: Primary network interface (e.g., "eth0")
//   - BRIDGE_MODE: VM networking mode (internal, external, both)
//   - PRIVATE_SUBNET: NAT network subnet (e.g., "10.0.0.0/24")
//
// Storage Configuration:
//   - ZFS_RAID: ZFS RAID level (single, raid0, raid1)
//   - DISKS: Comma-separated list of disk devices
//
// Tailscale Configuration:
//   - INSTALL_TAILSCALE: Enable Tailscale (true/false/yes/no/1/0)
//   - TAILSCALE_AUTH_KEY: Tailscale auth key (sensitive)
//   - TAILSCALE_SSH: Enable SSH over Tailscale (true/false)
//   - TAILSCALE_WEBUI: Expose WebUI via Tailscale (true/false)
package config

import (
	"os"
	"strings"
)

// parseBool converts common boolean string representations to bool.
// Accepts: "true", "yes", "1" (case-insensitive) as true.
// All other values return false.
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "yes" || s == "1"
}

// EnvVarSet returns true if the environment variable with the given name
// was explicitly set, even if its value is empty.
// This distinguishes between unset variables and variables set to "".
func EnvVarSet(name string) bool {
	_, exists := os.LookupEnv(name)
	return exists
}

// parseDisksEnv parses a comma-separated list of disk paths from an environment variable.
// It trims whitespace from each element and filters out empty strings.
// Returns nil if no valid disk paths remain after filtering.
func parseDisksEnv(v string) []string {
	disks := strings.Split(v, ",")

	for i := range disks {
		disks[i] = strings.TrimSpace(disks[i])
	}

	filtered := make([]string, 0, len(disks))

	for _, d := range disks {
		if d != "" {
			filtered = append(filtered, d)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	return filtered
}

// LoadFromEnv loads configuration values from environment variables into cfg.
// Only non-empty environment variable values override existing configuration;
// empty or unset variables leave the current values unchanged.
// Sensitive fields (RootPassword, SSHPublicKey, TailscaleAuthKey) are loaded
// from env but are never persisted to configuration files.
func LoadFromEnv(cfg *Config) {
	if cfg == nil {
		return
	}

	loadSystemEnv(cfg)
	loadNetworkEnv(cfg)
	loadStorageEnv(cfg)
	loadTailscaleEnv(cfg)
}

// loadSystemEnv loads system configuration from environment variables.
func loadSystemEnv(cfg *Config) {
	if v := os.Getenv("PVE_HOSTNAME"); v != "" {
		cfg.System.Hostname = v
	}

	if v := os.Getenv("PVE_DOMAIN_SUFFIX"); v != "" {
		cfg.System.DomainSuffix = v
	}

	if v := os.Getenv("PVE_TIMEZONE"); v != "" {
		cfg.System.Timezone = v
	}

	if v := os.Getenv("PVE_EMAIL"); v != "" {
		cfg.System.Email = v
	}

	if v := os.Getenv("PVE_ROOT_PASSWORD"); v != "" {
		cfg.System.RootPassword = v
	}

	if v := os.Getenv("PVE_SSH_PUBLIC_KEY"); v != "" {
		cfg.System.SSHPublicKey = v
	}
}

// loadNetworkEnv loads network configuration from environment variables.
func loadNetworkEnv(cfg *Config) {
	if v := os.Getenv("INTERFACE_NAME"); v != "" {
		cfg.Network.InterfaceName = v
	}

	if v := os.Getenv("BRIDGE_MODE"); v != "" {
		mode := BridgeMode(strings.ToLower(v))
		if mode.IsValid() {
			cfg.Network.BridgeMode = mode
		}
	}

	if v := os.Getenv("PRIVATE_SUBNET"); v != "" {
		cfg.Network.PrivateSubnet = v
	}
}

// loadStorageEnv loads storage configuration from environment variables.
func loadStorageEnv(cfg *Config) {
	if v := os.Getenv("ZFS_RAID"); v != "" {
		raid := ZFSRaid(strings.ToLower(v))
		if raid.IsValid() {
			cfg.Storage.ZFSRaid = raid
		}
	}

	if v := os.Getenv("DISKS"); v != "" {
		if disks := parseDisksEnv(v); disks != nil {
			cfg.Storage.Disks = disks
		}
	}
}

// loadTailscaleEnv loads Tailscale configuration from environment variables.
// Boolean fields use EnvVarSet to distinguish unset from "false".
// TAILSCALE_AUTH_KEY is a sensitive field loaded from env but never persisted.
func loadTailscaleEnv(cfg *Config) {
	if EnvVarSet("INSTALL_TAILSCALE") {
		cfg.Tailscale.Enabled = parseBool(os.Getenv("INSTALL_TAILSCALE"))
	}

	if v := os.Getenv("TAILSCALE_AUTH_KEY"); v != "" {
		cfg.Tailscale.AuthKey = v
	}

	if EnvVarSet("TAILSCALE_SSH") {
		cfg.Tailscale.SSH = parseBool(os.Getenv("TAILSCALE_SSH"))
	}

	if EnvVarSet("TAILSCALE_WEBUI") {
		cfg.Tailscale.WebUI = parseBool(os.Getenv("TAILSCALE_WEBUI"))
	}
}
