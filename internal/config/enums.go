package config

// BridgeMode defines the network bridge mode for VM networking.
type BridgeMode string

const (
	// BridgeModeInternal creates NAT network for VMs (private IPs).
	BridgeModeInternal BridgeMode = "internal"
	// BridgeModeExternal allows VMs to get public IPs.
	BridgeModeExternal BridgeMode = "external"
	// BridgeModeBoth creates both NAT and external bridges.
	BridgeModeBoth BridgeMode = "both"
)

// String returns the string representation of BridgeMode.
func (b BridgeMode) String() string {
	return string(b)
}

// IsValid checks if the BridgeMode is a valid value.
func (b BridgeMode) IsValid() bool {
	switch b {
	case BridgeModeInternal, BridgeModeExternal, BridgeModeBoth:
		return true
	}

	return false
}

// ZFSRaid defines the ZFS RAID level.
type ZFSRaid string

const (
	// ZFSRaidSingle is a single disk configuration (no redundancy).
	ZFSRaidSingle ZFSRaid = "single"
	// ZFSRaid0 is striped (no redundancy, max performance).
	ZFSRaid0 ZFSRaid = "raid0"
	// ZFSRaid1 is mirrored (requires 2+ disks).
	ZFSRaid1 ZFSRaid = "raid1"
)

// String returns the string representation of ZFSRaid.
func (z ZFSRaid) String() string {
	return string(z)
}

// IsValid checks if the ZFSRaid is a valid value.
func (z ZFSRaid) IsValid() bool {
	switch z {
	case ZFSRaidSingle, ZFSRaid0, ZFSRaid1:
		return true
	}

	return false
}
