#!/bin/bash

# End-to-End Tests for dontrm
# Tests the actual binary as users would use it in production
#
# CRITICAL SAFETY: This script checks for a control file to ensure
# it's running in a Docker container and will NOT run on the host PC

# Note: NOT using set -e so tests can continue even if individual tests fail

# SAFETY CHECK: Verify we're in Docker container
CONTROL_FILE="/tmp/.docker-e2e-safe-env"
if [ ! -f "$CONTROL_FILE" ]; then
    echo "=========================================="
    echo "FATAL: E2E tests MUST run in Docker!"
    echo "=========================================="
    echo "The control file $CONTROL_FILE was not found."
    echo "This safety mechanism prevents accidental execution on your host PC."
    echo "Use 'just e2e' to run tests safely in Docker."
    exit 1
fi

# Test counters
PASS=0
FAIL=0
TOTAL=0

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test framework functions
pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    PASS=$((PASS+1))
    TOTAL=$((TOTAL+1))
}

fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    echo -e "${RED}   ${2}${NC}"
    FAIL=$((FAIL+1))
    TOTAL=$((TOTAL+1))
}

test_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Helper functions
create_test_file() {
    local file=$1
    mkdir -p "$(dirname "$file")"
    echo "test content" > "$file"
}

cleanup_test() {
    rm -rf /tmp/dontrm-e2e-test-* 2>/dev/null || true
}

# Test: Version command
test_version() {
    test_header "Version Command"

    local output
    output=$(dontrm version 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ] && echo "$output" | grep -q "DON'T rm!"; then
        pass "dontrm version displays version"
    else
        fail "dontrm version failed" "Exit code: $exit_code, Output: $output"
    fi
}

# Test: Dangerous path protection
test_dangerous_paths() {
    test_header "Dangerous Path Protection"

    # Test root path
    output=$(dontrm -rf / 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ] && echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "Blocks: dontrm -rf /"
    else
        fail "Failed to block root path" "Exit code: $exit_code"
    fi

    # Test /etc
    output=$(dontrm -rf /etc 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ] && echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "Blocks: dontrm -rf /etc"
    else
        fail "Failed to block /etc" "Exit code: $exit_code"
    fi

    # Test /usr/bin
    output=$(dontrm -rf /usr/bin 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ] && echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "Blocks: dontrm -rf /usr/bin"
    else
        fail "Failed to block /usr/bin" "Exit code: $exit_code"
    fi

    # Test /var
    output=$(dontrm -rf /var 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ] && echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "Blocks: dontrm -rf /var"
    else
        fail "Failed to block /var" "Exit code: $exit_code"
    fi

    # Test /tmp (should block top-level)
    output=$(dontrm -rf /tmp 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ] && echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "Blocks: dontrm -rf /tmp"
    else
        fail "Failed to block /tmp" "Exit code: $exit_code"
    fi
}

# Test: Safe file deletions
test_safe_deletions() {
    test_header "Safe File Operations"

    cleanup_test

    # Test single file deletion
    local test_file="/tmp/dontrm-e2e-test-file.txt"
    create_test_file "$test_file"

    if [ -f "$test_file" ]; then
        DRY_RUN=1 dontrm "$test_file" 2>&1
        exit_code=$?
        if [ $exit_code -eq 0 ] && [ -f "$test_file" ]; then
            pass "Safe file deletion (with DRY_RUN): file not deleted"
        else
            fail "DRY_RUN test failed" "File was deleted or wrong exit code"
        fi
        rm -f "$test_file"
    fi

    # Test actual deletion (in safe location)
    create_test_file "$test_file"
    unset DRY_RUN
    dontrm "$test_file" 2>&1
    exit_code=$?
    if [ $exit_code -eq 0 ] && [ ! -f "$test_file" ]; then
        pass "Safe file deletion: file deleted successfully"
    else
        fail "File deletion failed" "Exit code: $exit_code"
    fi

    # Test directory deletion
    local test_dir="/tmp/dontrm-e2e-test-dir"
    mkdir -p "$test_dir"
    echo "content" > "$test_dir/file.txt"

    dontrm -rf "$test_dir" 2>&1
    exit_code=$?
    if [ $exit_code -eq 0 ] && [ ! -d "$test_dir" ]; then
        pass "Directory deletion with -rf: directory removed"
    else
        fail "Directory deletion failed" "Exit code: $exit_code"
    fi

    cleanup_test
}

# Test: DRY_RUN mode
test_dry_run() {
    test_header "DRY_RUN Mode"

    cleanup_test

    # Test with DRY_RUN=1
    local test_file="/tmp/dontrm-e2e-test-dryrun.txt"
    create_test_file "$test_file"

    DRY_RUN=1 dontrm "$test_file" 2>&1
    exit_code=$?
    if [ $exit_code -eq 0 ] && [ -f "$test_file" ]; then
        pass "DRY_RUN=1: file preserved"
    else
        fail "DRY_RUN=1 failed" "File was deleted or wrong exit code"
    fi

    # Test with DRY_RUN=true
    DRY_RUN=true dontrm "$test_file" 2>&1
    exit_code=$?
    if [ $exit_code -eq 0 ] && [ -f "$test_file" ]; then
        pass "DRY_RUN=true: file preserved"
    else
        fail "DRY_RUN=true failed" "File was deleted or wrong exit code"
    fi

    # Test that dangerous paths still blocked in DRY_RUN
    output=$(DRY_RUN=1 dontrm -rf /etc 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ] && echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "DRY_RUN=1: still blocks dangerous paths"
    else
        fail "DRY_RUN=1 didn't block dangerous path" "Exit code: $exit_code"
    fi

    cleanup_test
}

# Test: sudo usage
test_with_sudo() {
    test_header "sudo Usage"

    cleanup_test

    # Test sudo with dangerous path (should still block)
    output=$(sudo dontrm -rf /etc 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ] && echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "sudo dontrm: blocks dangerous paths"
    else
        fail "sudo dontrm didn't block dangerous path" "Exit code: $exit_code"
    fi

    # Test sudo with safe path
    local test_file="/tmp/dontrm-e2e-test-sudo.txt"
    sudo bash -c "echo 'test' > $test_file"
    sudo chmod 644 "$test_file"

    if [ -f "$test_file" ]; then
        # Use sudo with -E to preserve environment variables
        DRY_RUN=1 sudo -E dontrm "$test_file" 2>&1
        exit_code=$?
        if [ $exit_code -eq 0 ] && [ -f "$test_file" ]; then
            pass "sudo dontrm with DRY_RUN: works correctly"
        else
            fail "sudo dontrm with DRY_RUN failed" "Exit code: $exit_code, File exists: $([ -f "$test_file" ] && echo 'yes' || echo 'no')"
        fi
    fi

    cleanup_test
}

# Test: Shell compatibility
test_shells() {
    test_header "Shell Compatibility"

    cleanup_test

    # Test with bash
    if command -v bash >/dev/null 2>&1; then
        output=$(bash -c 'dontrm version' 2>&1)
        if echo "$output" | grep -q "DON'T rm!"; then
            pass "Works in bash"
        else
            fail "Failed in bash" "Output: $output"
        fi
    fi

    # Test with zsh
    if command -v zsh >/dev/null 2>&1; then
        output=$(zsh -c 'dontrm version' 2>&1)
        if echo "$output" | grep -q "DON'T rm!"; then
            pass "Works in zsh"
        else
            fail "Failed in zsh" "Output: $output"
        fi
    fi

    # Test with fish
    if command -v fish >/dev/null 2>&1; then
        output=$(fish -c 'dontrm version' 2>&1)
        if echo "$output" | grep -q "DON'T rm!"; then
            pass "Works in fish"
        else
            fail "Failed in fish" "Output: $output"
        fi
    fi
}

# Test: Exit codes
test_exit_codes() {
    test_header "Exit Codes"

    cleanup_test

    # Test successful command (version)
    dontrm version >/dev/null 2>&1
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        pass "Exit code 0 for successful command"
    else
        fail "Wrong exit code for successful command" "Got: $exit_code"
    fi

    # Test blocked dangerous path
    dontrm -rf /etc >/dev/null 2>&1
    exit_code=$?
    if [ $exit_code -eq 1 ]; then
        pass "Exit code 1 for blocked dangerous path"
    else
        fail "Wrong exit code for blocked path" "Got: $exit_code"
    fi

    # Test successful DRY_RUN
    local test_file="/tmp/dontrm-e2e-test-exit.txt"
    create_test_file "$test_file"
    DRY_RUN=1 dontrm "$test_file" >/dev/null 2>&1
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        pass "Exit code 0 for DRY_RUN"
    else
        fail "Wrong exit code for DRY_RUN" "Got: $exit_code"
    fi

    cleanup_test
}

# Test: Flag parsing
test_flags() {
    test_header "Flag Parsing"

    # Test --no-preserve-root still blocked
    output=$(dontrm -rf --no-preserve-root / 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ] && echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "Blocks even with --no-preserve-root"
    else
        fail "--no-preserve-root bypass not prevented" "Exit code: $exit_code"
    fi

    # Test double dash
    output=$(dontrm -- /etc 2>&1)
    exit_code=$?
    if [ $exit_code -eq 1 ]; then
        pass "Double dash (--) parsing works"
    else
        fail "Double dash parsing failed" "Exit code: $exit_code"
    fi
}

# Test: Error messages
test_error_messages() {
    test_header "Error Messages"

    # Test error message for root path
    output=$(dontrm -rf / 2>&1)
    if echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "Error message contains 'Blocked dangerous operation'"
    else
        fail "Error message incorrect" "Got: $output"
    fi

    # Test error message for /etc
    output=$(dontrm /etc 2>&1)
    if echo "$output" | grep -q "Blocked dangerous operation"; then
        pass "Error message for /etc is correct"
    else
        fail "Error message for /etc incorrect" "Got: $output"
    fi
}

# Main test execution
main() {
    echo ""
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}dontrm End-to-End Test Suite${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${GREEN}✓ Control file verified: $CONTROL_FILE${NC}"
    echo -e "${GREEN}✓ Running in safe Docker environment${NC}"
    echo ""

    # Run all test suites
    test_version
    test_dangerous_paths
    test_safe_deletions
    test_dry_run
    test_with_sudo
    test_shells
    test_exit_codes
    test_flags
    test_error_messages

    # Final cleanup
    cleanup_test

    # Summary
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Total Tests: $TOTAL"
    echo -e "${GREEN}Passed: $PASS${NC}"
    echo -e "${RED}Failed: $FAIL${NC}"
    echo ""

    if [ $FAIL -eq 0 ]; then
        echo -e "${GREEN}✅ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some tests failed!${NC}"
        exit 1
    fi
}

# Run main
main
