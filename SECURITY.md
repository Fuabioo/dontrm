# Security Policy

## Overview

`dontrm` is a safety wrapper around the `rm` command designed to prevent catastrophic system deletions. Security is our highest priority.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < 1.0   | :x:                |

## What dontrm Protects Against

### Protected Operations

dontrm blocks these dangerous patterns:

1. **Top-level system paths**
   - `/`, `/bin`, `/boot`, `/dev`, `/etc`, `/lib`, `/lib64`, `/opt`, `/proc`, `/root`, `/run`, `/sbin`, `/srv`, `/sys`, `/usr`, `/var`
   - Example: `dontrm -rf /` → **BLOCKED**

2. **Wildcard operations on system directories**
   - `/usr/bin/*`, `/etc/*`, etc.
   - Example: `dontrm -rf /etc/*` → **BLOCKED**

3. **Recursive glob patterns**
   - `/**/*` and similar patterns that expand to system paths
   - Example: `dontrm /**/*` → **BLOCKED**

### What dontrm DOES NOT Protect

dontrm is not a comprehensive safety solution. It does **NOT** protect against:

- **User home directories**: `/home/user` can be deleted (intentional design)
- **Data directories**: `/data`, `/mnt`, `/media` wildcards are allowed
- **Specific files**: `/etc/passwd` (individual files in system dirs)
- **Subdirectories**: `/usr/bin/go/*` (subdirectories of system dirs)
- **Non-standard system paths**: Custom installation directories
- **Network/remote filesystems**: NFS, SMB mounts
- **Symlink exploitation**: Following symlinks to protected paths

## Security Considerations

### Use Cases

dontrm is designed for:
- Preventing accidental `sudo rm -rf /` disasters
- Catching copy-paste errors from internet commands
- Protecting against typos in wildcard patterns

dontrm is **NOT** designed for:
- Preventing malicious actions by authorized users
- Filesystem access control (use proper permissions instead)
- Comprehensive system protection (use backups, snapshots, immutable infrastructure)

### DRY_RUN Mode

Always test dangerous commands with `DRY_RUN` first:

```bash
# Test before running
DRY_RUN=1 dontrm -rf /some/path

# If safe, run for real
dontrm -rf /some/path
```

### Sudo Usage

dontrm requires sudo for operations that need elevated privileges. Be aware:
- Running with sudo bypasses user-level protections
- Always double-check commands before using sudo
- Consider using `DRY_RUN=1` even with sudo

### Known Limitations

1. **Symlink Following**: dontrm checks the provided path, not what symlinks resolve to
2. **Race Conditions**: File system state can change between check and execution
3. **Glob Expansion**: Shell expands globs before dontrm sees them (usually safe, but be aware)
4. **Custom System Paths**: If you've installed system software in non-standard locations, those aren't protected

## Reporting a Vulnerability

### What to Report

Please report:
- Bypasses of safety checks
- Ways to delete protected paths
- Race conditions in protection logic
- Misleading error messages that could cause dangerous operations
- Security issues in test infrastructure

### What NOT to Report

These are known and expected:
- User home directories can be deleted (by design)
- Specific files in system directories can be deleted (by design)
- The tool can be bypassed by using `/usr/bin/rm` directly (by design)

### How to Report

**For security vulnerabilities, please email:** fabio@fuabioo.com

**DO NOT** open a public GitHub issue for security vulnerabilities.

Include in your report:
1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if you have one)

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: Within 7 days
  - High: Within 14 days
  - Medium: Within 30 days
  - Low: Next release cycle

## Security Testing

### Test Safety

Our test infrastructure prioritizes safety:

1. **Docker Isolation**: All tests run in Docker containers (unit + E2E)
2. **Control File Mechanism**: Tests verify they're in safe environment
3. **No Local Execution**: Tests cannot run on host machine
4. **85% Coverage**: Comprehensive test coverage ensures protection works (accounts for untestable main() wrapper)

See [TESTING.md](TESTING.md) for details.

### Continuous Integration

Every commit is automatically tested for:
- Protection logic correctness (unit tests)
- Real-world binary behavior (E2E tests in bash/zsh/fish)
- Code quality via linting
- Race conditions via race detector
- Coverage maintenance (85% minimum)

## Best Practices for Users

### General Safety

1. **Always use DRY_RUN first** for untested patterns
2. **Read error messages carefully** - they explain what was blocked and why
3. **Use absolute paths** when possible for clarity
4. **Verify wildcards expand correctly** before running
5. **Maintain backups** - dontrm is not a backup solution

### Integration Safety

If integrating dontrm into scripts or tools:

```bash
# Good: Check exit status
if dontrm "$file"; then
    echo "Deleted successfully"
else
    echo "Deletion blocked or failed"
fi

# Good: Use DRY_RUN for validation
if DRY_RUN=1 dontrm "$path" 2>/dev/null; then
    # Path is safe to delete
    dontrm "$path"
fi
```

### Alias Configuration

If aliasing `rm` to `dontrm`:

```bash
# In .bashrc or .zshrc
alias rm='dontrm'

# Keep a way to access real rm if needed
alias dangerrm='/usr/bin/rm'  # Use with extreme caution
```

## Responsible Disclosure

We appreciate security researchers who:
- Report vulnerabilities privately first
- Allow reasonable time for fixes before public disclosure
- Provide clear reproduction steps
- Suggest potential fixes

In return, we commit to:
- Acknowledging receipt within 48 hours
- Providing regular status updates
- Crediting researchers in release notes (unless they prefer anonymity)
- Fixing critical issues promptly

## Security Champions

Security contributions are recognized in:
- Release notes
- Project README
- Security hall of fame (for significant findings)

Thank you for helping keep dontrm safe!

## Additional Resources

- [TESTING.md](TESTING.md) - Test safety and infrastructure
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guidelines
- [README.md](README.md) - Usage and installation
