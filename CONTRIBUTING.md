# Contributing to dontrm

Thank you for your interest in contributing to dontrm! This project implements a safe wrapper around the `rm` command to prevent catastrophic system deletions.

## Code of Conduct

Be respectful, constructive, and professional in all interactions.

## Getting Started

### Prerequisites

- Go 1.25 or later
- Docker (required for testing)
- `just` command runner (recommended) - Install: `cargo install just` or see [justfile.systems](https://github.com/casey/just)
- golangci-lint - Install: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s`

### Setup

```bash
# Clone the repository
git clone https://github.com/Fuabioo/dontrm.git
cd dontrm

# Install dependencies
go mod download

# Verify setup by running tests (in Docker)
just test
```

## Development Workflow

### Making Changes

1. **Fork and clone** the repository
2. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** following our coding standards
4. **Write tests** for new functionality (both unit and E2E if applicable)
5. **Run unit tests** to ensure logic correctness:
   ```bash
   just test
   ```
6. **Run E2E tests** to validate real-world behavior:
   ```bash
   just e2e
   ```
   Or run all tests at once:
   ```bash
   just test-all
   ```
7. **Check coverage** meets 85% threshold:
   ```bash
   just coverage
   ```
8. **Lint your code**:
   ```bash
   just lint
   ```
9. **Commit your changes** with a clear commit message
10. **Push to your fork** and create a Pull Request

### Commit Messages

Write clear, descriptive commit messages:

```
Add protection for /opt/* wildcard patterns

- Extend system path detection to include /opt
- Add test cases for /opt directory operations
- Update documentation with new protection details
```

## Testing Requirements

**CRITICAL**: This project has strict testing requirements for safety reasons.

### Test Safety

- **NEVER** run `go test` directly on your host machine
- **ALWAYS** use `just test` which runs tests in Docker
- Tests will panic if run outside Docker (by design)
- See [TESTING.md](TESTING.md) for detailed testing guide

### Coverage Requirements

- All code must maintain **85% test coverage** minimum (accounts for untestable main() wrapper)
- Run `just coverage` to check current coverage
- CI will fail if coverage drops below 85%
- Focus on testing edge cases and error paths

### Writing Tests

#### Unit Tests (main_test.go)

```go
func TestYourFunction(t *testing.T) {
    // Tests automatically check for Docker environment via TestMain
    // Focus on logic validation, not actual file operations

    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"test case 1", "input1", "expected1"},
        {"test case 2", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := yourFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

#### E2E Tests (e2e_test.sh)

Add new test functions to `e2e_test.sh` for testing real binary behavior:

```bash
test_your_feature() {
    test_header "Your Feature Name"

    cleanup_test

    # Create test file
    local test_file="/tmp/dontrm-e2e-test-yourfeature.txt"
    create_test_file "$test_file"

    # Test your feature
    output=$(dontrm --your-flag "$test_file" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        pass "Your feature works"
    else
        fail "Your feature failed" "Exit code: $exit_code"
    fi

    cleanup_test
}
```

Then add `test_your_feature` to the `main()` function.

**E2E Test Guidelines:**
- Always use `/tmp/dontrm-e2e-test-*` for file operations
- Use `pass()` and `fail()` helpers for consistent output
- Call `cleanup_test` at start and end of your test
- Test actual binary behavior, not Go functions
- Test cross-shell compatibility if relevant (bash/zsh/fish)

## Code Style

### Go Standards

- Follow standard Go formatting: `gofmt`
- Use `goimports` for import organization
- Follow Go naming conventions
- Document exported functions and types

### Linting

All code must pass golangci-lint:

```bash
just lint
```

Common issues to avoid:
- Unchecked errors
- Unused variables
- Inefficient assignments
- Missing documentation on exported symbols

### Security Considerations

This project deals with dangerous file operations. When contributing:

1. **Never bypass safety checks** - All system path protections must remain intact
2. **Test dangerous patterns thoroughly** - Add tests for any new protection logic
3. **Document security implications** - Explain why protection is needed
4. **Use DRY_RUN for manual testing** - Always test with `DRY_RUN=1` first

## Pull Request Process

1. **Ensure all checks pass**:
   - Tests pass in Docker (unit + E2E)
   - Coverage meets 85%
   - Linting passes
   - Build succeeds

2. **Update documentation** if needed:
   - Update README.md for user-facing changes
   - Update TESTING.md for test changes
   - Add comments for complex logic

3. **Describe your changes**:
   - What problem does this solve?
   - How does it work?
   - Are there breaking changes?
   - Have you tested it thoroughly?

4. **Request review** and address feedback

5. **Squash commits** if requested before merge

## Project Structure

```
dontrm/
├── main.go              # Main application logic
├── main_test.go         # Unit test suite
├── justfile             # Task runner commands
├── Dockerfile.test      # Docker unit test environment
├── Dockerfile.e2e       # Docker E2E test environment
├── e2e_test.sh          # End-to-end test script (bash/zsh/fish)
├── .golangci.yml        # Linter configuration
├── .github/
│   └── workflows/
│       ├── test.yml     # CI testing workflow (lint, unit, E2E, build)
│       └── release.yml  # Release workflow
├── TESTING.md           # Testing guide
├── CONTRIBUTING.md      # This file
├── SECURITY.md          # Security policy
└── README.md            # Project overview
```

## Areas for Contribution

### High Priority

- Additional protection patterns for dangerous paths
- Performance optimizations
- Cross-platform compatibility improvements
- Documentation improvements

### Good First Issues

- Add more test cases
- Improve error messages
- Add code comments
- Fix typos in documentation

### Advanced

- Implement configurable rm path (see TODO in README)
- Create virtualized test environment improvements
- Add support for more complex glob patterns
- Implement logging/audit trail

## Questions or Problems?

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Provide clear reproduction steps for bugs
- Include environment details (OS, Go version, etc.)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in release notes and the project README.

Thank you for helping make `dontrm` safer and better!
