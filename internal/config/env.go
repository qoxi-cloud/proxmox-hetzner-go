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

// LoadFromEnv loads configuration values from environment variables.
// Environment variables override existing values in the Config.
// Only variables that are explicitly set (not empty) override the config.
// Sensitive fields (RootPassword, SSHPublicKey, AuthKey) are loaded but
// never saved to files.
func LoadFromEnv(cfg *Config) {
	if cfg == nil {
		return
	}

	// System configuration
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
