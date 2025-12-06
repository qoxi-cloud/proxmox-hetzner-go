package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBridgeMode_String(t *testing.T) {
	tests := []struct {
		name     string
		mode     BridgeMode
		expected string
	}{
		{"internal mode", BridgeModeInternal, "internal"},
		{"external mode", BridgeModeExternal, "external"},
		{"both mode", BridgeModeBoth, "both"},
		{"empty mode", BridgeMode(""), ""},
		{"invalid mode", BridgeMode("invalid"), "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

func TestBridgeMode_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		mode     BridgeMode
		expected bool
	}{
		{"internal is valid", BridgeModeInternal, true},
		{"external is valid", BridgeModeExternal, true},
		{"both is valid", BridgeModeBoth, true},
		{"empty is invalid", BridgeMode(""), false},
		{"invalid string is invalid", BridgeMode("invalid"), false},
		{"uppercase internal is invalid", BridgeMode("Internal"), false},
		{"partial match is invalid", BridgeMode("intern"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.IsValid())
		})
	}
}

func TestZFSRaid_String(t *testing.T) {
	tests := []struct {
		name     string
		raid     ZFSRaid
		expected string
	}{
		{"single raid", ZFSRaidSingle, "single"},
		{"raid0", ZFSRaid0, "raid0"},
		{"raid1", ZFSRaid1, "raid1"},
		{"empty raid", ZFSRaid(""), ""},
		{"invalid raid", ZFSRaid("raid5"), "raid5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.raid.String())
		})
	}
}

func TestZFSRaid_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		raid     ZFSRaid
		expected bool
	}{
		{"single is valid", ZFSRaidSingle, true},
		{"raid0 is valid", ZFSRaid0, true},
		{"raid1 is valid", ZFSRaid1, true},
		{"empty is invalid", ZFSRaid(""), false},
		{"raid5 is invalid", ZFSRaid("raid5"), false},
		{"raid10 is invalid", ZFSRaid("raid10"), false},
		{"uppercase RAID0 is invalid", ZFSRaid("RAID0"), false},
		{"raidz is invalid", ZFSRaid("raidz"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.raid.IsValid())
		})
	}
}

func TestBridgeMode_MarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		mode     BridgeMode
		expected string
	}{
		{"internal mode", BridgeModeInternal, "internal\n"},
		{"external mode", BridgeModeExternal, "external\n"},
		{"both mode", BridgeModeBoth, "both\n"},
		{"empty mode", BridgeMode(""), "\"\"\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.mode)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestBridgeMode_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    BridgeMode
		expectError bool
		errorMsg    string
	}{
		{"internal mode", "internal", BridgeModeInternal, false, ""},
		{"external mode", "external", BridgeModeExternal, false, ""},
		{"both mode", "both", BridgeModeBoth, false, ""},
		{"empty string", "", BridgeMode(""), false, ""},
		{"invalid mode", "invalid", BridgeMode(""), true, "invalid bridge mode"},
		{"uppercase", "Internal", BridgeMode(""), true, "invalid bridge mode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mode BridgeMode
			err := yaml.Unmarshal([]byte(tt.input), &mode)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, mode)
			}
		})
	}
}

func TestZFSRaid_MarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		raid     ZFSRaid
		expected string
	}{
		{"single", ZFSRaidSingle, "single\n"},
		{"raid0", ZFSRaid0, "raid0\n"},
		{"raid1", ZFSRaid1, "raid1\n"},
		{"empty", ZFSRaid(""), "\"\"\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.raid)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestZFSRaid_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    ZFSRaid
		expectError bool
		errorMsg    string
	}{
		{"single", "single", ZFSRaidSingle, false, ""},
		{"raid0", "raid0", ZFSRaid0, false, ""},
		{"raid1", "raid1", ZFSRaid1, false, ""},
		{"empty string", "", ZFSRaid(""), false, ""},
		{"invalid raid5", "raid5", ZFSRaid(""), true, "invalid ZFS raid level"},
		{"uppercase RAID0", "RAID0", ZFSRaid(""), true, "invalid ZFS raid level"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raid ZFSRaid
			err := yaml.Unmarshal([]byte(tt.input), &raid)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, raid)
			}
		})
	}
}

func TestEnumTypes_RoundTrip(t *testing.T) {
	type testConfig struct {
		Bridge BridgeMode `yaml:"bridge"`
		Raid   ZFSRaid    `yaml:"raid"`
	}

	original := testConfig{
		Bridge: BridgeModeExternal,
		Raid:   ZFSRaid1,
	}

	data, err := yaml.Marshal(original)
	require.NoError(t, err)

	var decoded testConfig
	err = yaml.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Bridge, decoded.Bridge)
	assert.Equal(t, original.Raid, decoded.Raid)
}
