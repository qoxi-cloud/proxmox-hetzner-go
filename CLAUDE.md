# CLAUDE.md - Project Instructions

> This file contains instructions for Claude AI when working on this project.

## Project Overview

**Name**: proxmox-hetzner-go  
**Purpose**: TUI-based installer for Proxmox VE on Hetzner dedicated servers  
**Language**: Go 1.24+  
**TUI Framework**: Bubbletea (github.com/charmbracelet/bubbletea)  
**Architecture**: RPG (Repository Planning Graph) methodology

## Language Requirements

**All content in this repository MUST be in English only.** This includes:
- Commit messages
- Pull request titles and descriptions
- Code comments
- Documentation files
- Variable and function names
- Log messages and user-facing strings
- Branch names

## Development Environment

### Tool Versions

Defined in `.tool-versions`:
```
golang 1.24.0
nodejs 24.11.0
```

### MCP Servers

MCP servers are installed locally via `package.json` (not globally).
Configuration is in `.mcp.json` using `./node_modules/.bin/` paths.

### Cross-Platform Development

| Environment | Platform | Purpose |
|-------------|----------|---------|
| **Development** | macOS | Coding, unit tests, TUI development |
| **Testing** | Linux VM / Hetzner | Integration tests, real hardware tests |
| **Production** | Hetzner Rescue System | Actual Proxmox installation |

### Building for Linux from macOS

```bash
# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o build/pve-install-linux ./cmd/pve-install

# Or use Makefile
make build-linux
```

### Platform Differences to Consider

| Aspect | macOS | Linux (Hetzner) |
|--------|-------|-----------------|
| Network interfaces | `en0`, `en1` | `eth0`, `enp0s31f6` |
| Disk listing | `diskutil list` | `lsblk` |
| IP commands | BSD syntax | GNU iproute2 |
| Paths | `/Users/...` | `/root/...` |
| ZFS | Not native | Supported |
| KVM | Not available | Available |

### What Works on macOS

✅ All Go code compilation  
✅ TUI development and testing  
✅ Unit tests with MockExecutor  
✅ Config file parsing  
✅ Validation logic  
✅ Code formatting and linting  

### What Requires Linux

❌ Real ZFS operations  
❌ Network bridge testing  
❌ KVM/QEMU with acceleration  
❌ Actual Proxmox installation  
❌ Integration tests with real hardware  

### Testing Strategy

```
macOS (Development)
├── Unit tests (go test ./...)
├── TUI tests (teatest)
├── Mocked system commands
└── Code coverage

Linux VM (CI/CD)
├── Integration tests
├── Real command execution
└── Docker-based testing

Hetzner Server (Manual)
├── E2E testing
├── Real installation
└── Hardware validation
```

### Recommended Workflow

1. **Develop on macOS** — Write code, run unit tests
2. **Push to GitHub** — CI runs Linux tests
3. **Test on Hetzner** — Manual E2E testing when needed
4. **Release** — Cross-compiled Linux binaries

## Tech Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.24+ |
| TUI Framework | Bubbletea | v1.2+ |
| TUI Components | Bubbles | v0.20+ |
| Styling | Lipgloss | v1.0+ |
| CLI Framework | Cobra | v1.8+ |
| Config Management | Viper | v1.19+ |
| Testing | Go testing + testify | v1.9+ |
| TUI Testing | teatest | latest |
| Version Management | asdf | latest |
| Release | goreleaser | latest |

## Project Structure (RPG-aligned)

```
proxmox-hetzner-go/
├── cmd/
│   └── pve-install/
│       └── main.go                 # CLI entry point (Cobra)
├── internal/
│   ├── config/                     # Capability: Configuration Management
│   │   ├── config.go              # Config struct & defaults
│   │   ├── validation.go          # Input validation
│   │   ├── env.go                 # Environment variables
│   │   └── file.go                # YAML file operations
│   ├── exec/                       # Capability: Command Execution
│   │   ├── command.go             # Executor interface & RealExecutor
│   │   └── mock.go                # MockExecutor for tests
│   ├── tui/                        # Capability: Terminal User Interface
│   │   ├── model.go               # Main Bubbletea model
│   │   ├── styles.go              # Lipgloss styles
│   │   ├── screens.go             # Screen rendering
│   │   ├── navigation.go          # Screen flow logic (NEW)
│   │   ├── messages.go            # Custom tea.Msg types
│   │   └── keys.go                # Key bindings
│   ├── installer/                  # Capability: Installation Engine
│   │   ├── installer.go           # Main orchestration
│   │   ├── preflight.go           # Pre-flight checks (NEW)
│   │   ├── detection.go           # Hardware detection
│   │   ├── download.go            # ISO download with progress
│   │   ├── answerfile.go          # Answer file generation (NEW)
│   │   ├── proxmox.go             # Proxmox installation via QEMU
│   │   ├── network.go             # Network config generation
│   │   ├── ssh.go                 # SSH hardening
│   │   ├── system.go              # System optimization
│   │   ├── tailscale.go           # Tailscale installation
│   │   ├── finalize.go            # Finalization & cleanup (NEW)
│   │   └── logging.go             # Log file management
│   └── testutil/
│       └── testutil.go            # Test helpers
├── test/
│   └── e2e_test.go                # End-to-end tests
├── configs/
│   └── example.yaml               # Example config file
├── scripts/
│   └── install.sh                 # One-line installer script
├── .github/
│   └── workflows/
│       ├── build.yml              # CI build workflow
│       └── release.yml            # Release workflow
├── .tool-versions                 # asdf: golang 1.24.0, nodejs 24.11.0
├── package.json                   # MCP server dependencies (local)
├── .mcp.json                      # MCP configuration
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── CLAUDE.md                      # This file
├── PRD_RPG.md                     # RPG-style Product Requirements
├── TASKS.md                       # GitHub tasks list
└── .goreleaser.yaml               # Release configuration
```

## Module Dependencies (RPG Dependency Graph)

### Foundation Layer (Phase 0) - No dependencies
- `config/config.go` - Config struct, DefaultConfig()
- `exec/command.go` - Executor interface, RealExecutor
- `exec/mock.go` - MockExecutor for testing

### Validation Layer (Phase 1)
- `config/validation.go` → depends on [config.go]
- `config/file.go` → depends on [config.go]
- `config/env.go` → depends on [config.go]
- `installer/logging.go` → no dependencies

### TUI Foundation (Phase 2)
- `tui/styles.go` → no dependencies
- `tui/messages.go` → no dependencies
- `tui/keys.go` → no dependencies

### TUI Model (Phase 3)
- `tui/model.go` → depends on [config, styles, messages]
- `tui/navigation.go` → depends on [model, config]
- `tui/screens.go` → depends on [model, styles]

### Installer Steps (Phase 4)
All depend on [exec, logging]:
- `installer/preflight.go`
- `installer/detection.go`
- `installer/download.go`
- `installer/answerfile.go`
- `installer/proxmox.go`
- `installer/network.go`
- `installer/ssh.go`
- `installer/system.go`
- `installer/tailscale.go`
- `installer/finalize.go`

### Orchestration (Phase 5)
- `installer/installer.go` → depends on [all installer steps]

### Integration (Phase 6)
- TUI ↔ Installer connection

### CLI (Phase 7)
- `cmd/pve-install/main.go` → depends on [config, tui, installer]

## Key Conventions

### Code Style

1. **Error Handling**: Always return errors, never panic
   ```go
   // ✅ Good
   func doSomething() error {
       if err := action(); err != nil {
           return fmt.Errorf("action failed: %w", err)
       }
       return nil
   }
   
   // ❌ Bad
   func doSomething() {
       if err := action(); err != nil {
           panic(err)
       }
   }
   ```

2. **Naming**: Use descriptive names, follow Go conventions
   ```go
   // ✅ Good
   type PreflightStep struct {}
   func (s *PreflightStep) Execute(ctx context.Context) error
   
   // ❌ Bad
   type Step struct {}
   func (s *Step) Do() error
   ```

3. **Comments**: Document exported types and functions
   ```go
   // Config holds all installation configuration.
   // It can be loaded from YAML files or environment variables.
   type Config struct {
       // Hostname is the server hostname (RFC 1123 compliant).
       Hostname string `yaml:"hostname"`
   }
   ```

4. **Step Interface**: All installer steps implement this interface
   ```go
   type Step interface {
       Name() string
       Execute(ctx context.Context) error
   }
   ```

### Bubbletea Patterns

1. **Model Structure**:
   ```go
   type Model struct {
       // State
       screen     Screen
       config     *config.Config
       err        error
       
       // UI Components
       inputs     []textinput.Model
       spinner    spinner.Model
       
       // Navigation
       cursor     int
       
       // Window
       width, height int
   }
   ```

2. **Update Pattern**:
   ```go
   func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case tea.KeyMsg:
           return m.handleKey(msg)
       case tea.WindowSizeMsg:
           m.width = msg.Width
           m.height = msg.Height
           return m, nil
       }
       return m, nil
   }
   ```

3. **View Pattern**:
   ```go
   func (m Model) View() string {
       switch m.screen {
       case ScreenWelcome:
           return m.viewWelcome()
       case ScreenHostname:
           return m.viewInput("Hostname", "Enter hostname", 0)
       // ...
       }
       return ""
   }
   ```

### Testing Patterns

1. **Unit Tests**: Use table-driven tests
   ```go
   func TestValidateHostname(t *testing.T) {
       tests := []struct {
           name    string
           input   string
           wantErr bool
       }{
           {"valid", "pve-server", false},
           {"empty", "", true},
           {"too long", strings.Repeat("a", 64), true},
       }
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := validateHostname(tt.input)
               if (err != nil) != tt.wantErr {
                   t.Errorf("validateHostname() error = %v, wantErr %v", err, tt.wantErr)
               }
           })
       }
   }
   ```

2. **Mock Executor**: Use for testing system commands
   ```go
   mock := exec.NewMockExecutor()
   mock.SetOutput("ip link show", "eth0: state UP")
   
   step := &DetectionStep{executor: mock, config: cfg}
   err := step.Execute(context.Background())
   
   assert.NoError(t, err)
   assert.Equal(t, "eth0", cfg.Network.InterfaceName)
   ```

3. **TUI Tests**: Use teatest package
   ```go
   func TestWelcomeScreen(t *testing.T) {
       m := tui.New(config.DefaultConfig())
       tm := teatest.NewTestModel(t, m)
       
       tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
       
       // Verify screen changed
       out := tm.FinalModel(t).(tui.Model)
       assert.Equal(t, tui.ScreenHostname, out.Screen())
   }
   ```

### Commit Message Format

Use emoji conventional commit format:

```text
<emoji> <type>: <short description>

<detailed explanation of changes>

Changes:
- Bullet point list of specific changes
- Each change on its own line
- Focus on "what" and "why"

<additional context if needed>
```

**Emoji Reference:**

| Emoji | Type | Description |
|-------|------|-------------|
| ✨ | `feat` | New features |
| 🐛 | `fix` | Bug fixes |
| 🔒️ | `security` | Security fixes |
| ♻️ | `refactor` | Code restructuring |
| 📝 | `docs` | Documentation |
| 🔧 | `chore` | Configuration, tooling |
| ⚡️ | `perf` | Performance improvements |
| 🩹 | `fix` | Simple non-critical fixes |
| 🚑️ | `hotfix` | Critical hotfixes |
| ✅ | `test` | Adding or updating tests |
| 🏗️ | `build` | Build system changes |
| 👷 | `ci` | CI/CD changes |

**Example:**

```text
✨ feat: add pre-flight checks step

Implemented PreflightStep for installation validation.

Changes:
- Added internal/installer/preflight.go
- Check root, required tools, network, KVM
- Return all failed checks (not just first)
```

### Pull Request Format

When creating pull requests, use the template at `.github/pull_request_template.md`. PR titles must follow the same emoji conventional format as commit messages.

**PR Title Format:**

```text
<emoji> <type>: <short description>
```

**PR Body Structure:**

```markdown
## Summary

Brief description of what this PR does

## Changes

- List specific changes
- Each change on its own line

## Type of Change

- [x] New feature (`feat`)

## Testing

- [x] Unit tests pass (`go test ./...`)
- [x] Linting passes (`golangci-lint run`)
- [x] Manual testing performed

## Checklist

- [x] Code follows project conventions (see CLAUDE.md)
- [x] All content is in English
- [x] Commit messages follow emoji conventional format
- [x] Tests added for new functionality
```

## Important Rules

### DO ✅

- Always use context.Context for cancellation
- Always use the Executor interface for system commands
- Always validate user input before using
- Always log important operations to the log file
- Always handle window resize in TUI
- Always write tests for new functionality
- Use `internal/` for packages not meant for external use
- Use meaningful commit messages
- Follow RPG dependency order when implementing

### DON'T ❌

- Don't use `os/exec` directly, use the Executor interface
- Don't hardcode paths, use constants or config
- Don't skip error handling
- Don't use global state
- Don't write to stdout directly in TUI mode (use View())
- Don't block the main thread with long operations
- Don't create circular dependencies between modules

## Feature Flags

These features are explicitly **NOT** included:

- ❌ Non-interactive mode (removed from scope)
- ❌ Test/dry-run mode (removed from scope)
- ❌ IPv6 configuration (deferred to v2)
- ❌ Let's Encrypt SSL (deferred to v2)
- ❌ Multiple repository types (deferred to v2)

## CLI Flags

| Flag | Description |
|------|-------------|
| `-c, --config` | Load configuration from YAML file |
| `-s, --save-config` | Save configuration to file after input |
| `-v, --verbose` | Enable verbose logging |
| `-h, --help` | Show help |
| `--version` | Show version |

## Configuration

### Priority Order (highest to lowest)
1. User input in TUI
2. Environment variables
3. Config file values
4. Default values

### Sensitive Fields (never saved to file)
- `RootPassword`
- `SSHPublicKey`
- `TailscaleAuthKey`

## Installation Steps Order

Steps execute in this exact order (from RPG Phase 4):

1. **Pre-flight checks** - Root, tools, network, KVM validation
2. **Hardware detection** - Interfaces, disks, SSH keys auto-detect
3. **Download Proxmox ISO** - With progress callback
4. **Generate answer file** - TOML format for auto-install
5. **Install Proxmox** - Via QEMU with KVM acceleration
6. **Configure networking** - Bridge setup (NAT/external/both)
7. **Apply SSH hardening** - Modern ciphers, key-only auth
8. **System optimization** - Packages, ZFS ARC, timezone
9. **Install Tailscale** - If enabled in config
10. **Finalize** - Restart services, cleanup, verify

## Common Tasks

### Adding a New Screen

1. Add screen constant to `internal/tui/model.go`:
   ```go
   const (
       ScreenWelcome Screen = iota
       ScreenHostname
       ScreenNewScreen  // Add here
       // ...
   )
   ```

2. Add view method to `internal/tui/screens.go`:
   ```go
   func (m Model) viewNewScreen() string {
       // Render screen
   }
   ```

3. Update `View()` switch statement in `model.go`
4. Update navigation in `navigation.go` (nextScreen, prevScreen)
5. Add tests

### Adding a New Installation Step

1. Create new file in `internal/installer/` (e.g., `newstep.go`)
2. Implement Step interface:
   ```go
   type NewStep struct {
       config   *config.Config
       executor exec.Executor
       logger   *Logger
   }
   
   func (s *NewStep) Name() string { return "New Step" }
   func (s *NewStep) Execute(ctx context.Context) error { ... }
   ```
3. Add to steps slice in `installer.go`
4. Add callbacks integration
5. Add unit tests with MockExecutor
6. Update integration tests

### Adding a New Config Field

1. Add field to `Config` struct in `internal/config/config.go`
2. Add yaml/env tags
3. Add validation in `internal/config/validation.go`
4. Add to `DefaultConfig()` if has default
5. Update `LoadFromEnv()` if env var needed
6. Add tests for validation
7. Update TUI screens if user input needed
8. Update example.yaml

## Running the Project

```bash
# First-time setup (after cloning)
asdf install          # Install Go and Node.js
npm install           # Install MCP servers

# Build for current platform (macOS)
make build

# Build for Linux (cross-compile)
make build-linux

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Format code
make fmt

# Clean build artifacts
make clean
```

## Troubleshooting

### TUI Not Rendering Correctly
- Check terminal supports ANSI colors
- Ensure `tea.WithAltScreen()` is used
- Check window size handling

### Tests Failing
- Ensure MockExecutor is properly configured
- Check test isolation (no shared state)
- Verify teardown in tests

### Cross-compilation Issues
- Verify `GOOS=linux GOARCH=amd64` is set
- Check CGO is disabled (CGO_ENABLED=0)

## GitHub CLI

Use `gh` (GitHub CLI) for all GitHub operations:

```bash
# Issues
gh issue list
gh issue create --title "Title" --body "Body"
gh issue view 123
gh issue close 123

# Pull Requests
gh pr create --title "Title" --body "Body"
gh pr list
gh pr view 123
gh pr merge 123

# Workflows
gh run list
gh run view 123

# API (for advanced operations)
gh api repos/{owner}/{repo}/issues
```

## Resources

- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [Cobra CLI](https://github.com/spf13/cobra)
- [Proxmox Auto-Install](https://pve.proxmox.com/wiki/Automated_Installation)
- [RPG Methodology](https://github.com/eyaltoledano/claude-task-master)
- [GitHub CLI](https://cli.github.com/manual/)

## Project Documents

| Document | Description |
|----------|-------------|
| `PRD_RPG.md` | RPG-style Product Requirements (capabilities, dependencies, phases) |
| `README.md` | User documentation |
| `CLAUDE.md` | This file - AI instructions |

---

## Recommended MCP Servers

MCP (Model Context Protocol) servers extend Claude's capabilities for developing this project.

Configuration is in a separate file: `.mcp.json`

### 🎯 Task Management

| MCP Server | Description |
|------------|-------------|
| **task-master-ai** | AI-powered task management (parse PRD, track tasks, generate subtasks) |

**Standard Tools (15)**: get_tasks, next_task, get_task, set_task_status, update_subtask, parse_prd, expand_task, initialize_project, analyze_project_complexity, expand_all, add_subtask, remove_task, generate, add_task, complexity_report

### 🔧 Core Development

| MCP Server | Description |
|------------|-------------|
| **filesystem** | Secure file operations |
| **memory** | Persistent memory between sessions |

### 🔍 Code & Documentation

| MCP Server | Description |
|------------|-------------|
| **context7** | Up-to-date library documentation |
| **sequential-thinking** | Step-by-step problem solving |
| **fetch** | Web content fetching |

### MCP Server Usage

| Task | MCP Server |
|------|------------|
| Parse PRD to tasks | task-master-ai |
| Get next task | task-master-ai |
| Update task status | task-master-ai |
| Read/write code | filesystem |
| Search documentation | fetch, context7 |
| Complex tasks | sequential-thinking |
| Save context | memory |
