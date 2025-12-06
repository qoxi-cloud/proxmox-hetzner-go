# proxmox-hetzner-go

TUI-based installer for Proxmox VE on Hetzner dedicated servers - no KVM console required.

## Features

- **Interactive TUI** - Beautiful terminal interface powered by Bubbletea
- **Zero KVM Required** - Install Proxmox VE without physical console access
- **Auto-Detection** - Automatically detects network interfaces, disks, and SSH keys
- **ZFS Support** - Full ZFS filesystem support with configurable RAID levels
- **Network Flexibility** - NAT, external, or both bridge modes for VMs
- **SSH Hardening** - Modern ciphers and key-only authentication
- **Tailscale Integration** - Optional VPN setup for secure remote access
- **Configuration Files** - Save and reuse installation configurations

## Requirements

- Hetzner dedicated server with Rescue System access
- Linux environment (Rescue System is Debian-based)
- Root access
- Network connectivity
- At least one disk for installation

## Quick Start

Boot your Hetzner server into Rescue System, then run:

```bash
curl -sSL https://raw.githubusercontent.com/qoxi-cloud/proxmox-hetzner-go/main/scripts/install.sh | bash
```

This will download and run the installer with the interactive TUI.

## Installation from Source

### Prerequisites

- Go 1.24 or later
- Git

### Build

```bash
# Clone the repository
git clone https://github.com/qoxi-cloud/proxmox-hetzner-go.git
cd proxmox-hetzner-go

# Build for current platform
make build

# Build for Linux (cross-compile from macOS)
make build-linux
```

The binary will be created in the `build/` directory.

### Cross-Compile for Linux

If you're developing on macOS and need to deploy to a Linux server:

```bash
# Using make
make build-linux

# Or manually
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/pve-install-linux ./cmd/pve-install
```

## Usage

### Basic Usage

```bash
# Run with interactive TUI
./pve-install

# Show help
./pve-install --help

# Show version
./pve-install version
```

### CLI Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | `-c` | Load configuration from YAML file |
| `--save-config` | `-s` | Save configuration to file after input |
| `--verbose` | `-v` | Enable verbose logging |
| `--help` | `-h` | Show help message |
| `--version` | | Show version information |

### Examples

```bash
# Use a configuration file
./pve-install --config /path/to/config.yaml

# Enable verbose logging
./pve-install --verbose

# Save configuration after TUI input
./pve-install --save-config

# Combine flags
./pve-install -c config.yaml -v
```

## Configuration

Configuration can be provided via:

1. TUI input (highest priority)
2. Environment variables (prefix: `PVE_`)
3. YAML configuration file
4. Default values (lowest priority)

See [configs/example.yaml](configs/example.yaml) for a complete configuration reference with all available options.

### Configuration Sections

| Section | Description |
|---------|-------------|
| `hostname` | Server hostname (RFC 1123 compliant) |
| `network` | Network interface, IP, gateway, DNS, bridge mode |
| `storage` | Disk selection, filesystem type, ZFS options |
| `proxmox` | Repository type, admin email |
| `ssh` | Port, public key, hardening options |
| `tailscale` | VPN enable, auth key, SSH/Funnel options |
| `system` | Timezone and other system settings |

### Environment Variables

All configuration options can be set via environment variables. Environment variables override config file values but are overridden by TUI user input.

> **Note:** System configuration variables use the `PVE_` prefix, while network, storage, and Tailscale variables use descriptive names without prefix for clarity and brevity.

#### System Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `PVE_HOSTNAME` | Server hostname (RFC 1123 compliant) | `pve-server` |
| `PVE_DOMAIN_SUFFIX` | Domain suffix for FQDN | `local` |
| `PVE_TIMEZONE` | Server timezone | `Europe/Kyiv` |
| `PVE_EMAIL` | Admin email address | `admin@example.com` |
| `PVE_ROOT_PASSWORD` | Root password (sensitive) | - |
| `PVE_SSH_PUBLIC_KEY` | SSH public key (sensitive) | - |

#### Network Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `INTERFACE_NAME` | Primary network interface | `eth0` |
| `BRIDGE_MODE` | VM networking mode | `internal`, `external`, `both` |
| `PRIVATE_SUBNET` | NAT network subnet | `10.0.0.0/24` |

#### Storage Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `ZFS_RAID` | ZFS RAID level | `single`, `raid0`, `raid1` |
| `DISKS` | Disk devices (comma-separated) | `/dev/sda,/dev/sdb` |

#### Tailscale Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `INSTALL_TAILSCALE` | Enable Tailscale installation | `true`, `false`, `yes`, `no`, `1`, `0` |
| `TAILSCALE_AUTH_KEY` | Tailscale auth key (sensitive) | - |
| `TAILSCALE_SSH` | Enable SSH over Tailscale | `true`, `false`, `yes`, `no`, `1`, `0` |
| `TAILSCALE_WEBUI` | Expose WebUI via Tailscale | `true`, `false`, `yes`, `no`, `1`, `0` |

#### Example Usage

```bash
# Basic setup
export PVE_HOSTNAME=pve-server
export PVE_TIMEZONE=Europe/Berlin
export PVE_EMAIL=admin@example.com
./pve-install

# With storage configuration
export ZFS_RAID=raid1
export DISKS="/dev/sda,/dev/sdb"
./pve-install

# With Tailscale
export INSTALL_TAILSCALE=true
export TAILSCALE_AUTH_KEY=tskey-auth-xxx
./pve-install
```

> **Note:** Sensitive fields (`PVE_ROOT_PASSWORD`, `PVE_SSH_PUBLIC_KEY`, `TAILSCALE_AUTH_KEY`) are loaded from environment variables but are never persisted to configuration files.

## Development

### Developer Tools

- Go 1.24+
- Node.js 24+ (for MCP servers)
- asdf (recommended for version management)

### Setup

```bash
# Install tool versions
asdf install

# Install MCP server dependencies
npm install
```

### Commands

```bash
# Build
make build           # Build for current platform
make build-linux     # Cross-compile for Linux

# Test
make test            # Run tests
make test-coverage   # Run tests with coverage report

# Code quality
make lint            # Run linter
make fmt             # Format code

# Cleanup
make clean           # Remove build artifacts

# Help
make help            # Show all available targets
```

### Project Structure

```text
proxmox-hetzner-go/
├── cmd/pve-install/    # CLI entry point
├── internal/           # Private packages
│   ├── config/         # Configuration management
│   ├── exec/           # Command execution
│   ├── tui/            # Terminal user interface
│   └── installer/      # Installation logic
├── pkg/version/        # Version information
├── configs/            # Example configurations
├── scripts/            # Installation scripts
└── build/              # Build output (git-ignored)
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
