package config

// SystemConfig holds system-level configuration settings for the server.
// It can be loaded from YAML files or environment variables.
type SystemConfig struct {
	// Hostname is the server hostname (RFC 1123 compliant).
	Hostname string `yaml:"hostname" env:"PVE_HOSTNAME"`

	// DomainSuffix is the domain suffix (e.g., "local" or "example.com").
	DomainSuffix string `yaml:"domain_suffix" env:"PVE_DOMAIN_SUFFIX"`

	// Timezone is the server timezone (e.g., "Europe/Kyiv").
	Timezone string `yaml:"timezone" env:"PVE_TIMEZONE"`

	// Email is the admin email for notifications.
	Email string `yaml:"email" env:"PVE_EMAIL"`

	// RootPassword is the root password (excluded from file serialization).
	RootPassword string `yaml:"-" env:"PVE_ROOT_PASSWORD"`

	// SSHPublicKey is the SSH public key for authentication (excluded from file serialization).
	SSHPublicKey string `yaml:"-" env:"PVE_SSH_PUBLIC_KEY"`
}

// NetworkConfig holds network configuration options.
type NetworkConfig struct {
	// InterfaceName is the primary network interface (e.g., "eth0").
	InterfaceName string `yaml:"interface" env:"INTERFACE_NAME"`

	// BridgeMode defines VM networking mode (internal, external, both).
	BridgeMode BridgeMode `yaml:"bridge_mode" env:"BRIDGE_MODE"`

	// PrivateSubnet is the NAT network subnet (e.g., "10.0.0.0/24").
	PrivateSubnet string `yaml:"private_subnet" env:"PRIVATE_SUBNET"`
}

// StorageConfig holds storage and disk configuration.
type StorageConfig struct {
	// ZFSRaid is the ZFS RAID level (single, raid0, raid1).
	ZFSRaid ZFSRaid `yaml:"zfs_raid" env:"ZFS_RAID"`

	// Disks is the list of disk devices to use (e.g., "/dev/sda", "/dev/sdb").
	Disks []string `yaml:"disks" env:"DISKS" envSeparator:","`
}

// TailscaleConfig holds Tailscale VPN configuration settings.
type TailscaleConfig struct {
	// Enabled controls whether Tailscale should be installed.
	Enabled bool `yaml:"enabled" env:"INSTALL_TAILSCALE"`

	// AuthKey is the Tailscale authentication key (excluded from file serialization).
	AuthKey string `yaml:"-" env:"TAILSCALE_AUTH_KEY"`

	// SSH enables SSH advertisement on the Tailscale network.
	SSH bool `yaml:"ssh" env:"TAILSCALE_SSH"`

	// WebUI exposes Proxmox interface via Tailscale.
	WebUI bool `yaml:"webui" env:"TAILSCALE_WEBUI"`
}
