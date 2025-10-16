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

To make `rm` use `dontrm` automatically:

```bash
# Add to ~/.bashrc or ~/.zshrc
alias rm='dontrm'

# Keep access to real rm if needed (use with EXTREME caution)
alias unsafe-rm='/usr/bin/rm'
```

#### Making the Alias Work with Sudo

By default, `sudo rm` won't use your alias because sudo runs commands in a clean environment. To make `sudo rm` use `dontrm`:

**Option 1: Add alias to root's bashrc (Recommended)**

```bash
# Edit root's bashrc
sudo nano /root/.bashrc

# Add the alias
alias rm='dontrm'

# Reload root's bashrc
sudo bash -c "source /root/.bashrc"
```

**Option 2: Use sudo with alias expansion**

```bash
# Add to your ~/.bashrc or ~/.zshrc
alias sudo='sudo '  # Note the trailing space - this makes sudo expand aliases

# Now 'sudo rm' will use your alias
# But this affects ALL sudo commands, not just rm
```

**Option 3: Create a wrapper script**

```bash
# Create a wrapper script
sudo tee /usr/local/bin/rm-safe >/dev/null <<'EOF'
#!/bin/bash
exec /usr/bin/dontrm "$@"
EOF

sudo chmod +x /usr/local/bin/rm-safe

# Add to root's bashrc
sudo bash -c "echo 'alias rm=\"/usr/local/bin/rm-safe\"' >> /root/.bashrc"
```

**Testing sudo alias:**

```bash
# Test if sudo uses dontrm
sudo rm --version  # Should show "DON'T rm!" not GNU rm version

# If it shows GNU rm version, the alias isn't active for sudo
```

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

## How It Works

1. **Argument Validation**: Before executing any deletion, `dontrm` inspects all arguments
2. **Pattern Matching**: Checks arguments against known dangerous system paths
3. **Glob Expansion**: Evaluates wildcards to detect if they expand to system directories
4. **Safety First**: If any dangerous pattern is detected, operation is blocked with clear error
5. **Otherwise, Execute**: If safe, passes arguments directly to `/usr/bin/rm`

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
│ Execute /usr/bin/rm     │──── ✅
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
