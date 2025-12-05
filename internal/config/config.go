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
	InterfaceName string `yaml:"interface" env:"PVE_INTERFACE_NAME"`

	// BridgeMode defines VM networking mode (internal, external, both).
	BridgeMode BridgeMode `yaml:"bridge_mode" env:"PVE_BRIDGE_MODE"`

	// PrivateSubnet is the NAT network subnet (e.g., "10.0.0.0/24").
	PrivateSubnet string `yaml:"private_subnet" env:"PVE_PRIVATE_SUBNET"`
}
