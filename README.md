# dontrm

[![Test and Lint](https://github.com/Fuabioo/dontrm/actions/workflows/test.yml/badge.svg)](https://github.com/Fuabioo/dontrm/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/Fuabioo/dontrm/branch/main/graph/badge.svg)](https://codecov.io/gh/Fuabioo/dontrm)
[![Go Report Card](https://goreportcard.com/badge/github.com/Fuabioo/dontrm)](https://goreportcard.com/report/github.com/Fuabioo/dontrm)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**A safe wrapper around `rm` that prevents catastrophic system deletions.**

Tired of fearing every `rm -rf` command? `dontrm` is a drop-in replacement for `rm` that blocks dangerous operations like `rm -rf /` or `rm -rf /etc/*` while allowing normal file deletions to proceed safely.

## What It Protects

### ✅ Blocks These Dangerous Operations

- **Top-level system directories**: `/`, `/bin`, `/boot`, `/dev`, `/etc`, `/lib`, `/lib64`, `/opt`, `/proc`, `/root`, `/run`, `/sbin`, `/srv`, `/sys`, `/usr`, `/var`
- **System directory wildcards**: `/usr/bin/*`, `/etc/*`, etc.
- **Works even with sudo**: Protection cannot be bypassed with elevated privileges
- **Works with common flags**: `-rf`, `--no-preserve-root`, etc.

### ⚠️ Does NOT Protect

- **User home directories**: `/home/user` can be deleted (by design)
- **Data directories**: `/data`, `/mnt`, `/media` directories
- **Specific files**: Individual files in system directories like `/etc/passwd`
- **Subdirectories**: Subdirectories like `/usr/bin/go/*`
- **Symlink following**: Symlinks that resolve to protected paths

See [SECURITY.md](SECURITY.md) for complete details on protection scope and limitations.

## Installation

### Quick Install (Recommended)

Visit https://dontrm.fuabioo.com/#installation for the latest installation script.

### Build from Source

**Prerequisites**: Go 1.25+

```sh
# Clone the repository
git clone https://github.com/Fuabioo/dontrm.git
cd dontrm

# Build and install
go build -ldflags="-s -w" -o dontrm .
sudo mv dontrm /usr/bin/dontrm

# Or use just (if installed)
just build
just install
```

### Alias Setup (Optional but Recommended)

To make `rm` use `dontrm` automatically, add an alias to your shell configuration:

**Bash** (`~/.bashrc`):
```bash
alias rm='dontrm'
alias unsafe-rm='/usr/bin/rm'  # Keep access to real rm (use with EXTREME caution)
```

**Zsh** (`~/.zshrc`):
```zsh
alias rm='dontrm'
alias unsafe-rm='/usr/bin/rm'
```

**Fish** (`~/.config/fish/config.fish`):
```fish
alias rm 'dontrm'
alias unsafe-rm '/usr/bin/rm'
```

**Sh/Dash** (`~/.profile`):
```sh
alias rm='dontrm'
alias unsafe-rm='/usr/bin/rm'
```

After adding the alias, restart your shell or source the config file:
```bash
source ~/.bashrc    # For bash
source ~/.zshrc     # For zsh
source ~/.profile   # For sh/dash
# For fish, just restart the shell
```

#### Making the Alias Work with Sudo

By default, `sudo rm` won't use your alias because sudo runs commands in a clean environment. Here are **shell-agnostic** approaches to make `sudo rm` use `dontrm`:

---

**Option 1: System-Wide Installation (Recommended - Works for All Shells)**

Install `dontrm` system-wide so it's available to all users including root:

```bash
# Build dontrm
go build -ldflags="-s -w" -o dontrm .

# Install to system location
sudo install -m 755 dontrm /usr/local/bin/dontrm

# Optional: Create symlink so 'rm' points to dontrm
sudo ln -sf /usr/local/bin/dontrm /usr/local/bin/rm

# Ensure /usr/local/bin is in PATH before /usr/bin
# Add to /etc/environment or /etc/profile:
export PATH="/usr/local/bin:$PATH"
```

**Pros**: Works immediately for all users, all shells, and `sudo` commands
**Cons**: System-wide change affects all users
**Security**: High - no shell configuration needed

---

**Option 2: Shell-Specific Alias for Root User**

Add an alias to root's shell configuration. **Note**: This only works for interactive sudo shells (`sudo -i`, `sudo -s`, `sudo bash`), not for direct commands like `sudo rm file.txt`.

**For Bash** (root's `~/.bashrc` or `/root/.bashrc`):
```bash
sudo tee -a /root/.bashrc >/dev/null <<'EOF'
# Use dontrm instead of rm
alias rm='dontrm'
EOF
```

**For Zsh** (root's `/root/.zshrc`):
```zsh
sudo tee -a /root/.zshrc >/dev/null <<'EOF'
# Use dontrm instead of rm
alias rm='dontrm'
EOF
```

**For Fish** (root's `/root/.config/fish/config.fish`):
```fish
# Create config directory if it doesn't exist
sudo mkdir -p /root/.config/fish

# Add alias
sudo tee -a /root/.config/fish/config.fish >/dev/null <<'EOF'
# Use dontrm instead of rm
alias rm 'dontrm'
EOF
```

**Pros**: Simple, doesn't affect non-sudo commands
**Cons**: Only works for `sudo -i` or `sudo bash`, not `sudo rm file.txt`
**Security**: High - isolated to root's shell

---

**Option 3: Alias Expansion in Bash/Zsh (Shell-Specific)**

**⚠️ Only works in Bash and Zsh, not Fish or other shells.**

Add to your user's shell config (`~/.bashrc` or `~/.zshrc`):
```bash
alias sudo='sudo '  # Trailing space makes sudo expand aliases
```

**Pros**: Works for `sudo rm file.txt` commands
**Cons**: Only works in bash/zsh, affects ALL sudo commands, can cause unexpected behavior
**Security**: Medium - changes sudo behavior globally

---

**Option 4: Configure sudoers (Advanced - Shell Agnostic)**

Preserve environment variables through sudo by editing `/etc/sudoers`:

```bash
# Edit sudoers safely
sudo visudo

# Add this line to preserve alias-related environment
Defaults env_keep += "BASH_FUNC_*"
```

**⚠️ Warning**: This has security implications and may not work reliably across all shells.

**Pros**: Shell-agnostic
**Cons**: Complex, security trade-offs, unreliable
**Security**: Lower - preserves environment variables

---

### Comparison Table

| Approach | Direct `sudo rm` | Interactive `sudo -i` | Shell Agnostic | Complexity | Recommended |
|----------|------------------|----------------------|----------------|------------|-------------|
| System-wide install | ✅ Yes | ✅ Yes | ✅ Yes | Medium | ⭐ **Best** |
| Root shell alias | ❌ No | ✅ Yes | ✅ Yes | Easy | ✅ Good |
| Alias expansion | ✅ Yes* | ✅ Yes | ❌ No (bash/zsh only) | Easy | ⚠️ Limited |
| sudoers config | ⚠️ Maybe | ⚠️ Maybe | ✅ Yes | Hard | ❌ Avoid |

*Only for bash/zsh

---

### Testing Your Setup

```bash
# Test regular dontrm (works for all methods)
dontrm version
# Expected output: DON'T rm! dev

# Test sudo with interactive shell (works for Options 1 & 2)
sudo -i
rm version  # Should show "DON'T rm!" not GNU rm
exit

# Test direct sudo rm (only works for Options 1 & 3)
DRY_RUN=1 sudo rm -rf /etc
# Expected: ⛔ Blocked dangerous operation

# If you see GNU rm output or actual deletion attempts, the alias isn't working
```

**Troubleshooting:**
- If `sudo rm` shows GNU rm, try `sudo -i` then `rm` to test if the root alias works
- For fish users: Shell aliases don't transfer to sudo by default - use Option 1 (system-wide install)
- Verify dontrm is installed: `which dontrm` and `sudo which dontrm`

## Quick Start

### Basic Usage

```sh
# Check version
dontrm version

# Delete a file normally
dontrm file.txt

# Delete directory recursively
dontrm -rf ./old-project/

# This will be BLOCKED
dontrm -rf /etc
# Error: ⛔ Blocked dangerous operation: known top level match: /etc

# This will also be BLOCKED
sudo dontrm -rf /
# Error: ⛔ Blocked dangerous operation: known top level match: /
```

### DRY_RUN Mode

Always test dangerous-looking commands with `DRY_RUN` first:

```sh
# Test mode - checks safety but doesn't actually delete
DRY_RUN=1 dontrm -rf /some/path/

# If no error, run for real
dontrm -rf /some/path/
```

`DRY_RUN` accepts `1`, `true`, or any truthy value.

## Configuration

`dontrm` supports the following environment variables for customization:

### Environment Variables

#### `DRY_RUN`
- **Description**: Test mode - performs safety checks but doesn't execute deletions
- **Values**: `1`, `true` (case-insensitive)
- **Default**: Not set (deletions execute normally)
- **Example**: `DRY_RUN=1 dontrm -rf /some/path/`

#### `DONTRM_RM_PATH`
- **Description**: Custom path to the `rm` binary
- **Values**: Absolute path to `rm` executable
- **Default**: Automatically detected via `PATH` lookup
- **Use Case**: Custom `rm` installations or testing
- **Example**: `DONTRM_RM_PATH=/bin/rm dontrm file.txt`

### Cross-Platform Support

`dontrm` works on:
- **Linux** (all distributions)
- **macOS** (automatically finds `rm` in `/bin/rm` or `/usr/bin/rm`)
- **Windows** (limited - requires WSL or Cygwin with `rm` installed)

The `rm` binary is automatically detected using your system's `PATH`. If you have a custom `rm` installation, use the `DONTRM_RM_PATH` environment variable to specify its location.

**Platform-Specific Notes:**
- **Linux**: Default `rm` path is typically `/usr/bin/rm` or `/bin/rm`
- **macOS**: Default `rm` path is typically `/bin/rm`
- **Custom installations**: Set `DONTRM_RM_PATH` to your `rm` location

## How It Works

1. **Argument Validation**: Before executing any deletion, `dontrm` inspects all arguments
2. **Pattern Matching**: Checks arguments against known dangerous system paths
3. **Glob Expansion**: Evaluates wildcards to detect if they expand to system directories
4. **Safety First**: If any dangerous pattern is detected, operation is blocked with clear error
5. **Otherwise, Execute**: If safe, locates `rm` in your PATH and executes it with the arguments

```
┌─────────────┐
│ dontrm args │
└──────┬──────┘
       │
       ▼
┌─────────────────────────┐
│ Check system paths?     │──── YES ──▶ BLOCK ⛔
└───────────┬─────────────┘
            │
           NO
            │
            ▼
┌─────────────────────────┐
│ Check glob expansions?  │──── YES ──▶ BLOCK ⛔
└───────────┬─────────────┘
            │
           NO
            │
            ▼
┌─────────────────────────┐
│ Find rm in PATH         │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Execute rm with args    │──── ✅
└─────────────────────────┘
```

## Development

### Prerequisites

- Go 1.25+
- Docker (required for testing)
- [`just`](https://github.com/casey/just) command runner (recommended)
- [golangci-lint](https://golangci-lint.run/welcome/install/)

### Testing

**CRITICAL**: All tests run in Docker containers for safety. Never run `go test` directly.

```sh
# Run unit tests (Go tests in Docker)
just test

# Run E2E tests (tests actual binary in bash/zsh/fish)
just e2e

# Run all tests (unit + E2E)
just test-all

# Check coverage (requires 85% minimum)
just coverage

# Run linting
just lint
```

See [TESTING.md](TESTING.md) for comprehensive testing documentation.

### Building

```sh
# Build binary
just build

# Clean artifacts
just clean

# Rebuild Docker test images
just rebuild-test-image
just rebuild-e2e-image
```

## Testing Philosophy

This project employs **defense-in-depth testing**:

- **Unit Tests**: Go tests validate logic correctness (87.3% coverage)
  - Run in Docker with control file safety check
  - Test argument parsing, pattern matching, edge cases

- **E2E Tests**: Bash script tests validate real-world usage
  - Tests actual compiled binary in Ubuntu environment
  - Validates bash, zsh, and fish compatibility
  - Tests sudo usage, exit codes, error messages
  - Creates and deletes real files (safely in Docker)

Both test suites run exclusively in Docker and cannot execute on host machines (enforced by control file checks).

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

**Quick contributing checklist**:
- ✅ All tests pass (`just test-all`)
- ✅ Coverage ≥ 85% (`just coverage`)
- ✅ Linting passes (`just lint`)
- ✅ Tests run in Docker (enforced automatically)

## Security

This project deals with dangerous file operations. Security is our top priority.

- See [SECURITY.md](SECURITY.md) for security policy
- Report vulnerabilities to: fabio@fuabioo.com
- All tests run in isolated Docker containers
- Multiple safety layers prevent accidental host PC damage

## Documentation

- **[TESTING.md](TESTING.md)** - Comprehensive testing guide, Docker safety mechanisms
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development setup, workflow, code standards
- **[SECURITY.md](SECURITY.md)** - Security policy, protection scope, vulnerability reporting

## FAQ

### Can I bypass dontrm's protection?

Yes, by calling `/usr/bin/rm` directly. `dontrm` is designed to prevent **accidents**, not malicious actions. If you really need to delete system files, use the real `rm` directly (but please don't).

### Does this slow down file deletion?

Negligibly. Validation adds ~1-2ms overhead for simple operations. For recursive operations on thousands of files, the actual deletion time far exceeds validation time.

### Why not protect user home directories?

By design, users should have full control over their home directories. The goal is to prevent **system-destroying** operations, not restrict legitimate user file management.

### Can I configure which paths are protected?

Not currently. Protection list is hardcoded based on standard Linux FHS (Filesystem Hierarchy Standard). See the [TODO](#roadmap) section for planned features.

## Roadmap

Track progress and suggest features in [GitHub Issues](https://github.com/Fuabioo/dontrm/issues).

Planned features:
- [ ] Configurable protection paths (config file support)
- [ ] Verbose/debug mode for troubleshooting
- [ ] Interactive mode (confirm before deletion)
- [ ] Trash/recycle bin functionality
- [ ] Plugin system for custom safety rules

Completed:
- [x] Comprehensive test suite with Docker isolation
- [x] CI/CD with automated testing
- [x] E2E tests with multi-shell support
- [x] Cross-platform builds (Linux/macOS/Windows)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by countless stories of `sudo rm -rf /` disasters across the internet. This is a small attempt to prevent future tragedies.

---

**⚠️ Remember**: `dontrm` is a safety net, not a security tool. Always double-check commands, maintain backups, and use `DRY_RUN=1` when in doubt.
