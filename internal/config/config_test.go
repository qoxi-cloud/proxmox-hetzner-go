package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Test constants for private network subnets (RFC 1918).
// These are intentionally hardcoded for testing network configuration validation.
const (
	testSubnetClassA  = "10.0.0.0/24"    // NOSONAR(go:S1313) Class A private range - test data
	testSubnetClassB  = "172.16.0.0/16"  // NOSONAR(go:S1313) Class B private range - test data
	testSubnetClassC  = "192.168.0.0/24" // NOSONAR(go:S1313) Class C private range - test data
	testSubnetClassC2 = "192.168.1.0/24" // NOSONAR(go:S1313) Class C private range - test data
)

// Test constants for commonly used test values.
// These constants avoid duplication and satisfy SonarCloud code smell checks.
const (
	testTimezoneKyiv      = "Europe/Kyiv"            // NOSONAR(go:S1313) Test timezone data
	testTimezoneNewYork   = "America/New_York"       // NOSONAR(go:S1313) Test timezone data
	testPassword          = "secret-password"        // NOSONAR(go:S1313) Test password - not real credential
	testTailscaleAuthKey  = "tskey-auth-secret123"   // NOSONAR(go:S1313) Test Tailscale key - not real
	testTailscaleAuthKey2 = "tskey-auth-supersecret" // NOSONAR(go:S1313) Test Tailscale key - not real
	testDefaultHostname   = "pve-qoxi-cloud"         // Default hostname per PRD specification
	testHostnamePveServer = "pve-server"             // Common test hostname
)

func TestSystemConfig_SensitiveFieldsOmittedFromYAML(t *testing.T) {
	tests := []struct {
		name             string
		cfg              SystemConfig
		shouldNotContain []string
		shouldContain    []string
	}{
		{
			name: "standard config with all fields",
			cfg: SystemConfig{
				Hostname:     testHostnamePveServer,
				DomainSuffix: "local",
				Timezone:     testTimezoneKyiv,
				Email:        "admin@example.com",
				RootPassword: testPassword,
				SSHPublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG...",
			},
			shouldNotContain: []string{testPassword, "ssh-ed25519", "root_password", "ssh_public_key"},
			shouldContain:    []string{"hostname: " + testHostnamePveServer, "domain_suffix: local", "timezone: " + testTimezoneKyiv, "email: admin@example.com"},
		},
		{
			name: "config with special characters in sensitive fields",
			cfg: SystemConfig{
				Hostname:     "test-host",
				DomainSuffix: "example.com",
				Timezone:     "UTC",
				Email:        "test@test.com",
				RootPassword: "p@ssw0rd!#$%",
				SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB...",
			},
			shouldNotContain: []string{"p@ssw0rd!#$%", "ssh-rsa AAAAB3", "root_password", "ssh_public_key"},
			shouldContain:    []string{"hostname: test-host", "domain_suffix: example.com"},
		},
		{
			name: "config with empty sensitive fields",
			cfg: SystemConfig{
				Hostname:     "empty-secrets",
				DomainSuffix: "local",
				Timezone:     "UTC",
				Email:        "admin@local",
				RootPassword: "",
				SSHPublicKey: "",
			},
			shouldNotContain: []string{"root_password", "ssh_public_key"},
			shouldContain:    []string{"hostname: empty-secrets"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			yamlStr := string(data)

			for _, str := range tt.shouldNotContain {
				assert.NotContains(t, yamlStr, str)
			}
			for _, str := range tt.shouldContain {
				assert.Contains(t, yamlStr, str)
			}
		})
	}
}

func TestSystemConfig_StandardFieldsSerializeCorrectly(t *testing.T) {
	tests := []struct {
		name string
		cfg  SystemConfig
	}{
		{
			name: "standard values",
			cfg: SystemConfig{
				Hostname:     "test-server",
				DomainSuffix: "example.com",
				Timezone:     "UTC",
				Email:        "test@example.com",
			},
		},
		{
			name: "empty strings",
			cfg: SystemConfig{
				Hostname:     "",
				DomainSuffix: "",
				Timezone:     "",
				Email:        "",
			},
		},
		{
			name: "special characters",
			cfg: SystemConfig{
				Hostname:     "pve-server-01",
				DomainSuffix: "sub.domain.example.com",
				Timezone:     testTimezoneNewYork,
				Email:        "admin+alerts@example.com",
			},
		},
		{
			name: "unicode characters",
			cfg: SystemConfig{
				Hostname:     "server",
				DomainSuffix: "example.com",
				Timezone:     testTimezoneKyiv,
				Email:        "user@example.com",
			},
		},
		{
			name: "long values",
			cfg: SystemConfig{
				Hostname:     "very-long-hostname-that-is-still-valid",
				DomainSuffix: "subdomain.another.level.example.com",
				Timezone:     "America/Argentina/Buenos_Aires",
				Email:        "very.long.email.address@subdomain.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			var result SystemConfig
			err = yaml.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.cfg.Hostname, result.Hostname)
			assert.Equal(t, tt.cfg.DomainSuffix, result.DomainSuffix)
			assert.Equal(t, tt.cfg.Timezone, result.Timezone)
			assert.Equal(t, tt.cfg.Email, result.Email)
		})
	}
}

func TestSystemConfig_RoundTripMarshalUnmarshal(t *testing.T) {
	original := SystemConfig{
		Hostname:     "production-pve",
		DomainSuffix: "prod.example.com",
		Timezone:     testTimezoneNewYork,
		Email:        "ops@company.com",
		RootPassword: "super-secret",
		SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&original)
	require.NoError(t, err)

	// Unmarshal back
	var restored SystemConfig
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Non-sensitive fields should be restored
	assert.Equal(t, original.Hostname, restored.Hostname)
	assert.Equal(t, original.DomainSuffix, restored.DomainSuffix)
	assert.Equal(t, original.Timezone, restored.Timezone)
	assert.Equal(t, original.Email, restored.Email)

	// Sensitive fields should be empty after round-trip (they were excluded)
	assert.Empty(t, restored.RootPassword)
	assert.Empty(t, restored.SSHPublicKey)
}

func TestSystemConfig_EnvironmentVariableTagsPresent(t *testing.T) {
	expectedEnvTags := map[string]string{
		"Hostname":     "PVE_HOSTNAME",
		"DomainSuffix": "PVE_DOMAIN_SUFFIX",
		"Timezone":     "PVE_TIMEZONE",
		"Email":        "PVE_EMAIL",
		"RootPassword": "PVE_ROOT_PASSWORD",
		"SSHPublicKey": "PVE_SSH_PUBLIC_KEY",
	}

	cfgType := reflect.TypeOf(SystemConfig{})

	for fieldName, expectedTag := range expectedEnvTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		envTag := field.Tag.Get("env")
		assert.Equal(t, expectedTag, envTag, "env tag mismatch for field %s", fieldName)
	}
}

func TestSystemConfig_YAMLTagsPresent(t *testing.T) {
	expectedYAMLTags := map[string]string{
		"Hostname":     "hostname",
		"DomainSuffix": "domain_suffix",
		"Timezone":     "timezone",
		"Email":        "email",
		"RootPassword": "-",
		"SSHPublicKey": "-",
	}

	cfgType := reflect.TypeOf(SystemConfig{})

	for fieldName, expectedTag := range expectedYAMLTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		yamlTag := field.Tag.Get("yaml")
		assert.Equal(t, expectedTag, yamlTag, "yaml tag mismatch for field %s", fieldName)
	}
}

func TestSystemConfig_AllFieldsExist(t *testing.T) {
	requiredFields := []string{
		"Hostname",
		"DomainSuffix",
		"Timezone",
		"Email",
		"RootPassword",
		"SSHPublicKey",
	}

	cfgType := reflect.TypeOf(SystemConfig{})

	assert.Equal(t, len(requiredFields), cfgType.NumField(), "unexpected number of fields")

	for _, fieldName := range requiredFields {
		field, found := cfgType.FieldByName(fieldName)
		assert.True(t, found, "required field %s not found", fieldName)
		assert.Equal(t, "string", field.Type.Kind().String(), "field %s should be string type", fieldName)
	}
}

// NetworkConfig tests

func TestNetworkConfig_BridgeModeSerializesToYAML(t *testing.T) {
	tests := []struct {
		name         string
		cfg          NetworkConfig
		expectedYAML string
	}{
		{
			name: "internal bridge mode",
			cfg: NetworkConfig{
				InterfaceName: "eth0",
				BridgeMode:    BridgeModeInternal,
				PrivateSubnet: testSubnetClassA,
			},
			expectedYAML: "bridge_mode: internal",
		},
		{
			name: "external bridge mode",
			cfg: NetworkConfig{
				InterfaceName: "enp0s31f6",
				BridgeMode:    BridgeModeExternal,
				PrivateSubnet: testSubnetClassC2,
			},
			expectedYAML: "bridge_mode: external",
		},
		{
			name: "both bridge mode",
			cfg: NetworkConfig{
				InterfaceName: "eth0",
				BridgeMode:    BridgeModeBoth,
				PrivateSubnet: testSubnetClassB,
			},
			expectedYAML: "bridge_mode: both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			yamlStr := string(data)
			assert.Contains(t, yamlStr, tt.expectedYAML)
		})
	}
}

func TestNetworkConfig_DeserializeFromYAML(t *testing.T) {
	tests := []struct {
		name           string
		yamlInput      string
		expectedMode   BridgeMode
		expectedIface  string
		expectedSubnet string
	}{
		{
			name: "valid internal mode",
			yamlInput: "interface: eth0\nbridge_mode: internal\nprivate_subnet: \"" +
				testSubnetClassA + "\"",
			expectedMode:   BridgeModeInternal,
			expectedIface:  "eth0",
			expectedSubnet: testSubnetClassA,
		},
		{
			name: "valid external mode",
			yamlInput: "interface: enp0s31f6\nbridge_mode: external\nprivate_subnet: \"" +
				testSubnetClassC + "\"",
			expectedMode:   BridgeModeExternal,
			expectedIface:  "enp0s31f6",
			expectedSubnet: testSubnetClassC,
		},
		{
			name: "valid both mode",
			yamlInput: "interface: eth1\nbridge_mode: both\nprivate_subnet: \"" +
				testSubnetClassB + "\"",
			expectedMode:   BridgeModeBoth,
			expectedIface:  "eth1",
			expectedSubnet: testSubnetClassB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg NetworkConfig
			err := yaml.Unmarshal([]byte(tt.yamlInput), &cfg)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedMode, cfg.BridgeMode)
			assert.Equal(t, tt.expectedIface, cfg.InterfaceName)
			assert.Equal(t, tt.expectedSubnet, cfg.PrivateSubnet)
		})
	}
}

func TestNetworkConfig_DeserializeInvalidBridgeMode(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
	}{
		{
			name: "invalid bridge mode value",
			yamlInput: `interface: eth0
bridge_mode: invalid_mode
private_subnet: "10.0.0.0/24"`,
		},
		{
			name: "empty bridge mode",
			yamlInput: `interface: eth0
bridge_mode: ""
private_subnet: "10.0.0.0/24"`,
		},
		{
			name: "missing bridge mode",
			yamlInput: `interface: eth0
private_subnet: "10.0.0.0/24"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg NetworkConfig
			err := yaml.Unmarshal([]byte(tt.yamlInput), &cfg)
			require.NoError(t, err)

			// Invalid or missing values result in empty/default BridgeMode
			assert.False(t, cfg.BridgeMode.IsValid())
		})
	}
}

func TestNetworkConfig_EnvironmentVariableTagsPresent(t *testing.T) {
	expectedEnvTags := map[string]string{
		"InterfaceName": "INTERFACE_NAME",
		"BridgeMode":    "BRIDGE_MODE",
		"PrivateSubnet": "PRIVATE_SUBNET",
	}

	cfgType := reflect.TypeOf(NetworkConfig{})

	for fieldName, expectedTag := range expectedEnvTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		envTag := field.Tag.Get("env")
		assert.Equal(t, expectedTag, envTag, "env tag mismatch for field %s", fieldName)
	}
}

func TestNetworkConfig_YAMLTagsPresent(t *testing.T) {
	expectedYAMLTags := map[string]string{
		"InterfaceName": "interface",
		"BridgeMode":    "bridge_mode",
		"PrivateSubnet": "private_subnet",
	}

	cfgType := reflect.TypeOf(NetworkConfig{})

	for fieldName, expectedTag := range expectedYAMLTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		yamlTag := field.Tag.Get("yaml")
		assert.Equal(t, expectedTag, yamlTag, "yaml tag mismatch for field %s", fieldName)
	}
}

func TestNetworkConfig_AllFieldsExist(t *testing.T) {
	expectedFields := map[string]string{
		"InterfaceName": "string",
		"BridgeMode":    "BridgeMode",
		"PrivateSubnet": "string",
	}

	cfgType := reflect.TypeOf(NetworkConfig{})
	assert.Equal(t, len(expectedFields), cfgType.NumField(), "unexpected number of fields")

	for fieldName, expectedType := range expectedFields {
		field, found := cfgType.FieldByName(fieldName)
		assert.True(t, found, "required field %s not found", fieldName)
		assert.Equal(t, expectedType, field.Type.Name(), "field %s type mismatch", fieldName)
	}
}

func TestNetworkConfig_RoundTripMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		cfg  NetworkConfig
	}{
		{
			name: "standard config",
			cfg: NetworkConfig{
				InterfaceName: "eth0",
				BridgeMode:    BridgeModeInternal,
				PrivateSubnet: testSubnetClassA,
			},
		},
		{
			name: "external mode config",
			cfg: NetworkConfig{
				InterfaceName: "enp0s31f6",
				BridgeMode:    BridgeModeExternal,
				PrivateSubnet: testSubnetClassC2,
			},
		},
		{
			name: "both mode config",
			cfg: NetworkConfig{
				InterfaceName: "eth1",
				BridgeMode:    BridgeModeBoth,
				PrivateSubnet: testSubnetClassB,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			var restored NetworkConfig
			err = yaml.Unmarshal(data, &restored)
			require.NoError(t, err)

			assert.Equal(t, tt.cfg.InterfaceName, restored.InterfaceName)
			assert.Equal(t, tt.cfg.BridgeMode, restored.BridgeMode)
			assert.Equal(t, tt.cfg.PrivateSubnet, restored.PrivateSubnet)
		})
	}
}

// StorageConfig tests

func TestStorageConfig_ZFSRaidSerializesToYAML(t *testing.T) {
	tests := []struct {
		name         string
		cfg          StorageConfig
		expectedYAML string
	}{
		{
			name: "single disk configuration",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaidSingle,
				Disks:   []string{"/dev/sda"},
			},
			expectedYAML: "zfs_raid: single",
		},
		{
			name: "raid0 configuration",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaid0,
				Disks:   []string{"/dev/sda", "/dev/sdb"},
			},
			expectedYAML: "zfs_raid: raid0",
		},
		{
			name: "raid1 configuration",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaid1,
				Disks:   []string{"/dev/sda", "/dev/sdb"},
			},
			expectedYAML: "zfs_raid: raid1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			yamlStr := string(data)
			assert.Contains(t, yamlStr, tt.expectedYAML)
		})
	}
}

func TestStorageConfig_DisksArraySerializesToYAML(t *testing.T) {
	tests := []struct {
		name      string
		cfg       StorageConfig
		wantDisks []string
	}{
		{
			name: "single disk",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaidSingle,
				Disks:   []string{"/dev/sda"},
			},
			wantDisks: []string{"/dev/sda"},
		},
		{
			name: "multiple disks",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaid1,
				Disks:   []string{"/dev/sda", "/dev/sdb"},
			},
			wantDisks: []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name: "many disks",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaid0,
				Disks:   []string{"/dev/nvme0n1", "/dev/nvme1n1", "/dev/nvme2n1"},
			},
			wantDisks: []string{"/dev/nvme0n1", "/dev/nvme1n1", "/dev/nvme2n1"},
		},
		{
			name: "empty disk list",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaidSingle,
				Disks:   []string{},
			},
			wantDisks: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			var restored StorageConfig
			err = yaml.Unmarshal(data, &restored)
			require.NoError(t, err)

			assert.Equal(t, tt.wantDisks, restored.Disks)
		})
	}
}

func TestStorageConfig_DeserializeFromYAML(t *testing.T) {
	tests := []struct {
		name          string
		yamlInput     string
		expectedRaid  ZFSRaid
		expectedDisks []string
	}{
		{
			name: "single disk config",
			yamlInput: `zfs_raid: single
disks:
  - /dev/sda`,
			expectedRaid:  ZFSRaidSingle,
			expectedDisks: []string{"/dev/sda"},
		},
		{
			name: "raid0 with two disks",
			yamlInput: `zfs_raid: raid0
disks:
  - /dev/sda
  - /dev/sdb`,
			expectedRaid:  ZFSRaid0,
			expectedDisks: []string{"/dev/sda", "/dev/sdb"},
		},
		{
			name: "raid1 with nvme disks",
			yamlInput: `zfs_raid: raid1
disks:
  - /dev/nvme0n1
  - /dev/nvme1n1`,
			expectedRaid:  ZFSRaid1,
			expectedDisks: []string{"/dev/nvme0n1", "/dev/nvme1n1"},
		},
		{
			name: "empty disks",
			yamlInput: `zfs_raid: single
disks: []`,
			expectedRaid:  ZFSRaidSingle,
			expectedDisks: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg StorageConfig
			err := yaml.Unmarshal([]byte(tt.yamlInput), &cfg)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedRaid, cfg.ZFSRaid)
			assert.Equal(t, tt.expectedDisks, cfg.Disks)
		})
	}
}

func TestStorageConfig_DeserializeInvalidZFSRaid(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
	}{
		{
			name: "invalid raid value",
			yamlInput: `zfs_raid: invalid_raid
disks:
  - /dev/sda`,
		},
		{
			name: "empty raid value",
			yamlInput: `zfs_raid: ""
disks:
  - /dev/sda`,
		},
		{
			name: "missing raid value",
			yamlInput: `disks:
  - /dev/sda`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg StorageConfig
			err := yaml.Unmarshal([]byte(tt.yamlInput), &cfg)
			require.NoError(t, err)

			// Invalid or missing values result in empty/default ZFSRaid
			assert.False(t, cfg.ZFSRaid.IsValid())
		})
	}
}

func TestStorageConfig_EnvironmentVariableTagsPresent(t *testing.T) {
	expectedEnvTags := map[string]string{
		"ZFSRaid": "ZFS_RAID",
		"Disks":   "DISKS",
	}

	cfgType := reflect.TypeOf(StorageConfig{})

	for fieldName, expectedTag := range expectedEnvTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		envTag := field.Tag.Get("env")
		assert.Equal(t, expectedTag, envTag, "env tag mismatch for field %s", fieldName)
	}
}

func TestStorageConfig_YAMLTagsPresent(t *testing.T) {
	expectedYAMLTags := map[string]string{
		"ZFSRaid": "zfs_raid",
		"Disks":   "disks",
	}

	cfgType := reflect.TypeOf(StorageConfig{})

	for fieldName, expectedTag := range expectedYAMLTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		yamlTag := field.Tag.Get("yaml")
		assert.Equal(t, expectedTag, yamlTag, "yaml tag mismatch for field %s", fieldName)
	}
}

func TestStorageConfig_AllFieldsExist(t *testing.T) {
	expectedFields := map[string]string{
		"ZFSRaid": "ZFSRaid",
		"Disks":   "slice",
	}

	cfgType := reflect.TypeOf(StorageConfig{})
	assert.Equal(t, len(expectedFields), cfgType.NumField(), "unexpected number of fields")

	for fieldName, expectedType := range expectedFields {
		field, found := cfgType.FieldByName(fieldName)
		assert.True(t, found, "required field %s not found", fieldName)
		if expectedType == "slice" {
			assert.Equal(t, reflect.Slice, field.Type.Kind(), "field %s should be slice type", fieldName)
		} else {
			assert.Equal(t, expectedType, field.Type.Name(), "field %s type mismatch", fieldName)
		}
	}
}

func TestStorageConfig_RoundTripMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		cfg  StorageConfig
	}{
		{
			name: "single disk config",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaidSingle,
				Disks:   []string{"/dev/sda"},
			},
		},
		{
			name: "raid0 config",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaid0,
				Disks:   []string{"/dev/sda", "/dev/sdb"},
			},
		},
		{
			name: "raid1 with nvme",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaid1,
				Disks:   []string{"/dev/nvme0n1", "/dev/nvme1n1"},
			},
		},
		{
			name: "empty disks",
			cfg: StorageConfig{
				ZFSRaid: ZFSRaidSingle,
				Disks:   []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			var restored StorageConfig
			err = yaml.Unmarshal(data, &restored)
			require.NoError(t, err)

			assert.Equal(t, tt.cfg.ZFSRaid, restored.ZFSRaid)
			assert.Equal(t, tt.cfg.Disks, restored.Disks)
		})
	}
}

func TestStorageConfig_EnvSeparatorTagPresent(t *testing.T) {
	cfgType := reflect.TypeOf(StorageConfig{})

	field, found := cfgType.FieldByName("Disks")
	require.True(t, found, "field Disks not found")

	envSeparatorTag := field.Tag.Get("envSeparator")
	assert.Equal(t, ",", envSeparatorTag, "envSeparator tag should be comma")
}

// TailscaleConfig tests

func TestTailscaleConfig_AuthKeyOmittedFromYAML(t *testing.T) {
	tests := []struct {
		name             string
		cfg              TailscaleConfig
		shouldNotContain []string
		shouldContain    []string
	}{
		{
			name: "standard config with auth key",
			cfg: TailscaleConfig{
				Enabled: true,
				AuthKey: testTailscaleAuthKey,
				SSH:     true,
				WebUI:   true,
			},
			shouldNotContain: []string{testTailscaleAuthKey, "auth_key"},
			shouldContain:    []string{"enabled: true", "ssh: true", "webui: true"},
		},
		{
			name: "disabled config with auth key",
			cfg: TailscaleConfig{
				Enabled: false,
				AuthKey: "tskey-auth-anothersecret",
				SSH:     false,
				WebUI:   false,
			},
			shouldNotContain: []string{"tskey-auth-anothersecret", "auth_key"},
			shouldContain:    []string{"enabled: false", "ssh: false", "webui: false"},
		},
		{
			name: "config with empty auth key",
			cfg: TailscaleConfig{
				Enabled: true,
				AuthKey: "",
				SSH:     true,
				WebUI:   false,
			},
			shouldNotContain: []string{"auth_key"},
			shouldContain:    []string{"enabled: true", "ssh: true", "webui: false"},
		},
		{
			name: "config with special characters in auth key",
			cfg: TailscaleConfig{
				Enabled: true,
				AuthKey: "tskey-auth-!@#$%^&*()_+-=[]{}|;':\",./<>?",
				SSH:     false,
				WebUI:   true,
			},
			shouldNotContain: []string{"tskey-auth-!@#$%^&*()_+-=[]{}|;':\",./<>?", "auth_key"},
			shouldContain:    []string{"enabled: true", "webui: true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			yamlStr := string(data)

			for _, str := range tt.shouldNotContain {
				assert.NotContains(t, yamlStr, str)
			}
			for _, str := range tt.shouldContain {
				assert.Contains(t, yamlStr, str)
			}
		})
	}
}

func TestTailscaleConfig_BooleanFieldsSerializeCorrectly(t *testing.T) {
	tests := []struct {
		name string
		cfg  TailscaleConfig
	}{
		{
			name: "all enabled",
			cfg: TailscaleConfig{
				Enabled: true,
				SSH:     true,
				WebUI:   true,
			},
		},
		{
			name: "all disabled",
			cfg: TailscaleConfig{
				Enabled: false,
				SSH:     false,
				WebUI:   false,
			},
		},
		{
			name: "mixed enabled states - SSH only",
			cfg: TailscaleConfig{
				Enabled: true,
				SSH:     true,
				WebUI:   false,
			},
		},
		{
			name: "mixed enabled states - WebUI only",
			cfg: TailscaleConfig{
				Enabled: true,
				SSH:     false,
				WebUI:   true,
			},
		},
		{
			name: "disabled with options enabled",
			cfg: TailscaleConfig{
				Enabled: false,
				SSH:     true,
				WebUI:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			var result TailscaleConfig
			err = yaml.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.cfg.Enabled, result.Enabled)
			assert.Equal(t, tt.cfg.SSH, result.SSH)
			assert.Equal(t, tt.cfg.WebUI, result.WebUI)
		})
	}
}

func TestTailscaleConfig_DeserializeFromYAML(t *testing.T) {
	tests := []struct {
		name        string
		yamlInput   string
		expectedCfg TailscaleConfig
	}{
		{
			name: "fully enabled config",
			yamlInput: `enabled: true
ssh: true
webui: true`,
			expectedCfg: TailscaleConfig{
				Enabled: true,
				SSH:     true,
				WebUI:   true,
			},
		},
		{
			name: "fully disabled config",
			yamlInput: `enabled: false
ssh: false
webui: false`,
			expectedCfg: TailscaleConfig{
				Enabled: false,
				SSH:     false,
				WebUI:   false,
			},
		},
		{
			name: "enabled with SSH only",
			yamlInput: `enabled: true
ssh: true
webui: false`,
			expectedCfg: TailscaleConfig{
				Enabled: true,
				SSH:     true,
				WebUI:   false,
			},
		},
		{
			name: "enabled with WebUI only",
			yamlInput: `enabled: true
ssh: false
webui: true`,
			expectedCfg: TailscaleConfig{
				Enabled: true,
				SSH:     false,
				WebUI:   true,
			},
		},
		{
			name:      "empty config defaults to false",
			yamlInput: `{}`,
			expectedCfg: TailscaleConfig{
				Enabled: false,
				SSH:     false,
				WebUI:   false,
			},
		},
		{
			name:      "partial config - only enabled",
			yamlInput: `enabled: true`,
			expectedCfg: TailscaleConfig{
				Enabled: true,
				SSH:     false,
				WebUI:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg TailscaleConfig
			err := yaml.Unmarshal([]byte(tt.yamlInput), &cfg)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCfg.Enabled, cfg.Enabled)
			assert.Equal(t, tt.expectedCfg.SSH, cfg.SSH)
			assert.Equal(t, tt.expectedCfg.WebUI, cfg.WebUI)
			// AuthKey should always be empty after deserialization (excluded from YAML)
			assert.Empty(t, cfg.AuthKey)
		})
	}
}

func TestTailscaleConfig_RoundTripMarshalUnmarshal(t *testing.T) {
	original := TailscaleConfig{
		Enabled: true,
		AuthKey: testTailscaleAuthKey2,
		SSH:     true,
		WebUI:   false,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&original)
	require.NoError(t, err)

	// Unmarshal back
	var restored TailscaleConfig
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Non-sensitive fields should be restored
	assert.Equal(t, original.Enabled, restored.Enabled)
	assert.Equal(t, original.SSH, restored.SSH)
	assert.Equal(t, original.WebUI, restored.WebUI)

	// Sensitive field (AuthKey) should be empty after round-trip
	assert.Empty(t, restored.AuthKey)
}

func TestTailscaleConfig_EnvironmentVariableTagsPresent(t *testing.T) {
	expectedEnvTags := map[string]string{
		"Enabled": "INSTALL_TAILSCALE",
		"AuthKey": "TAILSCALE_AUTH_KEY",
		"SSH":     "TAILSCALE_SSH",
		"WebUI":   "TAILSCALE_WEBUI",
	}

	cfgType := reflect.TypeOf(TailscaleConfig{})

	for fieldName, expectedTag := range expectedEnvTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		envTag := field.Tag.Get("env")
		assert.Equal(t, expectedTag, envTag, "env tag mismatch for field %s", fieldName)
	}
}

func TestTailscaleConfig_YAMLTagsPresent(t *testing.T) {
	expectedYAMLTags := map[string]string{
		"Enabled": "enabled",
		"AuthKey": "-",
		"SSH":     "ssh",
		"WebUI":   "webui",
	}

	cfgType := reflect.TypeOf(TailscaleConfig{})

	for fieldName, expectedTag := range expectedYAMLTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		yamlTag := field.Tag.Get("yaml")
		assert.Equal(t, expectedTag, yamlTag, "yaml tag mismatch for field %s", fieldName)
	}
}

func TestTailscaleConfig_AllFieldsExist(t *testing.T) {
	expectedFields := map[string]string{
		"Enabled": "bool",
		"AuthKey": "string",
		"SSH":     "bool",
		"WebUI":   "bool",
	}

	cfgType := reflect.TypeOf(TailscaleConfig{})
	assert.Equal(t, len(expectedFields), cfgType.NumField(), "unexpected number of fields")

	for fieldName, expectedType := range expectedFields {
		field, found := cfgType.FieldByName(fieldName)
		assert.True(t, found, "required field %s not found", fieldName)
		assert.Equal(t, expectedType, field.Type.Kind().String(), "field %s type mismatch", fieldName)
	}
}

func TestTailscaleConfig_AllSSHWebUICombinations(t *testing.T) {
	// Test all combinations of SSH and WebUI flags
	combinations := []struct {
		ssh   bool
		webui bool
	}{
		{false, false},
		{false, true},
		{true, false},
		{true, true},
	}

	for _, combo := range combinations {
		t.Run("SSH="+boolToStr(combo.ssh)+"/WebUI="+boolToStr(combo.webui), func(t *testing.T) {
			cfg := TailscaleConfig{
				Enabled: true,
				AuthKey: "test-key",
				SSH:     combo.ssh,
				WebUI:   combo.webui,
			}

			// Marshal to YAML
			data, err := yaml.Marshal(&cfg)
			require.NoError(t, err)

			// Unmarshal back
			var restored TailscaleConfig
			err = yaml.Unmarshal(data, &restored)
			require.NoError(t, err)

			// Verify the flags are preserved
			assert.Equal(t, combo.ssh, restored.SSH)
			assert.Equal(t, combo.webui, restored.WebUI)
		})
	}
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Config tests

func TestConfig_NestedStructsSerializeCorrectly(t *testing.T) {
	cfg := Config{
		System: SystemConfig{
			Hostname:     testHostnamePveServer,
			DomainSuffix: "local",
			Timezone:     testTimezoneKyiv,
			Email:        "admin@example.com",
			RootPassword: testPassword,
			SSHPublicKey: "ssh-ed25519 AAAAC3...",
		},
		Network: NetworkConfig{
			InterfaceName: "eth0",
			BridgeMode:    BridgeModeInternal,
			PrivateSubnet: testSubnetClassA,
		},
		Storage: StorageConfig{
			ZFSRaid: ZFSRaid1,
			Disks:   []string{"/dev/sda", "/dev/sdb"},
		},
		Tailscale: TailscaleConfig{
			Enabled: true,
			AuthKey: testTailscaleAuthKey,
			SSH:     true,
			WebUI:   false,
		},
		Verbose: true,
	}

	data, err := yaml.Marshal(&cfg)
	require.NoError(t, err)

	yamlStr := string(data)

	// Check nested structure serialization
	assert.Contains(t, yamlStr, "system:")
	assert.Contains(t, yamlStr, "network:")
	assert.Contains(t, yamlStr, "storage:")
	assert.Contains(t, yamlStr, "tailscale:")

	// Check nested values
	assert.Contains(t, yamlStr, "hostname: "+testHostnamePveServer)
	assert.Contains(t, yamlStr, "interface: eth0")
	assert.Contains(t, yamlStr, "zfs_raid: raid1")
	assert.Contains(t, yamlStr, "enabled: true")
}

func TestConfig_DeserializeFromYAML(t *testing.T) {
	yamlInput := `system:
  hostname: test-server
  domain_suffix: example.com
  timezone: UTC
  email: test@test.com
network:
  interface: enp0s31f6
  bridge_mode: external
  private_subnet: "192.168.0.0/24"
storage:
  zfs_raid: single
  disks:
    - /dev/nvme0n1
tailscale:
  enabled: true
  ssh: true
  webui: false`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlInput), &cfg)
	require.NoError(t, err)

	assert.Equal(t, "test-server", cfg.System.Hostname)
	assert.Equal(t, "example.com", cfg.System.DomainSuffix)
	assert.Equal(t, "UTC", cfg.System.Timezone)
	assert.Equal(t, "test@test.com", cfg.System.Email)

	assert.Equal(t, "enp0s31f6", cfg.Network.InterfaceName)
	assert.Equal(t, BridgeModeExternal, cfg.Network.BridgeMode)
	assert.Equal(t, testSubnetClassC, cfg.Network.PrivateSubnet)

	assert.Equal(t, ZFSRaidSingle, cfg.Storage.ZFSRaid)
	assert.Equal(t, []string{"/dev/nvme0n1"}, cfg.Storage.Disks)

	assert.True(t, cfg.Tailscale.Enabled)
	assert.True(t, cfg.Tailscale.SSH)
	assert.False(t, cfg.Tailscale.WebUI)

	// Verbose should be false (not in YAML)
	assert.False(t, cfg.Verbose)
}

func TestConfig_VerboseExcludedFromYAML(t *testing.T) {
	cfg := Config{
		System: SystemConfig{
			Hostname: "test",
		},
		Verbose: true,
	}

	data, err := yaml.Marshal(&cfg)
	require.NoError(t, err)

	yamlStr := string(data)
	assert.NotContains(t, yamlStr, "verbose")
}

func TestConfig_SensitiveFieldsNotSerialized(t *testing.T) {
	cfg := Config{
		System: SystemConfig{
			Hostname:     "pve",
			RootPassword: "super-secret-password",
			SSHPublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5...",
		},
		Tailscale: TailscaleConfig{
			Enabled: true,
			AuthKey: testTailscaleAuthKey2,
		},
	}

	data, err := yaml.Marshal(&cfg)
	require.NoError(t, err)

	yamlStr := string(data)
	assert.NotContains(t, yamlStr, "super-secret-password")
	assert.NotContains(t, yamlStr, "ssh-ed25519")
	assert.NotContains(t, yamlStr, testTailscaleAuthKey2)
	assert.NotContains(t, yamlStr, "root_password")
	assert.NotContains(t, yamlStr, "ssh_public_key")
	assert.NotContains(t, yamlStr, "auth_key")
}

func TestConfig_PartialConfigDeserialize(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
	}{
		{
			name:      "only system section",
			yamlInput: "system:\n  hostname: test",
		},
		{
			name:      "only network section",
			yamlInput: "network:\n  interface: eth0",
		},
		{
			name:      "only storage section",
			yamlInput: "storage:\n  zfs_raid: single",
		},
		{
			name:      "only tailscale section",
			yamlInput: "tailscale:\n  enabled: true",
		},
		{
			name:      "empty config",
			yamlInput: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := yaml.Unmarshal([]byte(tt.yamlInput), &cfg)
			require.NoError(t, err)
			// Should not error on partial configs - missing sections use zero values
		})
	}
}

func TestConfig_RoundTripMarshalUnmarshal(t *testing.T) {
	original := Config{
		System: SystemConfig{
			Hostname:     "production-pve",
			DomainSuffix: "prod.example.com",
			Timezone:     testTimezoneNewYork,
			Email:        "ops@company.com",
			RootPassword: "secret",
			SSHPublicKey: "ssh-rsa AAAAB...",
		},
		Network: NetworkConfig{
			InterfaceName: "eth0",
			BridgeMode:    BridgeModeBoth,
			PrivateSubnet: testSubnetClassA,
		},
		Storage: StorageConfig{
			ZFSRaid: ZFSRaid1,
			Disks:   []string{"/dev/sda", "/dev/sdb"},
		},
		Tailscale: TailscaleConfig{
			Enabled: true,
			AuthKey: "tskey-secret",
			SSH:     true,
			WebUI:   true,
		},
		Verbose: true,
	}

	data, err := yaml.Marshal(&original)
	require.NoError(t, err)

	var restored Config
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Non-sensitive fields should be restored
	assert.Equal(t, original.System.Hostname, restored.System.Hostname)
	assert.Equal(t, original.System.DomainSuffix, restored.System.DomainSuffix)
	assert.Equal(t, original.System.Timezone, restored.System.Timezone)
	assert.Equal(t, original.System.Email, restored.System.Email)

	assert.Equal(t, original.Network.InterfaceName, restored.Network.InterfaceName)
	assert.Equal(t, original.Network.BridgeMode, restored.Network.BridgeMode)
	assert.Equal(t, original.Network.PrivateSubnet, restored.Network.PrivateSubnet)

	assert.Equal(t, original.Storage.ZFSRaid, restored.Storage.ZFSRaid)
	assert.Equal(t, original.Storage.Disks, restored.Storage.Disks)

	assert.Equal(t, original.Tailscale.Enabled, restored.Tailscale.Enabled)
	assert.Equal(t, original.Tailscale.SSH, restored.Tailscale.SSH)
	assert.Equal(t, original.Tailscale.WebUI, restored.Tailscale.WebUI)

	// Sensitive fields should be empty after round-trip
	assert.Empty(t, restored.System.RootPassword)
	assert.Empty(t, restored.System.SSHPublicKey)
	assert.Empty(t, restored.Tailscale.AuthKey)

	// Verbose should be false (excluded from YAML)
	assert.False(t, restored.Verbose)
}

func TestConfig_YAMLTagsPresent(t *testing.T) {
	expectedYAMLTags := map[string]string{
		"System":    "system",
		"Network":   "network",
		"Storage":   "storage",
		"Tailscale": "tailscale",
		"Verbose":   "-",
	}

	cfgType := reflect.TypeOf(Config{})

	for fieldName, expectedTag := range expectedYAMLTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		yamlTag := field.Tag.Get("yaml")
		assert.Equal(t, expectedTag, yamlTag, "yaml tag mismatch for field %s", fieldName)
	}
}

func TestConfig_AllFieldsExist(t *testing.T) {
	expectedFields := map[string]string{
		"System":    "SystemConfig",
		"Network":   "NetworkConfig",
		"Storage":   "StorageConfig",
		"Tailscale": "TailscaleConfig",
		"Verbose":   "bool",
	}

	cfgType := reflect.TypeOf(Config{})
	assert.Equal(t, len(expectedFields), cfgType.NumField(), "unexpected number of fields")

	for fieldName, expectedType := range expectedFields {
		field, found := cfgType.FieldByName(fieldName)
		assert.True(t, found, "required field %s not found", fieldName)
		if expectedType == "bool" {
			assert.Equal(t, "bool", field.Type.Kind().String(), "field %s type mismatch", fieldName)
		} else {
			assert.Equal(t, expectedType, field.Type.Name(), "field %s type mismatch", fieldName)
		}
	}
}

// DefaultConfig tests

func TestDefaultConfig_ReturnsValidPointer(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)
}

func TestDefaultConfig_ReturnsNewInstanceEachCall(t *testing.T) {
	cfg1 := DefaultConfig()
	cfg2 := DefaultConfig()

	// Should be different pointers
	require.NotSame(t, cfg1, cfg2)

	// Modify cfg1 and verify cfg2 is not affected
	cfg1.System.Hostname = "modified-hostname"
	assert.NotEqual(t, cfg1.System.Hostname, cfg2.System.Hostname)
	assert.Equal(t, testDefaultHostname, cfg2.System.Hostname)
}

func TestDefaultConfig_SystemDefaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, testDefaultHostname, cfg.System.Hostname)
	assert.Equal(t, "local", cfg.System.DomainSuffix)
	assert.Equal(t, testTimezoneKyiv, cfg.System.Timezone)
	assert.Equal(t, "admin@qoxi.cloud", cfg.System.Email)
}

func TestDefaultConfig_NetworkDefaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, BridgeModeInternal, cfg.Network.BridgeMode)
	assert.Equal(t, testSubnetClassA, cfg.Network.PrivateSubnet)
	assert.Empty(t, cfg.Network.InterfaceName) // Should be auto-detected
}

func TestDefaultConfig_StorageDefaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, ZFSRaid1, cfg.Storage.ZFSRaid)
	assert.NotNil(t, cfg.Storage.Disks)
	assert.Empty(t, cfg.Storage.Disks) // Should be auto-detected
}

func TestDefaultConfig_TailscaleDefaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.False(t, cfg.Tailscale.Enabled)
	assert.True(t, cfg.Tailscale.SSH)
	assert.False(t, cfg.Tailscale.WebUI)
}

func TestDefaultConfig_VerboseDefault(t *testing.T) {
	cfg := DefaultConfig()

	assert.False(t, cfg.Verbose)
}

func TestDefaultConfig_SensitiveFieldsEmpty(t *testing.T) {
	cfg := DefaultConfig()

	// All sensitive fields should be empty strings
	assert.Empty(t, cfg.System.RootPassword)
	assert.Empty(t, cfg.System.SSHPublicKey)
	assert.Empty(t, cfg.Tailscale.AuthKey)
}

// FQDN method tests

func TestConfig_FQDN_WithHostnameAndDomainSuffix(t *testing.T) {
	tests := []struct {
		name         string
		hostname     string
		domainSuffix string
		expectedFQDN string
	}{
		{
			name:         "standard local domain",
			hostname:     testHostnamePveServer,
			domainSuffix: "local",
			expectedFQDN: testHostnamePveServer + ".local",
		},
		{
			name:         "example.com domain",
			hostname:     "production",
			domainSuffix: "example.com",
			expectedFQDN: "production.example.com",
		},
		{
			name:         "home.arpa domain",
			hostname:     "homelab",
			domainSuffix: "home.arpa",
			expectedFQDN: "homelab.home.arpa",
		},
		{
			name:         "subdomain style",
			hostname:     "pve01",
			domainSuffix: "dc1.prod.example.com",
			expectedFQDN: "pve01.dc1.prod.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				System: SystemConfig{
					Hostname:     tt.hostname,
					DomainSuffix: tt.domainSuffix,
				},
			}
			assert.Equal(t, tt.expectedFQDN, cfg.FQDN())
		})
	}
}

func TestConfig_FQDN_WithEmptyDomainSuffix(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			Hostname:     "standalone-server",
			DomainSuffix: "",
		},
	}
	assert.Equal(t, "standalone-server", cfg.FQDN())
}

func TestConfig_FQDN_WithEmptyHostname(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			Hostname:     "",
			DomainSuffix: "local",
		},
	}
	// Returns ".local" when hostname is empty
	assert.Equal(t, ".local", cfg.FQDN())
}

func TestConfig_FQDN_WithBothEmpty(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			Hostname:     "",
			DomainSuffix: "",
		},
	}
	// Returns empty string when both are empty
	assert.Equal(t, "", cfg.FQDN())
}

func TestConfig_FQDN_WithDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	// Default config has hostname "pve-qoxi-cloud" and domain_suffix "local"
	assert.Equal(t, testDefaultHostname+".local", cfg.FQDN())
}

func TestDefaultConfig_AllDefaultsMatchPRDSpecification(t *testing.T) {
	cfg := DefaultConfig()

	// This is a comprehensive test that verifies all defaults match PRD
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Hostname", cfg.System.Hostname, testDefaultHostname},
		{"DomainSuffix", cfg.System.DomainSuffix, "local"},
		{"Timezone", cfg.System.Timezone, "Europe/Kyiv"},
		{"Email", cfg.System.Email, "admin@qoxi.cloud"},
		{"BridgeMode", cfg.Network.BridgeMode, BridgeModeInternal},
		{"PrivateSubnet", cfg.Network.PrivateSubnet, testSubnetClassA},
		{"ZFSRaid", cfg.Storage.ZFSRaid, ZFSRaid1},
		{"TailscaleEnabled", cfg.Tailscale.Enabled, false},
		{"TailscaleSSH", cfg.Tailscale.SSH, true},
		{"TailscaleWebUI", cfg.Tailscale.WebUI, false},
		{"Verbose", cfg.Verbose, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}
