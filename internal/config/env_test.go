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

// Test constants for LoadFromEnv tests.
const (
	testHostname      = "test-server"
	testDomain        = "example.com"
	testTimezone      = "America/New_York"
	testEmail         = "test@example.com"
	testSSHKey        = "ssh-ed25519 AAAA... test@example.com"
	testMultiHostname = "multi-test"
	testMultiDomain   = "test.local"
	testMultiEmail    = "admin@test.local"
	testMultiSSHKey   = "ssh-rsa AAAAB3..."
	testPartial       = "partial-test"
	testModify        = "modify-test"

	// Network test constants.
	testInterfaceEth0       = "eth0"
	testInterfaceEnp        = "enp0s31f6"        // NOSONAR(go:S1313) test value
	testPrivateSubnet       = "192.168.100.0/24" // NOSONAR(go:S1313) RFC 1918 test value
	testPrivateSubnetSecond = "172.16.0.0/16"    // NOSONAR(go:S1313) RFC 1918 test value

	// Storage test constants.
	testDiskSda = "/dev/sda"
	testDiskSdb = "/dev/sdb"
	testDiskSdc = "/dev/sdc"

	// Error format strings.
	errFmtHostname      = "Hostname = %q, want %q"
	errFmtInterfaceName = "InterfaceName = %q, want %q"
	errFmtBridgeMode    = "BridgeMode = %q, want %q"
	errFmtPrivateSubnet = "PrivateSubnet = %q, want %q"
	errFmtZFSRaid       = "ZFSRaid = %q, want %q"
)

func TestLoadFromEnvNilConfig(t *testing.T) {
	// Should not panic when called with nil config
	LoadFromEnv(nil)
	// If we reached here without panic, the test passes
	t.Log("LoadFromEnv(nil) completed without panic")
}

func TestLoadFromEnvHostname(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.System.Hostname

	t.Setenv("PVE_HOSTNAME", testHostname)
	LoadFromEnv(cfg)

	if cfg.System.Hostname != testHostname {
		t.Errorf(errFmtHostname, cfg.System.Hostname, testHostname)
	}

	// Verify original was different
	if original == testHostname {
		t.Error("Default hostname should not be 'test-server'")
	}
}

func TestLoadFromEnvDomainSuffix(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_DOMAIN_SUFFIX", testDomain)
	LoadFromEnv(cfg)

	if cfg.System.DomainSuffix != testDomain {
		t.Errorf("DomainSuffix = %q, want %q", cfg.System.DomainSuffix, testDomain)
	}
}

func TestLoadFromEnvTimezone(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_TIMEZONE", testTimezone)
	LoadFromEnv(cfg)

	if cfg.System.Timezone != testTimezone {
		t.Errorf("Timezone = %q, want %q", cfg.System.Timezone, testTimezone)
	}
}

func TestLoadFromEnvEmail(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_EMAIL", testEmail)
	LoadFromEnv(cfg)

	if cfg.System.Email != testEmail {
		t.Errorf("Email = %q, want %q", cfg.System.Email, testEmail)
	}
}

func TestLoadFromEnvRootPassword(t *testing.T) {
	cfg := DefaultConfig()
	testValue := "supersecret" // NOSONAR(go:S2068) test value, not a real credential

	t.Setenv("PVE_ROOT_PASSWORD", testValue)
	LoadFromEnv(cfg)

	if cfg.System.RootPassword != testValue {
		t.Errorf("RootPassword = %q, want %q", cfg.System.RootPassword, testValue)
	}
}

func TestLoadFromEnvSSHPublicKey(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PVE_SSH_PUBLIC_KEY", testSSHKey)
	LoadFromEnv(cfg)

	if cfg.System.SSHPublicKey != testSSHKey {
		t.Errorf("SSHPublicKey = %q, want %q", cfg.System.SSHPublicKey, testSSHKey)
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
	testRootValue := "secret123" // NOSONAR(go:S2068) test value, not a real credential

	t.Setenv("PVE_HOSTNAME", testMultiHostname)
	t.Setenv("PVE_DOMAIN_SUFFIX", testMultiDomain)
	t.Setenv("PVE_TIMEZONE", "UTC")
	t.Setenv("PVE_EMAIL", testMultiEmail)
	t.Setenv("PVE_ROOT_PASSWORD", testRootValue)
	t.Setenv("PVE_SSH_PUBLIC_KEY", testMultiSSHKey)

	LoadFromEnv(cfg)

	if cfg.System.Hostname != testMultiHostname {
		t.Errorf(errFmtHostname, cfg.System.Hostname, testMultiHostname)
	}

	if cfg.System.DomainSuffix != testMultiDomain {
		t.Errorf("DomainSuffix = %q, want %q", cfg.System.DomainSuffix, testMultiDomain)
	}

	if cfg.System.Timezone != "UTC" {
		t.Errorf("Timezone = %q, want %q", cfg.System.Timezone, "UTC")
	}

	if cfg.System.Email != testMultiEmail {
		t.Errorf("Email = %q, want %q", cfg.System.Email, testMultiEmail)
	}

	if cfg.System.RootPassword != testRootValue {
		t.Errorf("RootPassword = %q, want %q", cfg.System.RootPassword, testRootValue)
	}

	if cfg.System.SSHPublicKey != testMultiSSHKey {
		t.Errorf("SSHPublicKey = %q, want %q", cfg.System.SSHPublicKey, testMultiSSHKey)
	}
}

func TestLoadFromEnvPartialOverride(t *testing.T) {
	cfg := DefaultConfig()
	originalDomain := cfg.System.DomainSuffix
	originalTimezone := cfg.System.Timezone

	// Only override hostname, leave others unchanged
	t.Setenv("PVE_HOSTNAME", testPartial)
	LoadFromEnv(cfg)

	if cfg.System.Hostname != testPartial {
		t.Errorf(errFmtHostname, cfg.System.Hostname, testPartial)
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

	t.Setenv("PVE_HOSTNAME", testModify)
	LoadFromEnv(cfg)

	// Verify the same config instance was modified
	if cfgPtr.System.Hostname != testModify {
		t.Errorf("Original config pointer not modified: got %q, want %q", cfgPtr.System.Hostname, testModify)
	}
}

func TestLoadFromEnvPreservesNonEnvFields(t *testing.T) {
	cfg := DefaultConfig()
	originalZFSRaid := cfg.Storage.ZFSRaid
	originalTailscaleEnabled := cfg.Tailscale.Enabled
	testPreserve := "preserve-test"

	t.Setenv("PVE_HOSTNAME", testPreserve)
	LoadFromEnv(cfg)

	// Verify non-env fields are untouched
	if cfg.Storage.ZFSRaid != originalZFSRaid {
		t.Errorf("Storage.ZFSRaid changed unexpectedly: got %v, want %v", cfg.Storage.ZFSRaid, originalZFSRaid)
	}

	if cfg.Tailscale.Enabled != originalTailscaleEnabled {
		t.Errorf("Tailscale.Enabled changed unexpectedly: got %v, want %v", cfg.Tailscale.Enabled, originalTailscaleEnabled)
	}
}

// Network configuration tests

func TestLoadFromEnvInterfaceNameEth0(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("INTERFACE_NAME", testInterfaceEth0)
	LoadFromEnv(cfg)

	if cfg.Network.InterfaceName != testInterfaceEth0 {
		t.Errorf(errFmtInterfaceName, cfg.Network.InterfaceName, testInterfaceEth0)
	}
}

func TestLoadFromEnvInterfaceNameEnp(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("INTERFACE_NAME", testInterfaceEnp)
	LoadFromEnv(cfg)

	if cfg.Network.InterfaceName != testInterfaceEnp {
		t.Errorf(errFmtInterfaceName, cfg.Network.InterfaceName, testInterfaceEnp)
	}
}

func TestLoadFromEnvBridgeModeInternal(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Network.BridgeMode = BridgeModeExternal // Set to different value first

	t.Setenv("BRIDGE_MODE", "internal")
	LoadFromEnv(cfg)

	if cfg.Network.BridgeMode != BridgeModeInternal {
		t.Errorf(errFmtBridgeMode, cfg.Network.BridgeMode, BridgeModeInternal)
	}
}

func TestLoadFromEnvBridgeModeExternal(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("BRIDGE_MODE", "external")
	LoadFromEnv(cfg)

	if cfg.Network.BridgeMode != BridgeModeExternal {
		t.Errorf(errFmtBridgeMode, cfg.Network.BridgeMode, BridgeModeExternal)
	}
}

func TestLoadFromEnvBridgeModeBoth(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("BRIDGE_MODE", "both")
	LoadFromEnv(cfg)

	if cfg.Network.BridgeMode != BridgeModeBoth {
		t.Errorf(errFmtBridgeMode, cfg.Network.BridgeMode, BridgeModeBoth)
	}
}

func TestLoadFromEnvBridgeModeCaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  BridgeMode
	}{
		{"uppercase INTERNAL", "INTERNAL", BridgeModeInternal},
		{"mixed case Internal", "Internal", BridgeModeInternal},
		{"uppercase EXTERNAL", "EXTERNAL", BridgeModeExternal},
		{"mixed case External", "External", BridgeModeExternal},
		{"uppercase BOTH", "BOTH", BridgeModeBoth},
		{"mixed case Both", "Both", BridgeModeBoth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Network.BridgeMode = "" // Clear default

			t.Setenv("BRIDGE_MODE", tt.input)
			LoadFromEnv(cfg)

			if cfg.Network.BridgeMode != tt.want {
				t.Errorf(errFmtBridgeMode, cfg.Network.BridgeMode, tt.want)
			}
		})
	}
}

func TestLoadFromEnvBridgeModeInvalidKeepsOriginal(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Network.BridgeMode

	// Set to an invalid value - should NOT change the config
	t.Setenv("BRIDGE_MODE", "invalid")
	LoadFromEnv(cfg)

	if cfg.Network.BridgeMode != original {
		t.Errorf("Invalid BRIDGE_MODE changed config: got %q, want %q", cfg.Network.BridgeMode, original)
	}
}

func TestLoadFromEnvBridgeModeEmptyKeepsOriginal(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Network.BridgeMode

	// Set to empty value - should NOT change the config
	t.Setenv("BRIDGE_MODE", "")
	LoadFromEnv(cfg)

	if cfg.Network.BridgeMode != original {
		t.Errorf("Empty BRIDGE_MODE changed config: got %q, want %q", cfg.Network.BridgeMode, original)
	}
}

func TestLoadFromEnvPrivateSubnet(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("PRIVATE_SUBNET", testPrivateSubnet)
	LoadFromEnv(cfg)

	if cfg.Network.PrivateSubnet != testPrivateSubnet {
		t.Errorf(errFmtPrivateSubnet, cfg.Network.PrivateSubnet, testPrivateSubnet)
	}
}

func TestLoadFromEnvPrivateSubnetEmptyKeepsOriginal(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Network.PrivateSubnet

	t.Setenv("PRIVATE_SUBNET", "")
	LoadFromEnv(cfg)

	if cfg.Network.PrivateSubnet != original {
		t.Errorf("Empty PRIVATE_SUBNET changed config: got %q, want %q", cfg.Network.PrivateSubnet, original)
	}
}

func TestLoadFromEnvNetworkMultipleFields(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("INTERFACE_NAME", testInterfaceEth0)
	t.Setenv("BRIDGE_MODE", "external")
	t.Setenv("PRIVATE_SUBNET", testPrivateSubnetSecond)

	LoadFromEnv(cfg)

	if cfg.Network.InterfaceName != testInterfaceEth0 {
		t.Errorf(errFmtInterfaceName, cfg.Network.InterfaceName, testInterfaceEth0)
	}

	if cfg.Network.BridgeMode != BridgeModeExternal {
		t.Errorf(errFmtBridgeMode, cfg.Network.BridgeMode, BridgeModeExternal)
	}

	if cfg.Network.PrivateSubnet != testPrivateSubnetSecond {
		t.Errorf(errFmtPrivateSubnet, cfg.Network.PrivateSubnet, testPrivateSubnetSecond)
	}
}

func TestLoadFromEnvInterfaceNameEmptyKeepsOriginal(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Network.InterfaceName = testInterfaceEnp // Set initial value

	t.Setenv("INTERFACE_NAME", "")
	LoadFromEnv(cfg)

	if cfg.Network.InterfaceName != testInterfaceEnp {
		t.Errorf("Empty INTERFACE_NAME changed config: got %q, want %q", cfg.Network.InterfaceName, testInterfaceEnp)
	}
}

// Storage configuration tests

func TestLoadFromEnvZFSRaidSingle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Storage.ZFSRaid = ZFSRaid1 // Set to different value first

	t.Setenv("ZFS_RAID", "single")
	LoadFromEnv(cfg)

	if cfg.Storage.ZFSRaid != ZFSRaidSingle {
		t.Errorf(errFmtZFSRaid, cfg.Storage.ZFSRaid, ZFSRaidSingle)
	}
}

func TestLoadFromEnvZFSRaidRaid0(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("ZFS_RAID", "raid0")
	LoadFromEnv(cfg)

	if cfg.Storage.ZFSRaid != ZFSRaid0 {
		t.Errorf(errFmtZFSRaid, cfg.Storage.ZFSRaid, ZFSRaid0)
	}
}

func TestLoadFromEnvZFSRaidRaid1(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Storage.ZFSRaid = ZFSRaidSingle // Set to different value first

	t.Setenv("ZFS_RAID", "raid1")
	LoadFromEnv(cfg)

	if cfg.Storage.ZFSRaid != ZFSRaid1 {
		t.Errorf(errFmtZFSRaid, cfg.Storage.ZFSRaid, ZFSRaid1)
	}
}

func TestLoadFromEnvZFSRaidCaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ZFSRaid
	}{
		{"uppercase SINGLE", "SINGLE", ZFSRaidSingle},
		{"mixed case Single", "Single", ZFSRaidSingle},
		{"uppercase RAID0", "RAID0", ZFSRaid0},
		{"mixed case Raid0", "Raid0", ZFSRaid0},
		{"uppercase RAID1", "RAID1", ZFSRaid1},
		{"mixed case Raid1", "Raid1", ZFSRaid1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Storage.ZFSRaid = "" // Clear default

			t.Setenv("ZFS_RAID", tt.input)
			LoadFromEnv(cfg)

			if cfg.Storage.ZFSRaid != tt.want {
				t.Errorf(errFmtZFSRaid, cfg.Storage.ZFSRaid, tt.want)
			}
		})
	}
}

func TestLoadFromEnvZFSRaidInvalidKeepsOriginal(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Storage.ZFSRaid

	// Set to an invalid value - should NOT change the config
	t.Setenv("ZFS_RAID", "invalid")
	LoadFromEnv(cfg)

	if cfg.Storage.ZFSRaid != original {
		t.Errorf("Invalid ZFS_RAID changed config: got %q, want %q", cfg.Storage.ZFSRaid, original)
	}
}

func TestLoadFromEnvZFSRaidEmptyKeepsOriginal(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Storage.ZFSRaid

	// Set to empty value - should NOT change the config
	t.Setenv("ZFS_RAID", "")
	LoadFromEnv(cfg)

	if cfg.Storage.ZFSRaid != original {
		t.Errorf("Empty ZFS_RAID changed config: got %q, want %q", cfg.Storage.ZFSRaid, original)
	}
}

func TestLoadFromEnvDisksSingle(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("DISKS", testDiskSda)
	LoadFromEnv(cfg)

	if len(cfg.Storage.Disks) != 1 {
		t.Fatalf("Disks length = %d, want 1", len(cfg.Storage.Disks))
	}

	if cfg.Storage.Disks[0] != testDiskSda {
		t.Errorf("Disks[0] = %q, want %q", cfg.Storage.Disks[0], testDiskSda)
	}
}

func TestLoadFromEnvDisksMultiple(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("DISKS", testDiskSda+","+testDiskSdb)
	LoadFromEnv(cfg)

	if len(cfg.Storage.Disks) != 2 {
		t.Fatalf("Disks length = %d, want 2", len(cfg.Storage.Disks))
	}

	if cfg.Storage.Disks[0] != testDiskSda {
		t.Errorf("Disks[0] = %q, want %q", cfg.Storage.Disks[0], testDiskSda)
	}

	if cfg.Storage.Disks[1] != testDiskSdb {
		t.Errorf("Disks[1] = %q, want %q", cfg.Storage.Disks[1], testDiskSdb)
	}
}

func TestLoadFromEnvDisksTrimsWhitespace(t *testing.T) {
	cfg := DefaultConfig()

	// Spaces around comma-separated values
	t.Setenv("DISKS", testDiskSda+" , "+testDiskSdb)
	LoadFromEnv(cfg)

	if len(cfg.Storage.Disks) != 2 {
		t.Fatalf("Disks length = %d, want 2", len(cfg.Storage.Disks))
	}

	if cfg.Storage.Disks[0] != testDiskSda {
		t.Errorf("Disks[0] = %q, want %q", cfg.Storage.Disks[0], testDiskSda)
	}

	if cfg.Storage.Disks[1] != testDiskSdb {
		t.Errorf("Disks[1] = %q, want %q", cfg.Storage.Disks[1], testDiskSdb)
	}
}

func TestLoadFromEnvDisksTrailingComma(t *testing.T) {
	cfg := DefaultConfig()

	// Trailing comma should be handled (filtered out empty string)
	t.Setenv("DISKS", testDiskSda+","+testDiskSdb+",")
	LoadFromEnv(cfg)

	if len(cfg.Storage.Disks) != 2 {
		t.Fatalf("Disks length = %d, want 2", len(cfg.Storage.Disks))
	}

	if cfg.Storage.Disks[0] != testDiskSda {
		t.Errorf("Disks[0] = %q, want %q", cfg.Storage.Disks[0], testDiskSda)
	}

	if cfg.Storage.Disks[1] != testDiskSdb {
		t.Errorf("Disks[1] = %q, want %q", cfg.Storage.Disks[1], testDiskSdb)
	}
}

func TestLoadFromEnvDisksEmptyKeepsOriginal(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Storage.Disks = []string{testDiskSdc} // Set initial value

	t.Setenv("DISKS", "")
	LoadFromEnv(cfg)

	if len(cfg.Storage.Disks) != 1 {
		t.Fatalf("Empty DISKS changed length: got %d, want 1", len(cfg.Storage.Disks))
	}

	if cfg.Storage.Disks[0] != testDiskSdc {
		t.Errorf("Empty DISKS changed config: got %q, want %q", cfg.Storage.Disks[0], testDiskSdc)
	}
}

func TestLoadFromEnvDisksOnlyCommasKeepsOriginal(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Storage.Disks = []string{testDiskSdc} // Set initial value

	// Only commas should result in all empty strings being filtered out
	t.Setenv("DISKS", ",,,")
	LoadFromEnv(cfg)

	if len(cfg.Storage.Disks) != 1 {
		t.Fatalf("Commas-only DISKS changed length: got %d, want 1", len(cfg.Storage.Disks))
	}

	if cfg.Storage.Disks[0] != testDiskSdc {
		t.Errorf("Commas-only DISKS changed config: got %q, want %q", cfg.Storage.Disks[0], testDiskSdc)
	}
}

func TestLoadFromEnvStorageMultipleFields(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("ZFS_RAID", "raid0")
	t.Setenv("DISKS", testDiskSda+","+testDiskSdb+","+testDiskSdc)

	LoadFromEnv(cfg)

	if cfg.Storage.ZFSRaid != ZFSRaid0 {
		t.Errorf(errFmtZFSRaid, cfg.Storage.ZFSRaid, ZFSRaid0)
	}

	if len(cfg.Storage.Disks) != 3 {
		t.Fatalf("Disks length = %d, want 3", len(cfg.Storage.Disks))
	}

	expectedDisks := []string{testDiskSda, testDiskSdb, testDiskSdc}
	for i, want := range expectedDisks {
		if cfg.Storage.Disks[i] != want {
			t.Errorf("Disks[%d] = %q, want %q", i, cfg.Storage.Disks[i], want)
		}
	}
}
