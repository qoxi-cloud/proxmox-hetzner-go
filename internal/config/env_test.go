package config

import (
	"os"
	"testing"
)

func TestParseBool(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// True values - "true" variations
		{"true lowercase", "true", true},
		{"true mixed case", "True", true},
		{"true uppercase", "TRUE", true},

		// True values - "yes" variations
		{"yes lowercase", "yes", true},
		{"yes mixed case", "Yes", true},
		{"yes uppercase", "YES", true},

		// True values - "1"
		{"one", "1", true},

		// False values - "false" variations
		{"false lowercase", "false", false},
		{"false mixed case", "False", false},
		{"false uppercase", "FALSE", false},

		// False values - "no" variations
		{"no lowercase", "no", false},
		{"no mixed case", "No", false},
		{"no uppercase", "NO", false},

		// False values - "0"
		{"zero", "0", false},

		// Empty and whitespace
		{"empty string", "", false},
		{"whitespace only", "   ", false},

		// Whitespace variations with valid values
		{"true with leading space", " true", true},
		{"true with trailing space", "true ", true},
		{"true with surrounding spaces", " true ", true},
		{"true with tab", "\ttrue", true},
		{"true with newline", "true\n", true},
		{"true with mixed whitespace", " \ttrue\n ", true},

		// Invalid values
		{"maybe", "maybe", false},
		{"two", "2", false},
		{"on", "on", false},
		{"off", "off", false},
		{"enabled", "enabled", false},
		{"disabled", "disabled", false},
		{"y", "y", false},
		{"n", "n", false},
		{"random string", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBool(tt.input)
			if got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnvVarSet(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		setValue *string // nil means unset, empty string means set to ""
		want     bool
	}{
		{
			name:     "variable set with value",
			envName:  "TEST_VAR_WITH_VALUE",
			setValue: ptrString("somevalue"),
			want:     true,
		},
		{
			name:     "variable set to empty string",
			envName:  "TEST_VAR_EMPTY",
			setValue: ptrString(""),
			want:     true,
		},
		{
			name:     "variable not set",
			envName:  "TEST_VAR_UNSET",
			setValue: nil,
			want:     false,
		},
		{
			name:     "variable with underscores",
			envName:  "TEST_VAR_WITH_UNDERSCORES",
			setValue: ptrString("value"),
			want:     true,
		},
		{
			name:     "variable with numbers",
			envName:  "TEST_VAR_123",
			setValue: ptrString("value"),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure the variable is unset before the test
			if err := os.Unsetenv(tt.envName); err != nil {
				t.Fatalf("failed to unset env var %q: %v", tt.envName, err)
			}

			if tt.setValue != nil {
				t.Setenv(tt.envName, *tt.setValue)
			}

			got := EnvVarSet(tt.envName)
			if got != tt.want {
				t.Errorf("EnvVarSet(%q) = %v, want %v", tt.envName, got, tt.want)
			}
		})
	}
}

// ptrString is a helper to create a pointer to a string.
func ptrString(s string) *string {
	return &s
}

func TestLoadFromEnvNilConfig(t *testing.T) {
	t.Helper()
	// Should not panic when called with nil config
	LoadFromEnv(nil)
}

func TestLoadFromEnvHostname(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.System.Hostname

	t.Setenv("PVE_HOSTNAME", "test-server")
	LoadFromEnv(cfg)

	if cfg.System.Hostname != "test-server" {
		t.Errorf("Hostname = %q, want %q", cfg.System.Hostname, "test-server")
	}

	// Verify original was different
	if original == "test-server" {
		t.Error("Default hostname should not be 'test-server'")
	}
}

func TestLoadFromEnvDomainSuffix(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_DOMAIN_SUFFIX", "example.com")
	LoadFromEnv(cfg)

	if cfg.System.DomainSuffix != "example.com" {
		t.Errorf("DomainSuffix = %q, want %q", cfg.System.DomainSuffix, "example.com")
	}
}

func TestLoadFromEnvTimezone(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_TIMEZONE", "America/New_York")
	LoadFromEnv(cfg)

	if cfg.System.Timezone != "America/New_York" {
		t.Errorf("Timezone = %q, want %q", cfg.System.Timezone, "America/New_York")
	}
}

func TestLoadFromEnvEmail(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_EMAIL", "test@example.com")
	LoadFromEnv(cfg)

	if cfg.System.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", cfg.System.Email, "test@example.com")
	}
}

func TestLoadFromEnvRootPassword(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_ROOT_PASSWORD", "supersecret")
	LoadFromEnv(cfg)

	if cfg.System.RootPassword != "supersecret" {
		t.Errorf("RootPassword = %q, want %q", cfg.System.RootPassword, "supersecret")
	}
}

func TestLoadFromEnvSSHPublicKey(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_SSH_PUBLIC_KEY", "ssh-ed25519 AAAA... test@example.com")
	LoadFromEnv(cfg)

	if cfg.System.SSHPublicKey != "ssh-ed25519 AAAA... test@example.com" {
		t.Errorf("SSHPublicKey = %q, want %q", cfg.System.SSHPublicKey, "ssh-ed25519 AAAA... test@example.com")
	}
}

func TestLoadFromEnvEmptyDoesNotOverride(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.System.Hostname

	// Set env var to empty string - should NOT override
	t.Setenv("PVE_HOSTNAME", "")
	LoadFromEnv(cfg)

	if cfg.System.Hostname != original {
		t.Errorf("Empty env var overrode Hostname: got %q, want %q", cfg.System.Hostname, original)
	}
}

func TestLoadFromEnvMultipleFields(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_HOSTNAME", "multi-test")
	t.Setenv("PVE_DOMAIN_SUFFIX", "test.local")
	t.Setenv("PVE_TIMEZONE", "UTC")
	t.Setenv("PVE_EMAIL", "admin@test.local")
	t.Setenv("PVE_ROOT_PASSWORD", "secret123")
	t.Setenv("PVE_SSH_PUBLIC_KEY", "ssh-rsa AAAAB3...")

	LoadFromEnv(cfg)

	if cfg.System.Hostname != "multi-test" {
		t.Errorf("Hostname = %q, want %q", cfg.System.Hostname, "multi-test")
	}
	if cfg.System.DomainSuffix != "test.local" {
		t.Errorf("DomainSuffix = %q, want %q", cfg.System.DomainSuffix, "test.local")
	}
	if cfg.System.Timezone != "UTC" {
		t.Errorf("Timezone = %q, want %q", cfg.System.Timezone, "UTC")
	}
	if cfg.System.Email != "admin@test.local" {
		t.Errorf("Email = %q, want %q", cfg.System.Email, "admin@test.local")
	}
	if cfg.System.RootPassword != "secret123" {
		t.Errorf("RootPassword = %q, want %q", cfg.System.RootPassword, "secret123")
	}
	if cfg.System.SSHPublicKey != "ssh-rsa AAAAB3..." {
		t.Errorf("SSHPublicKey = %q, want %q", cfg.System.SSHPublicKey, "ssh-rsa AAAAB3...")
	}
}

func TestLoadFromEnvPartialOverride(t *testing.T) {
	cfg := DefaultConfig()
	originalDomain := cfg.System.DomainSuffix
	originalTimezone := cfg.System.Timezone

	// Only override hostname, leave others unchanged
	t.Setenv("PVE_HOSTNAME", "partial-test")
	LoadFromEnv(cfg)

	if cfg.System.Hostname != "partial-test" {
		t.Errorf("Hostname = %q, want %q", cfg.System.Hostname, "partial-test")
	}
	if cfg.System.DomainSuffix != originalDomain {
		t.Errorf("DomainSuffix changed unexpectedly: got %q, want %q", cfg.System.DomainSuffix, originalDomain)
	}
	if cfg.System.Timezone != originalTimezone {
		t.Errorf("Timezone changed unexpectedly: got %q, want %q", cfg.System.Timezone, originalTimezone)
	}
}

func TestLoadFromEnvModifiesOriginalConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfgPtr := cfg // Keep pointer to verify same instance is modified

	t.Setenv("PVE_HOSTNAME", "modify-test")
	LoadFromEnv(cfg)

	// Verify the same config instance was modified
	if cfgPtr.System.Hostname != "modify-test" {
		t.Errorf("Original config pointer not modified: got %q, want %q", cfgPtr.System.Hostname, "modify-test")
	}
}

func TestLoadFromEnvPreservesNonSystemFields(t *testing.T) {
	cfg := DefaultConfig()
	originalBridgeMode := cfg.Network.BridgeMode
	originalZFSRaid := cfg.Storage.ZFSRaid
	originalTailscaleEnabled := cfg.Tailscale.Enabled

	t.Setenv("PVE_HOSTNAME", "preserve-test")
	LoadFromEnv(cfg)

	// Verify non-system fields are untouched
	if cfg.Network.BridgeMode != originalBridgeMode {
		t.Errorf("Network.BridgeMode changed unexpectedly: got %v, want %v", cfg.Network.BridgeMode, originalBridgeMode)
	}
	if cfg.Storage.ZFSRaid != originalZFSRaid {
		t.Errorf("Storage.ZFSRaid changed unexpectedly: got %v, want %v", cfg.Storage.ZFSRaid, originalZFSRaid)
	}
	if cfg.Tailscale.Enabled != originalTailscaleEnabled {
		t.Errorf("Tailscale.Enabled changed unexpectedly: got %v, want %v", cfg.Tailscale.Enabled, originalTailscaleEnabled)
	}
}
