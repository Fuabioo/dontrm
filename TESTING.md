# Testing Guide

## Critical Safety Requirement

**TESTS MUST ONLY RUN IN DOCKER CONTAINERS**

This project deals with dangerous file operations (`rm` commands). To prevent accidental file deletion on your host machine, all tests are designed to run exclusively in Docker containers.

## Safety Mechanisms

We employ multiple layers of defense to ensure tests never run on your host machine:

### 1. Control File Detection
The Docker test container creates a special control file at `/tmp/.docker-test-safe-env`. Tests check for this file at startup and immediately fail if it's not found:

```go
// Tests will panic if control file is missing
func requireDockerEnv() {
    if _, err := os.Stat(dockerTestControlFile); os.IsNotExist(err) {
        panic("FATAL: Tests MUST run in Docker container for safety!")
    }
}
```

### 2. No Local Test Commands
The `justfile` deliberately does NOT include a local test command. The only way to run tests is through Docker.

### 3. Smart Docker Image Caching
The test infrastructure uses intelligent caching to avoid unnecessary rebuilds while maintaining safety:
- Calculates SHA256 hash of `Dockerfile.test`
- Stores hash in `.dockertest.hash` (gitignored)
- Only rebuilds when Dockerfile changes

## Running Tests

### Prerequisites
- Docker installed and running
- `just` command runner (optional, but recommended)

### Quick Start

```bash
# Run all tests
just test

# Run tests with coverage report
just coverage

# Force rebuild of test image
just rebuild-test-image
```

### Without Just

If you don't have `just` installed:

```bash
# Build test image
docker build -f Dockerfile.test -t dontrm-test:latest .

# Run tests
docker run --rm -v $(pwd):/app -w /app dontrm-test:latest go test -v ./...

# Run with coverage
docker run --rm -v $(pwd):/app -w /app dontrm-test:latest go test -v -coverprofile=coverage.out ./...
```

## Coverage Requirements

This project enforces **85% code coverage** as a minimum threshold (accounting for the untestable main() wrapper function). The `just coverage` command will:
1. Run all tests with coverage profiling
2. Generate a coverage report
3. Check if coverage meets the 85% threshold
4. Fail if coverage is below threshold

```bash
$ just coverage
Running tests with coverage in Docker...
...
✅ Coverage is 87.3% - meets required 85%
```

## End-to-End Testing

In addition to unit tests, this project includes comprehensive **End-to-End (E2E) tests** that test the actual compiled binary as users would use it in production.

### Running E2E Tests

```bash
# Run E2E tests only
just e2e

# Run both unit tests and E2E tests
just test-all

# Force rebuild E2E Docker image
just rebuild-e2e-image
```

### Key Differences: Unit Tests vs E2E Tests

| Aspect | Unit Tests | E2E Tests |
|--------|-----------|-----------|
| **What's tested** | Go functions and logic | Actual compiled binary |
| **Environment** | Go test framework | Real Linux (Ubuntu) with bash/zsh/fish |
| **Test language** | Go | Bash scripts |
| **Control file** | `/tmp/.docker-test-safe-env` | `/tmp/.docker-e2e-safe-env` |
| **Dockerfile** | `Dockerfile.test` | `Dockerfile.e2e` (multi-stage) |
| **Purpose** | Validate logic correctness | Validate real-world usage |

### E2E Test Coverage

The E2E test suite (`e2e_test.sh`) validates:

- ✅ **Version command** - `dontrm version` works correctly
- ✅ **Dangerous path blocking** - Blocks `/`, `/etc`, `/usr/bin/*`, etc.
- ✅ **Safe deletions** - Actually creates and deletes files in test directories
- ✅ **DRY_RUN mode** - Verifies files are NOT deleted with `DRY_RUN=1`
- ✅ **sudo usage** - Tests `sudo dontrm` blocks dangerous paths
- ✅ **Shell compatibility** - Works in bash, zsh, and fish
- ✅ **Exit codes** - Correct exit codes (0 for success, 1 for blocked)
- ✅ **Error messages** - Proper error output to stderr
- ✅ **Flag parsing** - Tests `-rf`, `--no-preserve-root`, `--`, etc.

### E2E Safety Mechanisms

The E2E tests employ the same defense-in-depth safety approach:

1. **Control File Check**: First thing `e2e_test.sh` does is check for `/tmp/.docker-e2e-safe-env`
2. **Multi-stage Build**: Binary built in stage 1, tested in stage 2 - never touches host
3. **Docker-Only Execution**: No local e2e command - only runs via Docker
4. **Isolated Test Directory**: All file operations in `/tmp/dontrm-e2e-test-*` within container
5. **Separate Namespace**: Different control file and Docker image than unit tests

### E2E Test Environment

- **Base Image**: Ubuntu latest (realistic production environment)
- **Shells Installed**: bash, zsh, fish
- **Binary**: Statically compiled from source in multi-stage build
- **sudo**: Configured for testing (passwordless within container)
- **Test User**: Non-root user with sudo privileges

### Manual E2E Testing Without Just

If you don't have `just` installed:

```bash
# Build E2E image
docker build -f Dockerfile.e2e -t dontrm-e2e:latest .

# Run E2E tests
docker run --rm dontrm-e2e:latest
```

### Example E2E Test Output

```
========================================
dontrm End-to-End Test Suite
========================================
✓ Control file verified: /tmp/.docker-e2e-safe-env
✓ Running in safe Docker environment

========================================
Version Command
========================================
✅ PASS: dontrm version displays version

========================================
Dangerous Path Protection
========================================
✅ PASS: Blocks: dontrm -rf /
✅ PASS: Blocks: dontrm -rf /etc
✅ PASS: Blocks: dontrm -rf /usr/bin
✅ PASS: Blocks: dontrm -rf /var
✅ PASS: Blocks: dontrm -rf /tmp

========================================
Safe File Operations
========================================
✅ PASS: Safe file deletion (with DRY_RUN): file not deleted
✅ PASS: Safe file deletion: file deleted successfully
✅ PASS: Directory deletion with -rf: directory removed

========================================
Test Summary
========================================
Total Tests: 25
Passed: 25
Failed: 0

✅ All tests passed!
```

## What Happens If You Try to Run Tests Locally?

### Unit Tests
If you accidentally run `go test` on your host machine, the tests will immediately panic with:

```
FATAL: Tests MUST run in Docker container for safety!
The control file /tmp/.docker-test-safe-env was not found.
Use 'just test' to run tests safely in Docker.
```

### E2E Tests
If you try to run `./e2e_test.sh` on your host machine, it will immediately exit with:

```
==========================================
FATAL: E2E tests MUST run in Docker!
==========================================
The control file /tmp/.docker-e2e-safe-env was not found.
This safety mechanism prevents accidental execution on your host PC.
Use 'just e2e' to run tests safely in Docker.
```

These safety mechanisms prevent any potentially dangerous operations from executing on your system.

## Test Structure

### Current Unit Test Suites (main_test.go)

- `TestCheckArgsTopLevelPaths` - Tests top-level system path blocking
- `TestCheckArgsFilenamesWithDashes` - Tests double-dash and special filenames
- `TestCheckArgsRelativeAndSafePaths` - Tests safe path handling
- `TestCheckArgsEmptyAndFlags` - Tests edge cases with empty args and flags
- `TestIsGlob` - Tests glob pattern detection
- `TestIsTopLevelSystemPath` - Tests system path identification
- `TestSanitize` - Tests argument sorting and sanitization
- `TestEchoGlob` - Tests glob expansion
- `TestEvaluatePotentiallyDestructiveActions` - Tests glob-based dangerous pattern detection
- `TestRun` - Tests the main run() function
- `TestRunWithDifferentDryRunValues` - Tests DRY_RUN environment variable
- `TestDoubleDashStopParsingOptions` - Tests -- option parsing

### Current E2E Test Suites (e2e_test.sh)

- `test_version` - Tests version command
- `test_dangerous_paths` - Tests blocking of dangerous paths
- `test_safe_deletions` - Tests actual file and directory deletions
- `test_dry_run` - Tests DRY_RUN mode behavior
- `test_with_sudo` - Tests sudo usage patterns
- `test_shells` - Tests bash, zsh, fish compatibility
- `test_exit_codes` - Tests exit code correctness
- `test_flags` - Tests flag parsing
- `test_error_messages` - Tests error message accuracy

### Adding New Tests

**Unit Tests:**
1. All new tests automatically inherit the Docker-only requirement via `TestMain`
2. Tests should focus on logic validation, not actual file operations
3. Use temporary files for testing (automatically cleaned up)
4. Ensure new code maintains 85%+ coverage

**E2E Tests:**
1. Add new test functions to `e2e_test.sh`
2. Always use `/tmp/dontrm-e2e-test-*` directories for file operations
3. Use `pass()` and `fail()` helper functions for consistent output
4. Test real binary behavior, not Go functions
5. Add `cleanup_test` at the start and end of your test function

## Continuous Integration

GitHub Actions automatically runs tests in Docker on every push and pull request:

**Lint Job:**
- Runs go fmt
- Runs golangci-lint with 25+ linters

**Unit Test Job:**
- Builds fresh Docker test image
- Runs all unit tests with race detector
- Verifies coverage meets 85% threshold
- Uploads coverage to Codecov

**E2E Test Job:**
- Builds multi-stage E2E Docker image
- Runs comprehensive E2E test suite
- Tests actual binary in realistic environment
- Validates bash/zsh/fish compatibility

**Build Job:**
- Builds the binary
- Verifies binary executes correctly

All jobs must pass for CI to succeed. See `.github/workflows/test.yml` for details.

## Troubleshooting

### Docker Image Won't Build

```bash
# Check Docker is running
docker ps

# Try force rebuilding unit test image
just rebuild-test-image

# Try force rebuilding E2E test image
just rebuild-e2e-image
```

### "Image not found" Error

Test images are built automatically on first run. If you see this error:

```bash
# For unit tests
just rebuild-test-image

# For E2E tests
just rebuild-e2e-image
```

### Coverage Below 85%

If coverage drops below 85%:
1. Review what code is untested: `go tool cover -html=coverage.out`
2. Add tests for uncovered code paths
3. Ensure edge cases are covered

### E2E Tests Failing

If E2E tests fail:
1. Check the output for specific test failures
2. Rebuild the E2E image: `just rebuild-e2e-image`
3. Verify Docker has enough resources (memory/disk)
4. Check if shells (bash/zsh/fish) are properly installed in the image

## Philosophy

This project demonstrates defense-in-depth for dangerous operations:
- **Prevention**: Tests can't run locally by design
- **Detection**: Control file must exist
- **Validation**: Multiple test layers ensure protection works
- **Isolation**: Docker provides complete filesystem isolation

Never compromise on test safety for convenience.
