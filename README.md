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

All configuration options can be set via environment variables with the `PVE_` prefix:

```bash
export PVE_HOSTNAME=pve-server
export PVE_NETWORK_ADDRESS=192.168.1.100/24
export PVE_STORAGE_FILESYSTEM=zfs
./pve-install
```

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
