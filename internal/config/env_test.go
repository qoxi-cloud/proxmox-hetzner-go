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
		{testCaseTrueLowercase, "true", true},
		{"true mixed case", "True", true},
		{testCaseTrueUppercase, "TRUE", true},

		// True values - "yes" variations
		{testCaseYesLowercase, "yes", true},
		{"yes mixed case", "Yes", true},
		{"yes uppercase", "YES", true},

		// True values - "1"
		{"one", "1", true},

		// False values - "false" variations
		{testCaseFalseLowercase, "false", false},
		{"false mixed case", "False", false},
		{"false uppercase", "FALSE", false},

		// False values - "no" variations
		{testCaseNoLowercase, "no", false},
		{"no mixed case", "No", false},
		{"no uppercase", "NO", false},

		// False values - "0"
		{"zero", "0", false},

		// Empty and whitespace
		{testCaseEmptyString, "", false},
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

	// Test case name constants.
	testCaseEmptyString    = "empty string"
	testCaseTrueLowercase  = "true lowercase"
	testCaseTrueUppercase  = "TRUE uppercase"
	testCaseYesLowercase   = "yes lowercase"
	testCaseFalseLowercase = "false lowercase"
	testCaseNoLowercase    = "no lowercase"
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

func TestLoadFromEnvPreservesUnsetFields(t *testing.T) {
	cfg := DefaultConfig()
	originalZFSRaid := cfg.Storage.ZFSRaid
	originalTailscaleEnabled := cfg.Tailscale.Enabled
	originalTailscaleSSH := cfg.Tailscale.SSH
	testPreserve := "preserve-test"

	// Only set hostname - other env vars should preserve their defaults
	t.Setenv("PVE_HOSTNAME", testPreserve)
	LoadFromEnv(cfg)

	// Verify unset env vars don't change config
	if cfg.Storage.ZFSRaid != originalZFSRaid {
		t.Errorf("Storage.ZFSRaid changed unexpectedly: got %v, want %v", cfg.Storage.ZFSRaid, originalZFSRaid)
	}

	if cfg.Tailscale.Enabled != originalTailscaleEnabled {
		t.Errorf("Tailscale.Enabled changed unexpectedly: got %v, want %v", cfg.Tailscale.Enabled, originalTailscaleEnabled)
	}

	if cfg.Tailscale.SSH != originalTailscaleSSH {
		t.Errorf("Tailscale.SSH changed unexpectedly: got %v, want %v", cfg.Tailscale.SSH, originalTailscaleSSH)
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

// assertDisksEqual is a test helper that verifies disk slices match expected values.
func assertDisksEqual(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("Disks length = %d, want %d", len(got), len(want))
	}

	for i, w := range want {
		if got[i] != w {
			t.Errorf("Disks[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestLoadFromEnvZFSRaidValues(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ZFSRaid
		initial ZFSRaid
	}{
		{"single lowercase", "single", ZFSRaidSingle, ZFSRaid1},
		{"raid0 lowercase", "raid0", ZFSRaid0, ZFSRaid1},
		{"raid1 lowercase", "raid1", ZFSRaid1, ZFSRaidSingle},
		{"uppercase SINGLE", "SINGLE", ZFSRaidSingle, ZFSRaid1},
		{"mixed case Single", "Single", ZFSRaidSingle, ZFSRaid1},
		{"uppercase RAID0", "RAID0", ZFSRaid0, ZFSRaid1},
		{"mixed case Raid0", "Raid0", ZFSRaid0, ZFSRaid1},
		{"uppercase RAID1", "RAID1", ZFSRaid1, ZFSRaidSingle},
		{"mixed case Raid1", "Raid1", ZFSRaid1, ZFSRaidSingle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Storage.ZFSRaid = tt.initial

			t.Setenv("ZFS_RAID", tt.input)
			LoadFromEnv(cfg)

			if cfg.Storage.ZFSRaid != tt.want {
				t.Errorf(errFmtZFSRaid, cfg.Storage.ZFSRaid, tt.want)
			}
		})
	}
}

func TestLoadFromEnvZFSRaidInvalidKeepsOriginal(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid value", "invalid"},
		{testCaseEmptyString, ""},
		{"raid5 unsupported", "raid5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			original := cfg.Storage.ZFSRaid

			t.Setenv("ZFS_RAID", tt.input)
			LoadFromEnv(cfg)

			if cfg.Storage.ZFSRaid != original {
				t.Errorf("ZFS_RAID %q changed config: got %q, want %q", tt.input, cfg.Storage.ZFSRaid, original)
			}
		})
	}
}

func TestLoadFromEnvDisksValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single disk", testDiskSda, []string{testDiskSda}},
		{"two disks", testDiskSda + "," + testDiskSdb, []string{testDiskSda, testDiskSdb}},
		{"three disks", testDiskSda + "," + testDiskSdb + "," + testDiskSdc, []string{testDiskSda, testDiskSdb, testDiskSdc}},
		{"with spaces", testDiskSda + " , " + testDiskSdb, []string{testDiskSda, testDiskSdb}},
		{"trailing comma", testDiskSda + "," + testDiskSdb + ",", []string{testDiskSda, testDiskSdb}},
		{"leading comma", "," + testDiskSda + "," + testDiskSdb, []string{testDiskSda, testDiskSdb}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()

			t.Setenv("DISKS", tt.input)
			LoadFromEnv(cfg)

			assertDisksEqual(t, cfg.Storage.Disks, tt.want)
		})
	}
}

func TestLoadFromEnvDisksKeepsOriginal(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{testCaseEmptyString, ""},
		{"only commas", ",,,"},
		{"only spaces and commas", " , , "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Storage.Disks = []string{testDiskSdc}

			t.Setenv("DISKS", tt.input)
			LoadFromEnv(cfg)

			assertDisksEqual(t, cfg.Storage.Disks, []string{testDiskSdc})
		})
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

	assertDisksEqual(t, cfg.Storage.Disks, []string{testDiskSda, testDiskSdb, testDiskSdc})
}

// Tailscale configuration tests

// Error format strings for Tailscale tests.
const (
	errFmtTailscaleEnabled = "Tailscale.Enabled = %v, want %v"
	errFmtTailscaleSSH     = "Tailscale.SSH = %v, want %v"
	errFmtTailscaleWebUI   = "Tailscale.WebUI = %v, want %v"
	errFmtTailscaleAuthKey = "Tailscale.AuthKey = %q, want %q"
)

func TestLoadFromEnvTailscaleEnabledTrue(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{testCaseTrueLowercase, "true"},
		{testCaseYesLowercase, "yes"},
		{"one", "1"},
		{testCaseTrueUppercase, "TRUE"},
		{"Yes mixed case", "Yes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Tailscale.Enabled = false // Ensure it's false initially

			t.Setenv("INSTALL_TAILSCALE", tt.input)
			LoadFromEnv(cfg)

			if !cfg.Tailscale.Enabled {
				t.Errorf(errFmtTailscaleEnabled, cfg.Tailscale.Enabled, true)
			}
		})
	}
}

func TestLoadFromEnvTailscaleEnabledFalse(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{testCaseFalseLowercase, "false"},
		{testCaseNoLowercase, "no"},
		{"zero", "0"},
		{"FALSE uppercase", "FALSE"},
		{"No mixed case", "No"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Tailscale.Enabled = true // Ensure it's true initially

			t.Setenv("INSTALL_TAILSCALE", tt.input)
			LoadFromEnv(cfg)

			if cfg.Tailscale.Enabled {
				t.Errorf(errFmtTailscaleEnabled, cfg.Tailscale.Enabled, false)
			}
		})
	}
}

func TestLoadFromEnvTailscaleEnabledUnsetPreservesDefault(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Tailscale.Enabled

	// Do NOT set INSTALL_TAILSCALE - should preserve default
	LoadFromEnv(cfg)

	if cfg.Tailscale.Enabled != original {
		t.Errorf("Unset INSTALL_TAILSCALE changed config: got %v, want %v", cfg.Tailscale.Enabled, original)
	}
}

func TestLoadFromEnvTailscaleAuthKeyWithValue(t *testing.T) {
	cfg := DefaultConfig()
	testKey := "tskey-auth-xxxx-xxxxxxxxxxxxxxxxx" // NOSONAR(go:S2068) test value, not a real key

	t.Setenv("TAILSCALE_AUTH_KEY", testKey)
	LoadFromEnv(cfg)

	if cfg.Tailscale.AuthKey != testKey {
		t.Errorf(errFmtTailscaleAuthKey, cfg.Tailscale.AuthKey, testKey)
	}
}

func TestLoadFromEnvTailscaleAuthKeyEmptyPreservesOriginal(t *testing.T) {
	cfg := DefaultConfig()
	original := "existing-key" // NOSONAR(go:S2068) test value, not a real key
	cfg.Tailscale.AuthKey = original

	t.Setenv("TAILSCALE_AUTH_KEY", "")
	LoadFromEnv(cfg)

	if cfg.Tailscale.AuthKey != original {
		t.Errorf("Empty TAILSCALE_AUTH_KEY changed config: got %q, want %q", cfg.Tailscale.AuthKey, original)
	}
}

func TestLoadFromEnvTailscaleSSHTrue(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{testCaseTrueLowercase, "true"},
		{testCaseYesLowercase, "yes"},
		{"one", "1"},
		{testCaseTrueUppercase, "TRUE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Tailscale.SSH = false // Ensure it's false initially

			t.Setenv("TAILSCALE_SSH", tt.input)
			LoadFromEnv(cfg)

			if !cfg.Tailscale.SSH {
				t.Errorf(errFmtTailscaleSSH, cfg.Tailscale.SSH, true)
			}
		})
	}
}

func TestLoadFromEnvTailscaleSSHFalse(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{testCaseFalseLowercase, "false"},
		{testCaseNoLowercase, "no"},
		{"zero", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Tailscale.SSH = true // Ensure it's true initially (default)

			t.Setenv("TAILSCALE_SSH", tt.input)
			LoadFromEnv(cfg)

			if cfg.Tailscale.SSH {
				t.Errorf(errFmtTailscaleSSH, cfg.Tailscale.SSH, false)
			}
		})
	}
}

func TestLoadFromEnvTailscaleSSHUnsetPreservesDefault(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Tailscale.SSH

	// Do NOT set TAILSCALE_SSH - should preserve default
	LoadFromEnv(cfg)

	if cfg.Tailscale.SSH != original {
		t.Errorf("Unset TAILSCALE_SSH changed config: got %v, want %v", cfg.Tailscale.SSH, original)
	}
}

func TestLoadFromEnvTailscaleWebUITrue(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{testCaseTrueLowercase, "true"},
		{testCaseYesLowercase, "yes"},
		{"one", "1"},
		{testCaseTrueUppercase, "TRUE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Tailscale.WebUI = false // Ensure it's false initially (default)

			t.Setenv("TAILSCALE_WEBUI", tt.input)
			LoadFromEnv(cfg)

			if !cfg.Tailscale.WebUI {
				t.Errorf(errFmtTailscaleWebUI, cfg.Tailscale.WebUI, true)
			}
		})
	}
}

func TestLoadFromEnvTailscaleWebUIFalse(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{testCaseFalseLowercase, "false"},
		{testCaseNoLowercase, "no"},
		{"zero", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Tailscale.WebUI = true // Ensure it's true initially

			t.Setenv("TAILSCALE_WEBUI", tt.input)
			LoadFromEnv(cfg)

			if cfg.Tailscale.WebUI {
				t.Errorf(errFmtTailscaleWebUI, cfg.Tailscale.WebUI, false)
			}
		})
	}
}

func TestLoadFromEnvTailscaleWebUIUnsetPreservesDefault(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Tailscale.WebUI

	// Do NOT set TAILSCALE_WEBUI - should preserve default
	LoadFromEnv(cfg)

	if cfg.Tailscale.WebUI != original {
		t.Errorf("Unset TAILSCALE_WEBUI changed config: got %v, want %v", cfg.Tailscale.WebUI, original)
	}
}

func TestLoadFromEnvTailscaleMultipleFields(t *testing.T) {
	cfg := DefaultConfig()
	testKey := "tskey-auth-multi-test" // NOSONAR(go:S2068) test value, not a real key

	t.Setenv("INSTALL_TAILSCALE", "true")
	t.Setenv("TAILSCALE_AUTH_KEY", testKey)
	t.Setenv("TAILSCALE_SSH", "false")
	t.Setenv("TAILSCALE_WEBUI", "true")

	LoadFromEnv(cfg)

	if !cfg.Tailscale.Enabled {
		t.Errorf(errFmtTailscaleEnabled, cfg.Tailscale.Enabled, true)
	}

	if cfg.Tailscale.AuthKey != testKey {
		t.Errorf(errFmtTailscaleAuthKey, cfg.Tailscale.AuthKey, testKey)
	}

	if cfg.Tailscale.SSH {
		t.Errorf(errFmtTailscaleSSH, cfg.Tailscale.SSH, false)
	}

	if !cfg.Tailscale.WebUI {
		t.Errorf(errFmtTailscaleWebUI, cfg.Tailscale.WebUI, true)
	}
}

func TestLoadFromEnvTailscaleBooleanCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		input    string
		getField func(*Config) bool
		want     bool
	}{
		{"INSTALL_TAILSCALE True", "INSTALL_TAILSCALE", "True", func(c *Config) bool { return c.Tailscale.Enabled }, true},
		{"INSTALL_TAILSCALE TRUE", "INSTALL_TAILSCALE", "TRUE", func(c *Config) bool { return c.Tailscale.Enabled }, true},
		{"INSTALL_TAILSCALE yEs", "INSTALL_TAILSCALE", "yEs", func(c *Config) bool { return c.Tailscale.Enabled }, true},
		{"INSTALL_TAILSCALE YES", "INSTALL_TAILSCALE", "YES", func(c *Config) bool { return c.Tailscale.Enabled }, true},
		{"TAILSCALE_SSH TrUe", "TAILSCALE_SSH", "TrUe", func(c *Config) bool { return c.Tailscale.SSH }, true},
		{"TAILSCALE_WEBUI TRUE", "TAILSCALE_WEBUI", "TRUE", func(c *Config) bool { return c.Tailscale.WebUI }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()

			t.Setenv(tt.envName, tt.input)
			LoadFromEnv(cfg)

			got := tt.getField(cfg)
			if got != tt.want {
				t.Errorf("%s with value %q: got %v, want %v", tt.envName, tt.input, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

// Integration test constants.
const (
	// NOSONAR(go:S1313) - test IP address values, not real network addresses.
	intTestSubnet        = "172.16.0.0/24" // NOSONAR(go:S1313) RFC 1918 test value
	intTestInterfaceName = "enp0s31f6"     // NOSONAR(go:S1313) test value
)

// TestLoadFromEnvFullConfiguration verifies that all supported environment
// variables are correctly loaded into the configuration structure.
// This is a comprehensive integration test that sets all fields via env vars.
func TestLoadFromEnvFullConfiguration(t *testing.T) {
	cfg := DefaultConfig()

	// NOSONAR(go:S2068) - test credentials, not real values
	testPassword := "envpassword123" // NOSONAR(go:S2068) test value
	testAuthKey := "tskey-auth-123"  // NOSONAR(go:S2068) test value

	// System configuration
	t.Setenv("PVE_HOSTNAME", "env-server")
	t.Setenv("PVE_DOMAIN_SUFFIX", "env.local")
	t.Setenv("PVE_TIMEZONE", "America/New_York")
	t.Setenv("PVE_EMAIL", "env@test.com")
	t.Setenv("PVE_ROOT_PASSWORD", testPassword)
	t.Setenv("PVE_SSH_PUBLIC_KEY", "ssh-ed25519 AAAA... test@example.com")

	// Network configuration
	t.Setenv("INTERFACE_NAME", intTestInterfaceName)
	t.Setenv("BRIDGE_MODE", "both")
	t.Setenv("PRIVATE_SUBNET", intTestSubnet)

	// Storage configuration
	t.Setenv("ZFS_RAID", "raid0")
	t.Setenv("DISKS", "/dev/nvme0n1,/dev/nvme1n1")

	// Tailscale configuration
	t.Setenv("INSTALL_TAILSCALE", "true")
	t.Setenv("TAILSCALE_AUTH_KEY", testAuthKey)
	t.Setenv("TAILSCALE_SSH", "yes")
	t.Setenv("TAILSCALE_WEBUI", "1")

	LoadFromEnv(cfg)

	// Verify System configuration
	if cfg.System.Hostname != "env-server" {
		t.Errorf("System.Hostname = %q, want %q", cfg.System.Hostname, "env-server")
	}

	if cfg.System.DomainSuffix != "env.local" {
		t.Errorf("System.DomainSuffix = %q, want %q", cfg.System.DomainSuffix, "env.local")
	}

	if cfg.System.Timezone != "America/New_York" {
		t.Errorf("System.Timezone = %q, want %q", cfg.System.Timezone, "America/New_York")
	}

	if cfg.System.Email != "env@test.com" {
		t.Errorf("System.Email = %q, want %q", cfg.System.Email, "env@test.com")
	}

	if cfg.System.RootPassword != testPassword {
		t.Errorf("System.RootPassword = %q, want %q", cfg.System.RootPassword, testPassword)
	}

	if cfg.System.SSHPublicKey != "ssh-ed25519 AAAA... test@example.com" {
		t.Errorf("System.SSHPublicKey = %q, want %q", cfg.System.SSHPublicKey, "ssh-ed25519 AAAA... test@example.com")
	}

	// Verify Network configuration
	if cfg.Network.InterfaceName != intTestInterfaceName {
		t.Errorf("Network.InterfaceName = %q, want %q", cfg.Network.InterfaceName, intTestInterfaceName)
	}

	if cfg.Network.BridgeMode != BridgeModeBoth {
		t.Errorf("Network.BridgeMode = %q, want %q", cfg.Network.BridgeMode, BridgeModeBoth)
	}

	if cfg.Network.PrivateSubnet != intTestSubnet {
		t.Errorf("Network.PrivateSubnet = %q, want %q", cfg.Network.PrivateSubnet, intTestSubnet)
	}

	// Verify Storage configuration
	if cfg.Storage.ZFSRaid != ZFSRaid0 {
		t.Errorf("Storage.ZFSRaid = %q, want %q", cfg.Storage.ZFSRaid, ZFSRaid0)
	}

	expectedDisks := []string{"/dev/nvme0n1", "/dev/nvme1n1"}
	assertDisksEqual(t, cfg.Storage.Disks, expectedDisks)

	// Verify Tailscale configuration
	if !cfg.Tailscale.Enabled {
		t.Errorf("Tailscale.Enabled = %v, want %v", cfg.Tailscale.Enabled, true)
	}

	if cfg.Tailscale.AuthKey != testAuthKey {
		t.Errorf("Tailscale.AuthKey = %q, want %q", cfg.Tailscale.AuthKey, testAuthKey)
	}

	if !cfg.Tailscale.SSH {
		t.Errorf("Tailscale.SSH = %v, want %v", cfg.Tailscale.SSH, true)
	}

	if !cfg.Tailscale.WebUI {
		t.Errorf("Tailscale.WebUI = %v, want %v", cfg.Tailscale.WebUI, true)
	}
}

// TestLoadFromEnvPartialConfiguration verifies that setting only some
// environment variables correctly overrides those specific fields while
// preserving default values for all other fields.
func TestLoadFromEnvPartialConfiguration(t *testing.T) {
	cfg := DefaultConfig()

	// Store original defaults
	originalDomain := cfg.System.DomainSuffix
	originalTimezone := cfg.System.Timezone
	originalEmail := cfg.System.Email
	originalRootPassword := cfg.System.RootPassword
	originalSSHKey := cfg.System.SSHPublicKey
	originalBridgeMode := cfg.Network.BridgeMode
	originalPrivateSubnet := cfg.Network.PrivateSubnet
	originalZFSRaid := cfg.Storage.ZFSRaid
	originalDisks := cfg.Storage.Disks
	originalTailscaleEnabled := cfg.Tailscale.Enabled
	originalTailscaleSSH := cfg.Tailscale.SSH
	originalTailscaleWebUI := cfg.Tailscale.WebUI

	// Clear any externally-set env vars that would interfere with testing
	// This ensures the test is isolated from the host environment
	// Note: For string fields, setting to "" is sufficient (empty values are ignored)
	// For boolean fields that use EnvVarSet, we must actually unset them
	clearStringEnvVars := []string{
		"PVE_DOMAIN_SUFFIX", "PVE_TIMEZONE", "PVE_EMAIL",
		"PVE_ROOT_PASSWORD", "PVE_SSH_PUBLIC_KEY",
		"BRIDGE_MODE", "PRIVATE_SUBNET",
		"ZFS_RAID", "DISKS",
		"TAILSCALE_AUTH_KEY",
	}
	for _, env := range clearStringEnvVars {
		t.Setenv(env, "")
	}

	// Boolean env vars use EnvVarSet which checks existence, so we must unset them.
	// We store original values and restore them in cleanup.
	boolEnvVars := []string{"INSTALL_TAILSCALE", "TAILSCALE_SSH", "TAILSCALE_WEBUI"}
	for _, env := range boolEnvVars {
		originalVal, wasSet := os.LookupEnv(env)
		if wasSet {
			if err := os.Unsetenv(env); err != nil {
				t.Fatalf("failed to unset %s: %v", env, err)
			}
			envCopy := env         // Capture for closure
			valCopy := originalVal // Capture for closure
			t.Cleanup(func() {
				// Restore original value. Error ignored as this is cleanup code.
				//nolint:errcheck,usetesting // cleanup code, t.Setenv not available
				os.Setenv(envCopy, valCopy)
			})
		}
	}

	// Only set a subset of environment variables
	t.Setenv("PVE_HOSTNAME", "partial-server")
	t.Setenv("INTERFACE_NAME", "eth1")

	LoadFromEnv(cfg)

	// Verify set fields were updated
	if cfg.System.Hostname != "partial-server" {
		t.Errorf("System.Hostname = %q, want %q", cfg.System.Hostname, "partial-server")
	}

	if cfg.Network.InterfaceName != "eth1" {
		t.Errorf("Network.InterfaceName = %q, want %q", cfg.Network.InterfaceName, "eth1")
	}

	// Verify all other fields retain defaults
	if cfg.System.DomainSuffix != originalDomain {
		t.Errorf("System.DomainSuffix = %q, want %q (original)", cfg.System.DomainSuffix, originalDomain)
	}

	if cfg.System.Timezone != originalTimezone {
		t.Errorf("System.Timezone = %q, want %q (original)", cfg.System.Timezone, originalTimezone)
	}

	if cfg.System.Email != originalEmail {
		t.Errorf("System.Email = %q, want %q (original)", cfg.System.Email, originalEmail)
	}

	if cfg.System.RootPassword != originalRootPassword {
		t.Errorf("System.RootPassword changed unexpectedly")
	}

	if cfg.System.SSHPublicKey != originalSSHKey {
		t.Errorf("System.SSHPublicKey changed unexpectedly")
	}

	if cfg.Network.BridgeMode != originalBridgeMode {
		t.Errorf("Network.BridgeMode = %q, want %q (original)", cfg.Network.BridgeMode, originalBridgeMode)
	}

	if cfg.Network.PrivateSubnet != originalPrivateSubnet {
		t.Errorf("Network.PrivateSubnet = %q, want %q (original)", cfg.Network.PrivateSubnet, originalPrivateSubnet)
	}

	if cfg.Storage.ZFSRaid != originalZFSRaid {
		t.Errorf("Storage.ZFSRaid = %q, want %q (original)", cfg.Storage.ZFSRaid, originalZFSRaid)
	}

	if len(cfg.Storage.Disks) != len(originalDisks) {
		t.Errorf("Storage.Disks length changed unexpectedly")
	}

	if cfg.Tailscale.Enabled != originalTailscaleEnabled {
		t.Errorf("Tailscale.Enabled = %v, want %v (original)", cfg.Tailscale.Enabled, originalTailscaleEnabled)
	}

	// Note: AuthKey default is empty, verify it remains empty
	if cfg.Tailscale.AuthKey != "" {
		t.Errorf("Tailscale.AuthKey = %q, want empty (default)", cfg.Tailscale.AuthKey)
	}

	if cfg.Tailscale.SSH != originalTailscaleSSH {
		t.Errorf("Tailscale.SSH = %v, want %v (original)", cfg.Tailscale.SSH, originalTailscaleSSH)
	}

	if cfg.Tailscale.WebUI != originalTailscaleWebUI {
		t.Errorf("Tailscale.WebUI = %v, want %v (original)", cfg.Tailscale.WebUI, originalTailscaleWebUI)
	}
}

// TestLoadFromEnvOverridesFileConfig verifies that environment variables
// take priority over values loaded from a configuration file.
// Configuration priority: 1. TUI > 2. Environment > 3. File > 4. Defaults.
func TestLoadFromEnvOverridesFileConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := tempDir + "/test-config.yaml"

	// Create config with file-based values
	fileConfig := DefaultConfig()
	fileConfig.System.Hostname = "file-hostname"
	fileConfig.System.DomainSuffix = "file.local"
	fileConfig.System.Timezone = "Europe/London"
	fileConfig.Network.InterfaceName = "eth0"
	fileConfig.Network.BridgeMode = BridgeModeInternal
	fileConfig.Storage.ZFSRaid = ZFSRaid1
	fileConfig.Tailscale.Enabled = false
	fileConfig.Tailscale.SSH = false
	fileConfig.Tailscale.WebUI = false

	// Save config to file
	if err := fileConfig.SaveToFile(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Load config from file
	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Verify file values were loaded
	if cfg.System.Hostname != "file-hostname" {
		t.Fatalf("File config not loaded correctly: Hostname = %q", cfg.System.Hostname)
	}

	// Set environment variables that should override file values
	t.Setenv("PVE_HOSTNAME", "env-hostname")
	t.Setenv("PVE_DOMAIN_SUFFIX", "env.local")
	t.Setenv("INTERFACE_NAME", "enp0s25")
	t.Setenv("BRIDGE_MODE", "external")
	t.Setenv("ZFS_RAID", "raid0")
	t.Setenv("INSTALL_TAILSCALE", "true")
	t.Setenv("TAILSCALE_SSH", "true")
	t.Setenv("TAILSCALE_WEBUI", "true")

	// Apply environment variables
	LoadFromEnv(cfg)

	// Verify environment variables override file values
	if cfg.System.Hostname != "env-hostname" {
		t.Errorf("Env did not override Hostname: got %q, want %q", cfg.System.Hostname, "env-hostname")
	}

	if cfg.System.DomainSuffix != "env.local" {
		t.Errorf("Env did not override DomainSuffix: got %q, want %q", cfg.System.DomainSuffix, "env.local")
	}

	// Timezone was not set in env, should retain file value
	if cfg.System.Timezone != "Europe/London" {
		t.Errorf("Unset env var changed Timezone: got %q, want %q", cfg.System.Timezone, "Europe/London")
	}

	if cfg.Network.InterfaceName != "enp0s25" {
		t.Errorf("Env did not override InterfaceName: got %q, want %q", cfg.Network.InterfaceName, "enp0s25")
	}

	if cfg.Network.BridgeMode != BridgeModeExternal {
		t.Errorf("Env did not override BridgeMode: got %q, want %q", cfg.Network.BridgeMode, BridgeModeExternal)
	}

	if cfg.Storage.ZFSRaid != ZFSRaid0 {
		t.Errorf("Env did not override ZFSRaid: got %q, want %q", cfg.Storage.ZFSRaid, ZFSRaid0)
	}

	if !cfg.Tailscale.Enabled {
		t.Errorf("Env did not override Tailscale.Enabled: got %v, want %v", cfg.Tailscale.Enabled, true)
	}

	if !cfg.Tailscale.SSH {
		t.Errorf("Env did not override Tailscale.SSH: got %v, want %v", cfg.Tailscale.SSH, true)
	}

	if !cfg.Tailscale.WebUI {
		t.Errorf("Env did not override Tailscale.WebUI: got %v, want %v", cfg.Tailscale.WebUI, true)
	}
}

// TestLoadFromEnvEmptyVsUnset verifies the distinction between:
// - Empty string env vars (set but empty) - should NOT override
// - Unset env vars - should NOT override
// Both should preserve the existing configuration values.
func TestLoadFromEnvEmptyVsUnset(t *testing.T) {
	tests := []struct {
		name         string
		envName      string
		setValue     *string // nil = unset, "" = empty string
		initialValue string
		getField     func(*Config) string
	}{
		{
			name:         "PVE_HOSTNAME empty preserves value",
			envName:      "PVE_HOSTNAME",
			setValue:     ptrString(""),
			initialValue: "original-hostname",
			getField:     func(c *Config) string { return c.System.Hostname },
		},
		{
			name:         "PVE_DOMAIN_SUFFIX empty preserves value",
			envName:      "PVE_DOMAIN_SUFFIX",
			setValue:     ptrString(""),
			initialValue: "original.local",
			getField:     func(c *Config) string { return c.System.DomainSuffix },
		},
		{
			name:         "INTERFACE_NAME empty preserves value",
			envName:      "INTERFACE_NAME",
			setValue:     ptrString(""),
			initialValue: "eth99",
			getField:     func(c *Config) string { return c.Network.InterfaceName },
		},
		{
			name:         "PRIVATE_SUBNET empty preserves value",
			envName:      "PRIVATE_SUBNET",
			setValue:     ptrString(""),
			initialValue: "192.168.99.0/24", // NOSONAR(go:S1313) RFC 1918 test value
			getField:     func(c *Config) string { return c.Network.PrivateSubnet },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()

			// Set the initial value
			switch tt.envName {
			case "PVE_HOSTNAME":
				cfg.System.Hostname = tt.initialValue
			case "PVE_DOMAIN_SUFFIX":
				cfg.System.DomainSuffix = tt.initialValue
			case "INTERFACE_NAME":
				cfg.Network.InterfaceName = tt.initialValue
			case "PRIVATE_SUBNET":
				cfg.Network.PrivateSubnet = tt.initialValue
			}

			// Set env var to empty string
			if tt.setValue != nil {
				t.Setenv(tt.envName, *tt.setValue)
			}

			LoadFromEnv(cfg)

			got := tt.getField(cfg)
			if got != tt.initialValue {
				t.Errorf("%s: got %q, want %q (original)", tt.envName, got, tt.initialValue)
			}
		})
	}
}

// TestLoadFromEnvBooleanEdgeCases tests all variations of boolean parsing
// for the Tailscale boolean fields.
func TestLoadFromEnvBooleanEdgeCases(t *testing.T) {
	// True values
	trueValues := []string{"true", "True", "TRUE", "TrUe", "yes", "Yes", "YES", "yEs", "1"}

	// False values
	falseValues := []string{"false", "False", "FALSE", "FaLsE", "no", "No", "NO", "nO", "0"}

	// Invalid values that should be treated as false
	invalidValues := []string{"maybe", "2", "on", "off", "enabled", "disabled", "y", "n", ""}

	t.Run("INSTALL_TAILSCALE true values", func(t *testing.T) {
		for _, v := range trueValues {
			t.Run(v, func(t *testing.T) {
				cfg := DefaultConfig()
				cfg.Tailscale.Enabled = false

				t.Setenv("INSTALL_TAILSCALE", v)
				LoadFromEnv(cfg)

				if !cfg.Tailscale.Enabled {
					t.Errorf("INSTALL_TAILSCALE=%q: got false, want true", v)
				}
			})
		}
	})

	t.Run("INSTALL_TAILSCALE false values", func(t *testing.T) {
		for _, v := range falseValues {
			t.Run(v, func(t *testing.T) {
				cfg := DefaultConfig()
				cfg.Tailscale.Enabled = true

				t.Setenv("INSTALL_TAILSCALE", v)
				LoadFromEnv(cfg)

				if cfg.Tailscale.Enabled {
					t.Errorf("INSTALL_TAILSCALE=%q: got true, want false", v)
				}
			})
		}
	})

	t.Run("INSTALL_TAILSCALE invalid values treated as false", func(t *testing.T) {
		for _, v := range invalidValues {
			if v == "" {
				continue // Empty string is special case - env var is set but empty
			}

			t.Run(v, func(t *testing.T) {
				cfg := DefaultConfig()
				cfg.Tailscale.Enabled = true

				t.Setenv("INSTALL_TAILSCALE", v)
				LoadFromEnv(cfg)

				if cfg.Tailscale.Enabled {
					t.Errorf("INSTALL_TAILSCALE=%q: got true, want false (invalid treated as false)", v)
				}
			})
		}
	})

	t.Run("TAILSCALE_SSH true values", func(t *testing.T) {
		for _, v := range trueValues {
			t.Run(v, func(t *testing.T) {
				cfg := DefaultConfig()
				cfg.Tailscale.SSH = false

				t.Setenv("TAILSCALE_SSH", v)
				LoadFromEnv(cfg)

				if !cfg.Tailscale.SSH {
					t.Errorf("TAILSCALE_SSH=%q: got false, want true", v)
				}
			})
		}
	})

	t.Run("TAILSCALE_WEBUI true values", func(t *testing.T) {
		for _, v := range trueValues {
			t.Run(v, func(t *testing.T) {
				cfg := DefaultConfig()
				cfg.Tailscale.WebUI = false

				t.Setenv("TAILSCALE_WEBUI", v)
				LoadFromEnv(cfg)

				if !cfg.Tailscale.WebUI {
					t.Errorf("TAILSCALE_WEBUI=%q: got false, want true", v)
				}
			})
		}
	})

	t.Run("boolean with whitespace", func(t *testing.T) {
		whitespaceTests := []struct {
			input string
			want  bool
		}{
			{" true", true},
			{"true ", true},
			{" true ", true},
			{"\ttrue", true},
			{"true\n", true},
			{" \ttrue\n ", true},
			{" false ", false},
			{"\tno\n", false},
		}

		for _, tt := range whitespaceTests {
			t.Run(tt.input, func(t *testing.T) {
				cfg := DefaultConfig()
				cfg.Tailscale.Enabled = !tt.want

				t.Setenv("INSTALL_TAILSCALE", tt.input)
				LoadFromEnv(cfg)

				if cfg.Tailscale.Enabled != tt.want {
					t.Errorf("INSTALL_TAILSCALE=%q: got %v, want %v", tt.input, cfg.Tailscale.Enabled, tt.want)
				}
			})
		}
	})
}

// TestLoadFromEnvDisksFormats tests various formats for the DISKS environment variable.
func TestLoadFromEnvDisksFormats(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		want         []string
		wantOriginal bool // If true, original disks should be preserved
	}{
		// Valid formats
		{
			name:  "single disk no comma",
			input: "/dev/sda",
			want:  []string{"/dev/sda"},
		},
		{
			name:  "two disks",
			input: "/dev/sda,/dev/sdb",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "three disks",
			input: "/dev/sda,/dev/sdb,/dev/sdc",
			want:  []string{"/dev/sda", "/dev/sdb", "/dev/sdc"},
		},
		{
			name:  "nvme disks",
			input: "/dev/nvme0n1,/dev/nvme1n1",
			want:  []string{"/dev/nvme0n1", "/dev/nvme1n1"},
		},
		{
			name:  "mixed disk types",
			input: "/dev/sda,/dev/nvme0n1,/dev/vda",
			want:  []string{"/dev/sda", "/dev/nvme0n1", "/dev/vda"},
		},
		// Whitespace handling
		{
			name:  "spaces around commas",
			input: "/dev/sda , /dev/sdb",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "spaces before comma",
			input: "/dev/sda ,/dev/sdb",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "spaces after comma",
			input: "/dev/sda, /dev/sdb",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "tabs and spaces",
			input: "/dev/sda\t,\t/dev/sdb",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "leading whitespace",
			input: "  /dev/sda,/dev/sdb",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "trailing whitespace",
			input: "/dev/sda,/dev/sdb  ",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		// Edge cases with commas
		{
			name:  "trailing comma",
			input: "/dev/sda,/dev/sdb,",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "leading comma",
			input: ",/dev/sda,/dev/sdb",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "multiple trailing commas",
			input: "/dev/sda,/dev/sdb,,,",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name:  "consecutive commas",
			input: "/dev/sda,,/dev/sdb",
			want:  []string{"/dev/sda", "/dev/sdb"},
		},
		// Empty/invalid cases (should preserve original)
		{
			name:         "empty string",
			input:        "",
			wantOriginal: true,
		},
		{
			name:         "only commas",
			input:        ",,,",
			wantOriginal: true,
		},
		{
			name:         "only spaces",
			input:        "   ",
			wantOriginal: true,
		},
		{
			name:         "spaces and commas",
			input:        " , , , ",
			wantOriginal: true,
		},
	}

	originalDisks := []string{"/dev/original"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Storage.Disks = originalDisks

			t.Setenv("DISKS", tt.input)
			LoadFromEnv(cfg)

			if tt.wantOriginal {
				assertDisksEqual(t, cfg.Storage.Disks, originalDisks)
			} else {
				assertDisksEqual(t, cfg.Storage.Disks, tt.want)
			}
		})
	}
}

// TestLoadFromEnvEnumCaseVariations tests case-insensitive parsing
// for BridgeMode and ZFSRaid enum types.
func TestLoadFromEnvEnumCaseVariations(t *testing.T) {
	t.Run("BridgeMode variations", func(t *testing.T) {
		tests := []struct {
			input string
			want  BridgeMode
		}{
			// internal
			{"internal", BridgeModeInternal},
			{"Internal", BridgeModeInternal},
			{"INTERNAL", BridgeModeInternal},
			{"iNtErNaL", BridgeModeInternal},
			// external
			{"external", BridgeModeExternal},
			{"External", BridgeModeExternal},
			{"EXTERNAL", BridgeModeExternal},
			{"eXtErNaL", BridgeModeExternal},
			// both
			{"both", BridgeModeBoth},
			{"Both", BridgeModeBoth},
			{"BOTH", BridgeModeBoth},
			{"bOtH", BridgeModeBoth},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				cfg := DefaultConfig()
				cfg.Network.BridgeMode = "" // Clear default

				t.Setenv("BRIDGE_MODE", tt.input)
				LoadFromEnv(cfg)

				if cfg.Network.BridgeMode != tt.want {
					t.Errorf("BRIDGE_MODE=%q: got %q, want %q", tt.input, cfg.Network.BridgeMode, tt.want)
				}
			})
		}
	})

	t.Run("BridgeMode invalid preserves original", func(t *testing.T) {
		invalidValues := []string{"invalid", "nat", "bridge", "none", "default", ""}

		for _, v := range invalidValues {
			t.Run(v, func(t *testing.T) {
				cfg := DefaultConfig()
				original := cfg.Network.BridgeMode

				t.Setenv("BRIDGE_MODE", v)
				LoadFromEnv(cfg)

				if cfg.Network.BridgeMode != original {
					t.Errorf("BRIDGE_MODE=%q changed config: got %q, want %q", v, cfg.Network.BridgeMode, original)
				}
			})
		}
	})

	t.Run("ZFSRaid variations", func(t *testing.T) {
		tests := []struct {
			input string
			want  ZFSRaid
		}{
			// single
			{"single", ZFSRaidSingle},
			{"Single", ZFSRaidSingle},
			{"SINGLE", ZFSRaidSingle},
			{"sInGlE", ZFSRaidSingle},
			// raid0
			{"raid0", ZFSRaid0},
			{"Raid0", ZFSRaid0},
			{"RAID0", ZFSRaid0},
			{"rAiD0", ZFSRaid0},
			// raid1
			{"raid1", ZFSRaid1},
			{"Raid1", ZFSRaid1},
			{"RAID1", ZFSRaid1},
			{"rAiD1", ZFSRaid1},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				cfg := DefaultConfig()
				// Set to a different value to verify change
				if tt.want == ZFSRaid1 {
					cfg.Storage.ZFSRaid = ZFSRaidSingle
				} else {
					cfg.Storage.ZFSRaid = ZFSRaid1
				}

				t.Setenv("ZFS_RAID", tt.input)
				LoadFromEnv(cfg)

				if cfg.Storage.ZFSRaid != tt.want {
					t.Errorf("ZFS_RAID=%q: got %q, want %q", tt.input, cfg.Storage.ZFSRaid, tt.want)
				}
			})
		}
	})

	t.Run("ZFSRaid invalid preserves original", func(t *testing.T) {
		invalidValues := []string{"invalid", "raid5", "raid6", "raidz", "mirror", "stripe", ""}

		for _, v := range invalidValues {
			t.Run(v, func(t *testing.T) {
				cfg := DefaultConfig()
				original := cfg.Storage.ZFSRaid

				t.Setenv("ZFS_RAID", v)
				LoadFromEnv(cfg)

				if cfg.Storage.ZFSRaid != original {
					t.Errorf("ZFS_RAID=%q changed config: got %q, want %q", v, cfg.Storage.ZFSRaid, original)
				}
			})
		}
	})
}

// clearEnvForTest clears environment variables for test isolation.
// String vars are set to empty, boolean vars are unset and restored on cleanup.
func clearEnvForTest(t *testing.T, envVars []string) {
	t.Helper()

	boolVars := map[string]bool{
		"INSTALL_TAILSCALE": true,
		"TAILSCALE_SSH":     true,
		"TAILSCALE_WEBUI":   true,
	}

	for _, envName := range envVars {
		if boolVars[envName] {
			unsetEnvWithCleanup(t, envName)
		} else {
			t.Setenv(envName, "")
		}
	}
}

// unsetEnvWithCleanup unsets an environment variable and restores it on cleanup.
func unsetEnvWithCleanup(t *testing.T, envName string) {
	t.Helper()

	originalVal, wasSet := os.LookupEnv(envName)
	if !wasSet {
		return
	}

	if err := os.Unsetenv(envName); err != nil {
		t.Fatalf("failed to unset %s: %v", envName, err)
	}

	t.Cleanup(func() {
		// Restore original value. Error ignored as this is cleanup code.
		//nolint:errcheck,usetesting // cleanup code, t.Setenv not available
		os.Setenv(envName, originalVal)
	})
}

// TestLoadFromEnvAllFieldsIndependent verifies that setting one field
// does not affect any other fields in the configuration.
func TestLoadFromEnvAllFieldsIndependent(t *testing.T) {
	// All environment variables used by LoadFromEnv
	allEnvVars := []string{
		"PVE_HOSTNAME", "PVE_DOMAIN_SUFFIX", "PVE_TIMEZONE", "PVE_EMAIL",
		"PVE_ROOT_PASSWORD", "PVE_SSH_PUBLIC_KEY",
		"INTERFACE_NAME", "BRIDGE_MODE", "PRIVATE_SUBNET",
		"ZFS_RAID", "DISKS",
		"INSTALL_TAILSCALE", "TAILSCALE_AUTH_KEY", "TAILSCALE_SSH", "TAILSCALE_WEBUI",
	}

	// List of all environment variables
	envVars := []struct {
		name     string
		value    string
		validate func(*Config) bool
	}{
		{"PVE_HOSTNAME", "test-host", func(c *Config) bool { return c.System.Hostname == "test-host" }},
		{"PVE_DOMAIN_SUFFIX", "test.local", func(c *Config) bool { return c.System.DomainSuffix == "test.local" }},
		{"PVE_TIMEZONE", "UTC", func(c *Config) bool { return c.System.Timezone == "UTC" }},
		{"PVE_EMAIL", "test@test.com", func(c *Config) bool { return c.System.Email == "test@test.com" }},
		{"PVE_ROOT_PASSWORD", "testpass", func(c *Config) bool { return c.System.RootPassword == "testpass" }}, // NOSONAR
		{"PVE_SSH_PUBLIC_KEY", "ssh-rsa test", func(c *Config) bool { return c.System.SSHPublicKey == "ssh-rsa test" }},
		{"INTERFACE_NAME", "eth99", func(c *Config) bool { return c.Network.InterfaceName == "eth99" }},
		{"BRIDGE_MODE", "external", func(c *Config) bool { return c.Network.BridgeMode == BridgeModeExternal }},
		{"PRIVATE_SUBNET", "10.99.0.0/24", func(c *Config) bool { return c.Network.PrivateSubnet == "10.99.0.0/24" }}, // NOSONAR(go:S1313)
		{"ZFS_RAID", "raid0", func(c *Config) bool { return c.Storage.ZFSRaid == ZFSRaid0 }},
		{"DISKS", "/dev/test", func(c *Config) bool { return len(c.Storage.Disks) == 1 && c.Storage.Disks[0] == "/dev/test" }},
		{"INSTALL_TAILSCALE", "true", func(c *Config) bool { return c.Tailscale.Enabled }},
		{"TAILSCALE_AUTH_KEY", "tskey-test", func(c *Config) bool { return c.Tailscale.AuthKey == "tskey-test" }}, // NOSONAR
		{"TAILSCALE_SSH", "true", func(c *Config) bool { return c.Tailscale.SSH }},
		{"TAILSCALE_WEBUI", "true", func(c *Config) bool { return c.Tailscale.WebUI }},
	}

	for i, ev := range envVars {
		t.Run(ev.name, func(t *testing.T) {
			cfg := DefaultConfig()
			// Ensure boolean fields start as false for proper testing
			cfg.Tailscale.Enabled = false
			cfg.Tailscale.SSH = false
			cfg.Tailscale.WebUI = false

			// Clear all env vars first to isolate from external environment.
			// String vars can be set to empty, but boolean vars must be unset.
			clearEnvForTest(t, allEnvVars)

			// Only set this one env var
			t.Setenv(ev.name, ev.value)

			LoadFromEnv(cfg)

			// Verify this field was set correctly
			if !ev.validate(cfg) {
				t.Errorf("%s was not set correctly", ev.name)
			}

			// Verify no other fields were affected
			for j, other := range envVars {
				if i == j {
					continue // Skip the one we set
				}

				// Create a fresh default config to compare
				defaultCfg := DefaultConfig()
				defaultCfg.Tailscale.Enabled = false
				defaultCfg.Tailscale.SSH = false
				defaultCfg.Tailscale.WebUI = false

				// Get the current value and default value for comparison
				switch other.name {
				case "PVE_HOSTNAME":
					if cfg.System.Hostname != defaultCfg.System.Hostname {
						t.Errorf("Setting %s affected System.Hostname", ev.name)
					}
				case "PVE_DOMAIN_SUFFIX":
					if cfg.System.DomainSuffix != defaultCfg.System.DomainSuffix {
						t.Errorf("Setting %s affected System.DomainSuffix", ev.name)
					}
				case "PVE_TIMEZONE":
					if cfg.System.Timezone != defaultCfg.System.Timezone {
						t.Errorf("Setting %s affected System.Timezone", ev.name)
					}
				case "PVE_EMAIL":
					if cfg.System.Email != defaultCfg.System.Email {
						t.Errorf("Setting %s affected System.Email", ev.name)
					}
				case "PVE_ROOT_PASSWORD":
					if cfg.System.RootPassword != defaultCfg.System.RootPassword {
						t.Errorf("Setting %s affected System.RootPassword", ev.name)
					}
				case "PVE_SSH_PUBLIC_KEY":
					if cfg.System.SSHPublicKey != defaultCfg.System.SSHPublicKey {
						t.Errorf("Setting %s affected System.SSHPublicKey", ev.name)
					}
				case "INTERFACE_NAME":
					if cfg.Network.InterfaceName != defaultCfg.Network.InterfaceName {
						t.Errorf("Setting %s affected Network.InterfaceName", ev.name)
					}
				case "BRIDGE_MODE":
					if cfg.Network.BridgeMode != defaultCfg.Network.BridgeMode {
						t.Errorf("Setting %s affected Network.BridgeMode", ev.name)
					}
				case "PRIVATE_SUBNET":
					if cfg.Network.PrivateSubnet != defaultCfg.Network.PrivateSubnet {
						t.Errorf("Setting %s affected Network.PrivateSubnet", ev.name)
					}
				case "ZFS_RAID":
					if cfg.Storage.ZFSRaid != defaultCfg.Storage.ZFSRaid {
						t.Errorf("Setting %s affected Storage.ZFSRaid", ev.name)
					}
				case "DISKS":
					if len(cfg.Storage.Disks) != len(defaultCfg.Storage.Disks) {
						t.Errorf("Setting %s affected Storage.Disks", ev.name)
					}
				case "INSTALL_TAILSCALE":
					if cfg.Tailscale.Enabled != defaultCfg.Tailscale.Enabled {
						t.Errorf("Setting %s affected Tailscale.Enabled", ev.name)
					}
				case "TAILSCALE_AUTH_KEY":
					if cfg.Tailscale.AuthKey != defaultCfg.Tailscale.AuthKey {
						t.Errorf("Setting %s affected Tailscale.AuthKey", ev.name)
					}
				case "TAILSCALE_SSH":
					if cfg.Tailscale.SSH != defaultCfg.Tailscale.SSH {
						t.Errorf("Setting %s affected Tailscale.SSH", ev.name)
					}
				case "TAILSCALE_WEBUI":
					if cfg.Tailscale.WebUI != defaultCfg.Tailscale.WebUI {
						t.Errorf("Setting %s affected Tailscale.WebUI", ev.name)
					}
				}
			}
		})
	}
}
