package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
