# Repository Planning Graph (RPG) PRD: Proxmox VE Installer for Hetzner

## Overview

### Problem Statement

Installing Proxmox VE on Hetzner dedicated servers requires:
- KVM console access (not always available on budget servers)
- Manual network bridge configuration (error-prone, 15-20 minutes)
- Complex ZFS RAID setup (requires deep ZFS knowledge)
- SSH hardening (often skipped or misconfigured)
- Post-installation optimization (1-2 hours of manual work)

**Current bash solution** at [qoxi-cloud/proxmox-hetzner](https://github.com/qoxi-cloud/proxmox-hetzner) works but has limitations:
- Bash is hard to test and maintain
- Limited error handling capabilities
- Complex state management with global variables
- No type safety or compile-time checks
- Difficult to extend with new features

### Target Users

| Persona | Experience | Goals | Pain Points |
|---------|------------|-------|-------------|
| **DevOps Engineer** | 10+ years | Quick, reliable deployments | Manual work is tedious |
| **Homelab Enthusiast** | Intermediate | Personal virtualization | Complex networking/ZFS |
| **MSP Technician** | Junior-Mid | Standardized deployments | Inconsistent configs |

### Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Installation time | < 30 min | End-to-end timer |
| Success rate | > 95% | Successful completions |
| Test coverage | > 80% | Go coverage tool |
| Binary size | < 20 MB | Release artifact |
| User satisfaction | TUI usable by intermediate Linux users | User feedback |

---

## Functional Decomposition

### Capability: Configuration Management

Handles all configuration loading, saving, validation, and defaults.

#### Feature: Config Struct Definition
- **Description**: Define strongly-typed configuration with YAML and env tags
- **Inputs**: None (struct definition)
- **Outputs**: Config struct with all installation parameters
- **Behavior**: Provides defaults, supports nested structs for System/Network/Storage/Tailscale

#### Feature: Config Validation
- **Description**: Validate all configuration values before installation
- **Inputs**: Config struct
- **Outputs**: Validation result with all errors (not just first)
- **Behavior**: Validates hostname (RFC 1123), email, password (min 8 chars), SSH key format, subnet CIDR, enum values

#### Feature: YAML File Operations
- **Description**: Load/save configuration from/to YAML files
- **Inputs**: File path, Config struct
- **Outputs**: Loaded config or error, success/failure for save
- **Behavior**: Merge with defaults, exclude sensitive fields from save, create parent directories

#### Feature: Environment Variable Loading
- **Description**: Override config values from environment variables
- **Inputs**: Environment, Config struct
- **Outputs**: Modified config
- **Behavior**: Parse booleans (yes/no/true/false/1/0), parse comma-separated lists for disks

---

### Capability: Command Execution

Abstraction layer for system command execution, enabling testing.

#### Feature: Executor Interface
- **Description**: Define interface for running system commands
- **Inputs**: Command name, arguments, optional stdin
- **Outputs**: Exit code, stdout, stderr, error
- **Behavior**: Support context cancellation, timeout, output capture

#### Feature: Real Executor
- **Description**: Production executor using os/exec
- **Inputs**: Same as interface
- **Outputs**: Real command results
- **Behavior**: Execute actual system commands

#### Feature: Mock Executor
- **Description**: Test executor recording commands
- **Inputs**: Same as interface, plus configured responses
- **Outputs**: Configured mock responses
- **Behavior**: Record all executed commands, return preset outputs/errors

---

### Capability: Terminal User Interface

Interactive Bubbletea-based TUI for configuration and progress display.

#### Feature: Lipgloss Styles
- **Description**: Define visual styles for all UI elements
- **Inputs**: None (style definitions)
- **Outputs**: Style objects for titles, inputs, menus, status indicators
- **Behavior**: Consistent color palette (purple/green/amber/red), responsive to terminal size

#### Feature: TUI Model
- **Description**: Main Bubbletea model managing application state
- **Inputs**: Config pointer, window dimensions
- **Outputs**: Rendered views, state updates
- **Behavior**: Track current screen, cursor position, input values, installation progress

#### Feature: Welcome Screen
- **Description**: First screen with logo and feature overview
- **Inputs**: Window dimensions
- **Outputs**: Rendered ASCII art logo, feature list, continue prompt
- **Behavior**: Center content, wait for Enter to proceed

#### Feature: Text Input Screens
- **Description**: Screens for hostname, domain, email, password, SSH key, subnet
- **Inputs**: Current value, validation function
- **Outputs**: Rendered input field with validation feedback
- **Behavior**: Show placeholder, inline validation errors, env var indicator

#### Feature: Menu Selection Screens
- **Description**: Screens for bridge mode, ZFS RAID, Tailscale enable
- **Inputs**: Options list, current selection
- **Outputs**: Rendered menu with highlighted selection
- **Behavior**: Arrow key navigation, Enter to select, cursor wrapping

#### Feature: Summary Screen
- **Description**: Display all configuration for review before install
- **Inputs**: Complete config
- **Outputs**: Grouped configuration display with masked password
- **Behavior**: Show all values, allow going back to edit

#### Feature: Confirmation Screen
- **Description**: Final warning and explicit confirmation
- **Inputs**: Disk list to erase
- **Outputs**: Warning message, confirmation input
- **Behavior**: Require typing "yes" to proceed, display destructive warning

#### Feature: Progress Screen
- **Description**: Real-time installation progress display
- **Inputs**: Step list, current step, elapsed time
- **Outputs**: Step list with status icons, spinner, timer
- **Behavior**: Animate spinner, update step status, disable quit

#### Feature: Completion Screen
- **Description**: Success message with access information
- **Inputs**: Config (for URLs), Tailscale status
- **Outputs**: Web UI URL, SSH command, credentials reminder
- **Behavior**: Clear access instructions, exit on keypress

#### Feature: Error Screen
- **Description**: Failure message with troubleshooting info
- **Inputs**: Failed step, error message
- **Outputs**: Error details, log location, troubleshooting hints
- **Behavior**: Actionable error messages, exit on keypress

#### Feature: Navigation System
- **Description**: Screen flow logic with conditional paths
- **Inputs**: Current screen, config values
- **Outputs**: Next/previous screen
- **Behavior**: Skip subnet if external-only, skip Tailscale auth if disabled

---

### Capability: Installation Engine

Orchestrates all installation steps with progress callbacks.

#### Feature: Installer Framework
- **Description**: Step orchestration with callbacks
- **Inputs**: Config, Executor, Logger
- **Outputs**: Success/failure, step-by-step progress
- **Behavior**: Execute steps in order, stop on error, support cancellation

#### Feature: Pre-flight Checks
- **Description**: Validate system requirements before installation
- **Inputs**: Executor
- **Outputs**: Pass/fail with all failed checks
- **Behavior**: Check root, required tools, network, KVM

#### Feature: Hardware Detection
- **Description**: Auto-detect network interfaces, disks, SSH keys
- **Inputs**: Executor
- **Outputs**: Detected values
- **Behavior**: Find first UP interface, list disk devices, read authorized_keys

#### Feature: ISO Download
- **Description**: Download Proxmox ISO with progress
- **Inputs**: URL, destination path, progress callback
- **Outputs**: Downloaded file or error
- **Behavior**: Show progress, skip if exists, verify size

#### Feature: Answer File Generation
- **Description**: Generate Proxmox auto-install answer file (TOML)
- **Inputs**: Config
- **Outputs**: TOML content written to file
- **Behavior**: Include all required fields, escape special characters, handle RAID variants

#### Feature: Proxmox Installation
- **Description**: Execute Proxmox installation via QEMU
- **Inputs**: ISO path, disks, answer file
- **Outputs**: Success/failure
- **Behavior**: Build QEMU command, monitor progress, handle timeout

#### Feature: Network Configuration
- **Description**: Generate and apply network bridge configuration
- **Inputs**: Config (bridge mode, interface, subnet)
- **Outputs**: /etc/network/interfaces content
- **Behavior**: Generate NAT/external/both configs, iptables rules, IP forwarding

#### Feature: SSH Hardening
- **Description**: Apply security configuration to SSH
- **Inputs**: SSH public key
- **Outputs**: Hardened sshd_config, authorized_keys
- **Behavior**: Modern ciphers only, disable password auth, install key

#### Feature: System Optimization
- **Description**: Install packages and apply optimizations
- **Inputs**: Config (timezone, hostname)
- **Outputs**: Configured system
- **Behavior**: Install utilities, configure ZFS ARC, remove subscription notice

#### Feature: Tailscale Installation
- **Description**: Install and configure Tailscale VPN
- **Inputs**: Auth key, SSH/WebUI flags
- **Outputs**: Connected Tailscale node
- **Behavior**: Run install script, authenticate, configure features

#### Feature: Finalization
- **Description**: Restart services, cleanup, verify
- **Inputs**: None
- **Outputs**: Ready system
- **Behavior**: Restart SSH/networking, remove temp files, verify web UI

---

### Capability: CLI Interface

Command-line interface using Cobra.

#### Feature: Root Command
- **Description**: Main command with flags and execution
- **Inputs**: CLI arguments
- **Outputs**: Exit code
- **Behavior**: Parse flags, load config, start TUI or non-interactive

#### Feature: Version Command
- **Description**: Display version information
- **Inputs**: None
- **Outputs**: Version, commit, build date
- **Behavior**: Print formatted version info

#### Feature: Flags
- **Description**: CLI flag definitions
- **Inputs**: None
- **Outputs**: Flag values
- **Behavior**: --config, --save-config, --verbose, --version, --help

---

### Capability: Logging

Structured logging to file with optional stdout.

#### Feature: Logger
- **Description**: File-based logger with timestamps
- **Inputs**: Verbose flag, log path
- **Outputs**: Log entries to file
- **Behavior**: ISO 8601 timestamps, fallback path, flush on close

---

## Structural Decomposition

### Repository Structure

```
proxmox-hetzner-go/
├── cmd/
│   └── pve-install/
│       └── main.go                 # Maps to: CLI Interface
├── internal/
│   ├── config/                     # Maps to: Configuration Management
│   │   ├── config.go              # Config Struct Definition
│   │   ├── validation.go          # Config Validation
│   │   ├── file.go                # YAML File Operations
│   │   └── env.go                 # Environment Variable Loading
│   ├── exec/                       # Maps to: Command Execution
│   │   ├── command.go             # Executor Interface + Real Executor
│   │   └── mock.go                # Mock Executor
│   ├── tui/                        # Maps to: Terminal User Interface
│   │   ├── model.go               # TUI Model
│   │   ├── styles.go              # Lipgloss Styles
│   │   ├── screens.go             # All screen views
│   │   ├── navigation.go          # Navigation System
│   │   ├── messages.go            # Tea message types
│   │   └── keys.go                # Key bindings
│   ├── installer/                  # Maps to: Installation Engine
│   │   ├── installer.go           # Installer Framework
│   │   ├── preflight.go           # Pre-flight Checks
│   │   ├── detection.go           # Hardware Detection
│   │   ├── download.go            # ISO Download
│   │   ├── answerfile.go          # Answer File Generation
│   │   ├── proxmox.go             # Proxmox Installation
│   │   ├── network.go             # Network Configuration
│   │   ├── ssh.go                 # SSH Hardening
│   │   ├── system.go              # System Optimization
│   │   ├── tailscale.go           # Tailscale Installation
│   │   ├── finalize.go            # Finalization
│   │   └── logging.go             # Logger
│   └── testutil/                   # Test utilities
│       └── testutil.go
├── test/
│   └── e2e_test.go
├── configs/
│   └── example.yaml
├── scripts/
│   └── install.sh
├── .github/
│   └── workflows/
│       ├── build.yml
│       └── release.yml
├── .tool-versions                  # asdf: nodejs 22.11.0, golang 1.23.3
├── package.json                    # MCP server dependencies
├── .mcp.json                       # MCP configuration
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── CLAUDE.md
└── .goreleaser.yaml
```

### Module Definitions

#### Module: config
- **Maps to capability**: Configuration Management
- **Responsibility**: All configuration handling
- **Exports**:
  - `Config` struct
  - `DefaultConfig()` - returns config with defaults
  - `LoadFromFile(path)` - load YAML config
  - `SaveToFile(path)` - save YAML config
  - `LoadFromEnv(cfg)` - apply env var overrides
  - `Validate()` - validate config

#### Module: exec
- **Maps to capability**: Command Execution
- **Responsibility**: System command abstraction
- **Exports**:
  - `Executor` interface
  - `RealExecutor` struct
  - `MockExecutor` struct
  - `NewMockExecutor()`

#### Module: tui
- **Maps to capability**: Terminal User Interface
- **Responsibility**: All TUI rendering and state
- **Exports**:
  - `Model` struct
  - `New(cfg)` - create new TUI model
  - `Screen` enum

#### Module: installer
- **Maps to capability**: Installation Engine
- **Responsibility**: All installation steps
- **Exports**:
  - `Installer` struct
  - `New(cfg, exec, log)` - create installer
  - `Run(ctx)` - execute installation
  - `Step` interface

---

## Dependency Graph

### Foundation Layer (Phase 0)
No dependencies - built first.

- **config/config.go**: Defines Config struct and defaults
  - Provides: Config, SystemConfig, NetworkConfig, StorageConfig, TailscaleConfig, BridgeMode, ZFSRaid, DefaultConfig()
  
- **exec/command.go**: Defines Executor interface
  - Provides: Executor interface, RealExecutor

- **exec/mock.go**: Mock executor for testing
  - Provides: MockExecutor, NewMockExecutor()

### Validation Layer (Phase 1)
- **config/validation.go**: Depends on [config/config.go]
  - Provides: Validate(), validateHostname(), validateEmail(), validateSSHKey(), validateSubnet()

- **config/file.go**: Depends on [config/config.go]
  - Provides: LoadFromFile(), SaveToFile()

- **config/env.go**: Depends on [config/config.go]
  - Provides: LoadFromEnv(), EnvVarSet()

### Logging Layer (Phase 1)
- **installer/logging.go**: Depends on [nothing]
  - Provides: Logger, NewLogger()

### TUI Foundation Layer (Phase 2)
- **tui/styles.go**: Depends on [nothing]
  - Provides: All Lipgloss styles, color palette, Logo

- **tui/messages.go**: Depends on [nothing]
  - Provides: StepStartMsg, StepCompleteMsg, StepErrorMsg, InstallCompleteMsg

- **tui/keys.go**: Depends on [nothing]
  - Provides: KeyBindings

### TUI Model Layer (Phase 3)
- **tui/model.go**: Depends on [config/config.go, tui/styles.go, tui/messages.go]
  - Provides: Model, Screen enum, New(), Init(), Update(), View()

- **tui/navigation.go**: Depends on [tui/model.go, config/config.go]
  - Provides: nextScreen(), prevScreen(), saveCurrentInput()

- **tui/screens.go**: Depends on [tui/model.go, tui/styles.go, config/config.go]
  - Provides: viewWelcome(), viewHostname(), viewSummary(), viewInstalling(), etc.

### Installer Step Layer (Phase 4)
All depend on [exec/command.go, installer/logging.go]

- **installer/preflight.go**: + Depends on [config/config.go]
  - Provides: PreflightStep

- **installer/detection.go**: + Depends on [config/config.go]
  - Provides: DetectionStep

- **installer/download.go**: Depends on [exec, logging]
  - Provides: DownloadStep

- **installer/answerfile.go**: Depends on [config/config.go]
  - Provides: AnswerFileStep

- **installer/proxmox.go**: Depends on [exec, config, download, answerfile]
  - Provides: InstallStep

- **installer/network.go**: Depends on [config/config.go]
  - Provides: NetworkStep

- **installer/ssh.go**: Depends on [config/config.go]
  - Provides: SSHStep

- **installer/system.go**: Depends on [exec, config]
  - Provides: SystemStep

- **installer/tailscale.go**: Depends on [exec, config]
  - Provides: TailscaleStep

- **installer/finalize.go**: Depends on [exec, logging]
  - Provides: FinalizeStep

### Installer Orchestration Layer (Phase 5)
- **installer/installer.go**: Depends on [all installer steps, config, exec, logging]
  - Provides: Installer, New(), Run()

### Integration Layer (Phase 6)
- **tui/model.go (integration)**: Depends on [installer/installer.go, tui components]
  - Provides: startInstallation() - connects TUI to installer

### CLI Layer (Phase 7)
- **cmd/pve-install/main.go**: Depends on [config, tui, installer]
  - Provides: main(), CLI entry point

---

## Implementation Roadmap

### Phase 0: Foundation
**Goal**: Establish core types and abstractions

**Entry Criteria**: Clean repository with go.mod

**Tasks**:
- [ ] Task 1.1: Project setup (go.mod, Makefile, .tool-versions, package.json) - depends on: none
- [ ] Task 1.2: Config struct definition (config/config.go) - depends on: 1.1
- [ ] Task 1.6: Executor interface (exec/command.go, exec/mock.go) - depends on: 1.1

**Exit Criteria**: Can import config and exec packages without errors

**Delivers**: Foundation for all other modules

---

### Phase 1: Configuration & Logging
**Goal**: Complete configuration system

**Entry Criteria**: Phase 0 complete

**Tasks**:
- [ ] Task 1.3: Config validation (config/validation.go) - depends on: 1.2
- [ ] Task 1.4: YAML file operations (config/file.go) - depends on: 1.2
- [ ] Task 1.5: Environment variables (config/env.go) - depends on: 1.2
- [ ] Task 1.7: Logging system (installer/logging.go) - depends on: 1.1
- [ ] Task 1.8: Cobra CLI setup (cmd/pve-install/main.go) - depends on: 1.2, 1.4, 1.5

**Exit Criteria**: Can load config from file, env, CLI flags; validation works

**Delivers**: Full configuration system, basic CLI

---

### Phase 2: TUI Foundation
**Goal**: Basic TUI rendering capability

**Entry Criteria**: Phase 1 complete

**Tasks**:
- [ ] Task 2.1: Lipgloss styles (tui/styles.go) - depends on: 1.1
- [ ] Task 2.2: TUI model structure (tui/model.go) - depends on: 2.1, 1.2
- [ ] Task 2.3: TUI Init and Update (tui/model.go) - depends on: 2.2
- [ ] Task 2.4: Welcome screen (tui/screens.go) - depends on: 2.3

**Exit Criteria**: Can run TUI, see welcome screen, press Enter to continue

**Delivers**: Visible, interactive TUI (first user-facing output)

---

### Phase 3: TUI Screens
**Goal**: All configuration screens

**Entry Criteria**: Phase 2 complete

**Tasks**:
- [ ] Task 2.5: Text input screens (hostname, domain, etc.) - depends on: 2.4
- [ ] Task 2.6: Menu selection screens (bridge, ZFS, Tailscale) - depends on: 2.4
- [ ] Task 2.7: Summary and confirmation screens - depends on: 2.5, 2.6
- [ ] Task 2.8: Navigation system (nextScreen, prevScreen) - depends on: 2.5, 2.6, 2.7

**Exit Criteria**: Can navigate through all config screens, values saved to config

**Delivers**: Complete configuration flow

---

### Phase 4: Installation Engine
**Goal**: All installation steps (with mock)

**Entry Criteria**: Phase 1, 3 complete

**Tasks**:
- [ ] Task 3.1: Installer framework (installer/installer.go) - depends on: 1.6, 1.7
- [ ] Task 3.2: Pre-flight checks - depends on: 3.1
- [ ] Task 3.3: Hardware detection - depends on: 3.1
- [ ] Task 3.4: ISO download - depends on: 3.1
- [ ] Task 3.5: Answer file generation - depends on: 3.1
- [ ] Task 3.6: Proxmox installation - depends on: 3.4, 3.5
- [ ] Task 3.7: Network configuration - depends on: 3.6
- [ ] Task 3.8: SSH hardening - depends on: 3.6
- [ ] Task 3.9: System optimization - depends on: 3.6
- [ ] Task 3.10: Tailscale installation - depends on: 3.6
- [ ] Task 3.11: Finalization - depends on: 3.7, 3.8, 3.9

**Exit Criteria**: All steps execute with MockExecutor, callbacks fire

**Delivers**: Complete installation engine (testable)

---

### Phase 5: Progress UI & Integration
**Goal**: Connect TUI to installer

**Entry Criteria**: Phase 3, 4 complete

**Tasks**:
- [ ] Task 4.1: Installation progress screen - depends on: 2.2, 3.1
- [ ] Task 4.2: Progress message types - depends on: 4.1
- [ ] Task 4.3: TUI-Installer integration - depends on: 4.1, 4.2, 3.1
- [ ] Task 4.4: Completion screen - depends on: 4.1
- [ ] Task 4.5: Error screen - depends on: 4.1

**Exit Criteria**: Full flow: welcome → config → confirm → install → complete

**Delivers**: End-to-end working application

---

### Phase 6: Testing
**Goal**: Comprehensive test coverage

**Entry Criteria**: Phase 5 complete

**Tasks**:
- [ ] Task 5.1: Test utilities package - depends on: 1.6
- [ ] Tasks 5.2-5.5: Config unit tests - depends on: 1.2-1.5
- [ ] Task 5.6: Executor unit tests - depends on: 1.6
- [ ] Task 5.7: Logging unit tests - depends on: 1.7
- [ ] Tasks 5.8-5.12: Installer unit tests - depends on: 3.x
- [ ] Tasks 5.13-5.14: Installer integration tests - depends on: 3.11
- [ ] Tasks 5.15-5.20: TUI tests - depends on: 2.x, 4.x
- [ ] Task 5.21: E2E tests - depends on: 5.13, 5.20
- [ ] Task 5.22: CLI tests - depends on: 1.8
- [ ] Task 5.23: Coverage configuration - depends on: 5.21
- [ ] Task 5.24: Benchmark tests - depends on: 5.21
- [ ] Task 5.25: Test documentation - depends on: 5.23

**Exit Criteria**: 80%+ coverage, all tests pass

**Delivers**: Reliable, tested codebase

---

### Phase 7: Documentation & Release
**Goal**: Production-ready release

**Entry Criteria**: Phase 6 complete

**Tasks**:
- [ ] Task 6.1: README documentation - depends on: none
- [ ] Task 6.2: GitHub Actions CI - depends on: 5.1
- [ ] Task 6.3: Release automation (goreleaser) - depends on: 6.2
- [ ] Task 6.4: One-line installer script - depends on: 6.3

**Exit Criteria**: Published release, working install script

**Delivers**: Production deployment

---

## Test Strategy

### Test Pyramid

```
          /\
         /E2E\           ← 10% (Full flow with MockExecutor)
        /------\
       /Integration\     ← 30% (Installer + TUI flows)
      /------------\
     /  Unit Tests  \    ← 60% (Config, validation, generation)
    /----------------\
```

### Coverage Requirements

| Package | Minimum | Target |
|---------|---------|--------|
| config | 90% | 95% |
| exec | 80% | 90% |
| installer | 80% | 85% |
| tui | 70% | 80% |
| **Overall** | **80%** | **85%** |

### Critical Test Scenarios

#### Configuration
**Happy path**: Valid config loads, validates, saves
**Edge cases**: Empty values, boundary lengths, special characters
**Error cases**: Invalid hostname, bad SSH key, malformed YAML

#### Installer Steps
**Happy path**: All steps complete with mock
**Edge cases**: Missing disks, no network interface
**Error cases**: Network failure, disk full, command timeout
**Integration**: Step order correct, callbacks fire

#### TUI
**Happy path**: Navigate all screens, values save
**Edge cases**: Resize during install, rapid navigation
**Error cases**: Validation errors display, error screen
**Snapshot**: Golden file tests for screen rendering

### Test Generation Guidelines

1. Use table-driven tests for validation functions
2. Use MockExecutor for all system command tests
3. Use teatest for TUI interaction tests
4. Use testutil helpers for common setup
5. Name tests: Test{Function}_{Scenario}
6. Keep tests deterministic - no time.Sleep without mocking

---

## Technical Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     pve-install Binary                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────────┐  ┌───────────┐  │
│  │   CLI   │→ │ Config  │→ │     TUI     │→ │ Installer │  │
│  │ (Cobra) │  │ (Viper) │  │ (Bubbletea) │  │  (Steps)  │  │
│  └─────────┘  └─────────┘  └─────────────┘  └───────────┘  │
│                                                   │          │
│                                                   ▼          │
│                                            ┌───────────┐     │
│                                            │ Executor  │     │
│                                            │ (os/exec) │     │
│                                            └───────────┘     │
└─────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    Hetzner Rescue System                     │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────────┐ │
│  │  QEMU   │  │ Network │  │  Disks  │  │ Proxmox ISO     │ │
│  └─────────┘  └─────────┘  └─────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

| Component | Technology | Version | Rationale |
|-----------|------------|---------|-----------|
| Language | Go | 1.23+ | Single binary, cross-compile, strong typing |
| TUI Framework | Bubbletea | v1.2+ | Elm architecture, composable, testable |
| TUI Components | Bubbles | v0.20+ | Ready-made inputs, spinners, lists |
| TUI Styling | Lipgloss | v1.0+ | Declarative terminal styling |
| CLI | Cobra | v1.8+ | Industry standard, completion support |
| Config | Viper | v1.19+ | Multi-source config, env var support |
| Testing | testify | v1.9+ | Assertions, mocks |
| TUI Testing | teatest | latest | Bubbletea-specific testing |
| Version Mgmt | asdf | latest | Multi-runtime version management |
| Release | goreleaser | latest | Automated multi-platform releases |

### Data Flow

```
1. CLI Parse
   └─→ Load --config file (if provided)
       └─→ Apply environment variables
           └─→ Initialize TUI with config

2. TUI Configuration
   └─→ User inputs on each screen
       └─→ Save to config on navigation
           └─→ Validate on summary screen

3. Installation
   └─→ Create Installer with config
       └─→ Execute steps sequentially
           └─→ Send progress messages to TUI
               └─→ Update progress screen

4. Completion
   └─→ Display success/error screen
       └─→ Exit with appropriate code
```

---

## Risks and Mitigations

### Technical Risks

**Risk**: QEMU installation fails on certain hardware
- **Impact**: High - core functionality broken
- **Likelihood**: Medium
- **Mitigation**: Extensive hardware testing, configurable QEMU flags
- **Fallback**: Document manual recovery procedure

**Risk**: Network configuration locks out user
- **Impact**: High - server inaccessible
- **Likelihood**: Low
- **Mitigation**: Backup existing config, test on VM first
- **Fallback**: Hetzner rescue mode recovery guide

**Risk**: TUI not rendering correctly on all terminals
- **Impact**: Medium - poor UX
- **Likelihood**: Medium
- **Mitigation**: Test on common terminals, fallback styling
- **Fallback**: Non-interactive mode with --config

### Dependency Risks

**Risk**: Proxmox ISO URL changes
- **Impact**: Medium - download fails
- **Likelihood**: Low
- **Mitigation**: Configurable URL, version detection
- **Fallback**: Manual ISO specification flag

**Risk**: Tailscale install script changes
- **Impact**: Low - optional feature
- **Likelihood**: Low
- **Mitigation**: Pin to known version, handle failures gracefully

### Scope Risks

**Risk**: Feature creep from bash version
- **Impact**: Medium - delayed release
- **Likelihood**: High
- **Mitigation**: Strict MVP scope, future versions for extras
- **Fallback**: Document unsupported features

**Risk**: Test coverage delays release
- **Impact**: Medium - quality vs speed
- **Likelihood**: Medium
- **Mitigation**: Prioritize critical path tests
- **Fallback**: Ship with 70% coverage, improve post-release

---

## Appendix

### References

- Original bash project: https://github.com/qoxi-cloud/proxmox-hetzner
- Bubbletea documentation: https://github.com/charmbracelet/bubbletea
- Proxmox VE installation: https://pve.proxmox.com/wiki/Automated_Installation
- Hetzner Rescue System: https://docs.hetzner.com/robot/dedicated-server/troubleshooting/hetzner-rescue-system/

### Glossary

| Term | Definition |
|------|------------|
| **TUI** | Terminal User Interface - text-based graphical interface |
| **Bubbletea** | Go framework for building terminal apps using Elm architecture |
| **Lipgloss** | Go library for styling terminal output |
| **ZFS** | Zettabyte File System - advanced filesystem with RAID support |
| **vmbr** | Virtual Machine Bridge - network bridge for VMs |
| **NAT** | Network Address Translation - allows private IPs to access internet |
| **QEMU** | Quick Emulator - hardware virtualization tool |
| **Tailscale** | Mesh VPN service using WireGuard |

### Open Questions

1. ~~Should we support IPv6 configuration?~~ → Deferred to v2
2. ~~Should we add Let's Encrypt SSL support?~~ → Deferred to v2
3. ~~Should we support multiple repository types?~~ → Deferred to v2
4. How to handle partial installation failure recovery?
5. Should config file support multiple server profiles?

---

## Task Master Integration Notes

This PRD is structured for `task-master parse-prd` compatibility:

- **Capabilities** → Top-level tasks
- **Features** → Subtasks under capabilities
- **Dependency Graph** → Explicit task dependencies
- **Phases** → Task priority ordering
- **Test Strategy** → Test generation guidance

The dependency chain ensures:
1. Foundation modules built first (config, exec)
2. TUI foundation before screens
3. Installer steps before orchestration
4. All components before integration
5. Tests after implementation
6. Release after tests
