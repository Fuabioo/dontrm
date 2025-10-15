package main

import (
	"errors"
	"os"
	"strings"
	"testing"
)

const dockerTestControlFile = "/tmp/.docker-test-safe-env"

// TestMain ensures we're running in a safe Docker environment.
func TestMain(m *testing.M) {
	requireDockerEnv()
	os.Exit(m.Run())
}

// requireDockerEnv checks for the Docker control file and panics if not found.
// This prevents accidental test execution on the host machine.
func requireDockerEnv() {
	if _, err := os.Stat(dockerTestControlFile); os.IsNotExist(err) {
		panic("FATAL: Tests MUST run in Docker container for safety! " +
			"The control file " + dockerTestControlFile + " was not found. " +
			"Use 'just test' to run tests safely in Docker.")
	}
}

func TestCheckArgsTopLevelPaths(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorType   error
	}{
		{
			name:        "root path",
			args:        []string{"-rf", "/"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "root path with no-preserve-root",
			args:        []string{"-rf", "--no-preserve-root", "/"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "bin directory",
			args:        []string{"/bin"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "etc directory",
			args:        []string{"/etc"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "usr directory",
			args:        []string{"/usr"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "usr/bin subdirectory",
			args:        []string{"/usr/bin"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "var directory",
			args:        []string{"/var"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "tmp directory",
			args:        []string{"/tmp"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "tmp with trailing slash",
			args:        []string{"/tmp/"},
			expectError: true,
			errorType:   ErrTopLevelPath,
		},
		{
			name:        "home directory (allowed)",
			args:        []string{"/home/user"},
			expectError: false,
		},
		{
			name:        "specific file in etc (allowed)",
			args:        []string{"/etc/passwd"},
			expectError: false,
		},
		{
			name:        "specific file in usr/bin (allowed)",
			args:        []string{"/usr/bin/bash"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkArgs(test.args)

			if test.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !errors.Is(err, test.errorType) {
					t.Errorf("Expected error type %v, got %v", test.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestCheckArgsFilenamesWithDashes(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "filename starts with dash after --",
			args:        []string{"-rf", "--", "-foo", "bar"},
			expectError: false,
		},
		{
			name:        "filename starts with dash at root after --",
			args:        []string{"-rf", "--", "/-foo"},
			expectError: false,
		},
		{
			name:        "only flags",
			args:        []string{"-rf", "-v"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkArgs(test.args)
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestCheckArgsRelativeAndSafePaths(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "relative path",
			args:        []string{"./go.mod"},
			expectError: false,
		},
		{
			name:        "relative directory",
			args:        []string{"./somedir"},
			expectError: false,
		},
		{
			name:        "current directory files",
			args:        []string{"file1.txt", "file2.txt"},
			expectError: false,
		},
		{
			name:        "home user file",
			args:        []string{"/home/user/file.txt"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkArgs(test.args)
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestCheckArgsEmptyAndFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "empty args",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "only flags",
			args:        []string{"-rf", "-v", "-i"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkArgs(test.args)
			// Empty args and flags-only should be safe
			// The actual rm command will handle these cases
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestIsGlob(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Simple wildcard", "/usr/bin/*", true},
		{"Single char wildcard", "/home/?ser", true},
		{"Character set wildcard", "/tmp/[a-z]*", true},
		{"No wildcard", "/etc/passwd", false},
		{"Plain directory", "/var/log", false},
		{"Escaped asterisk", "/home/user/\\*", true}, // Contains *, even if escaped
		{"Double wildcard", "/**/*", true},
		{"Hidden glob", "/.*", true},
		{"Relative path no glob", "docs/index.html", false},
		{"Trailing wildcard only", "*", true},
		{"Wildcard and literal", "file[1-9].txt", true},
		{"Question mark", "file?.txt", true},
		{"Plain filename", "file.txt", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isGlob(test.input)
			if result != test.expected {
				t.Errorf("isGlob(%q) = %v; want %v", test.input, result, test.expected)
			}
		})
	}
}

func TestIsTopLevelSystemPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectMatch bool
		expectValue string
	}{
		{"root", "/", true, "/"},
		{"bin", "/bin", true, "/bin"},
		{"boot", "/boot", true, "/boot"},
		{"dev", "/dev", true, "/dev"},
		{"etc", "/etc", true, "/etc"},
		{"lib", "/lib", true, "/lib"},
		{"lib64", "/lib64", true, "/lib64"},
		{"opt", "/opt", true, "/opt"},
		{"proc", "/proc", true, "/proc"},
		{"root dir", "/root", true, "/root"},
		{"run", "/run", true, "/run"},
		{"sbin", "/sbin", true, "/sbin"},
		{"srv", "/srv", true, "/srv"},
		{"sys", "/sys", true, "/sys"},
		{"tmp", "/tmp", true, "/tmp"},
		{"usr", "/usr", true, "/usr"},
		{"usr/bin", "/usr/bin", true, "/usr/bin"},
		{"usr/sbin", "/usr/sbin", true, "/usr/sbin"},
		{"var", "/var", true, "/var"},
		{"home top level", "/home", true, "/home"},
		{"media", "/media", true, "/media"},
		{"mnt", "/mnt", true, "/mnt"},
		{"trailing slash", "/etc/", true, "/etc"},
		{"not system path", "/home/user/documents", false, ""},
		{"relative path", "./somedir", false, ""},
		{"non-system absolute", "/custom/path", false, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, match := isTopLevelSystemPath(test.path)
			if match != test.expectMatch {
				t.Errorf("Expected match=%v, got match=%v for path %q", test.expectMatch, match, test.path)
			}
			if value != test.expectValue {
				t.Errorf("Expected value=%q, got value=%q", test.expectValue, value)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"single value", []string{"foo"}, "foo"},
		{"multiple values", []string{"foo", "bar", "baz"}, "bar baz foo"},
		{"already sorted", []string{"a", "b", "c"}, "a b c"},
		{"reverse sorted", []string{"c", "b", "a"}, "a b c"},
		{"empty slice", []string{}, ""},
		{"duplicates", []string{"x", "y", "x"}, "x x y"},
		{"single item", []string{"test"}, "test"},
		{"with paths", []string{"/var", "/etc", "/bin"}, "/bin /etc /var"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := sanitize(test.input)
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestEchoGlob(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		expectError bool
		expectSelf  bool // Should return pattern itself when no glob
	}{
		{"no glob", "/etc/passwd", false, true},
		{"simple glob", "/tmp/*", false, false},
		{"invalid glob", "[", true, false},
		{"plain filename", "file.txt", false, true},
		{"malformed bracket", "[unclosed", true, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := echoGlob(test.pattern)
			if test.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if test.expectSelf && (len(result) != 1 || result[0] != test.pattern) {
					t.Errorf("Expected [%q], got %v", test.pattern, result)
				}
			}
		})
	}
}

func TestEvaluatePotentiallyDestructiveActions(t *testing.T) {
	// This function expands globs and checks for dangerous patterns
	// Note: empty strings may match empty directory expansions, but checkArgs
	// filters them out before calling this function, so it's safe
	tests := []struct {
		name        string
		tail        string
		shouldMatch bool
	}{
		{"single file", "file.txt", false},
		{"multiple files", "file1.txt file2.txt", false},
		{"non-existent path", "/nonexistent/path/file.txt", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pattern, matched := evaluatePotentiallyDestructiveActions(test.tail)
			if matched != test.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v for tail %q (pattern: %q)",
					test.shouldMatch, matched, test.tail, pattern)
			}
		})
	}
}

// TestRun tests the main run function.
func TestRun(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		dryRun       string
		expectedCode int
		checkOutput  bool
		expectOutput string
	}{
		{
			name:         "version command",
			args:         []string{"version"},
			expectedCode: 0,
			checkOutput:  true,
			expectOutput: "DON'T rm!",
		},
		{
			name:         "dangerous path blocked",
			args:         []string{"-rf", "/etc"},
			expectedCode: 1,
			checkOutput:  false,
		},
		{
			name:         "dry run with safe path",
			args:         []string{"/home/user/file.txt"},
			dryRun:       "1",
			expectedCode: 0,
			checkOutput:  false,
		},
		{
			name:         "dry run with dangerous path still blocked",
			args:         []string{"/etc"},
			dryRun:       "1",
			expectedCode: 1,
			checkOutput:  false,
		},
		{
			name:         "empty args in dry run",
			args:         []string{},
			expectedCode: 0, // Empty args pass validation, dry run returns 0
			checkOutput:  false,
		},
		{
			name:         "only flags",
			args:         []string{"-rf"},
			expectedCode: 0, // Passes validation, rm handles it
			checkOutput:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create temporary files for stdout/stderr
			tmpStdout, err := os.CreateTemp("", "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Remove(tmpStdout.Name()) }() //nolint:errcheck // cleanup in tests
			defer func() { _ = tmpStdout.Close() }()           //nolint:errcheck // cleanup in tests

			tmpStderr, err := os.CreateTemp("", "stderr")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Remove(tmpStderr.Name()) }() //nolint:errcheck // cleanup in tests
			defer func() { _ = tmpStderr.Close() }()           //nolint:errcheck // cleanup in tests

			// Set DRY_RUN if specified
			if test.dryRun != "" {
				t.Setenv("DRY_RUN", test.dryRun)
			} else {
				t.Setenv("DRY_RUN", "1") // Always use dry run in tests for safety
			}

			// Run the function
			exitCode := run(test.args, tmpStdout, tmpStderr)

			// Check exit code
			if exitCode != test.expectedCode {
				t.Errorf("Expected exit code %d, got %d", test.expectedCode, exitCode)
			}

			// Check output if requested
			if test.checkOutput {
				_, _ = tmpStdout.Seek(0, 0) //nolint:errcheck // test helper
				output := make([]byte, 1000)
				n, _ := tmpStdout.Read(output) //nolint:errcheck // test helper
				outputStr := string(output[:n])

				if !strings.Contains(outputStr, test.expectOutput) {
					t.Errorf("Expected output to contain %q, got %q", test.expectOutput, outputStr)
				}
			}
		})
	}
}

// TestRunWithDifferentDryRunValues tests DRY_RUN environment variable handling.
func TestRunWithDifferentDryRunValues(t *testing.T) {
	tests := []struct {
		name         string
		dryRunValue  string
		args         []string
		expectedCode int
	}{
		{
			name:         "DRY_RUN=1 with safe path",
			dryRunValue:  "1",
			args:         []string{"/home/user/file.txt"},
			expectedCode: 0,
		},
		{
			name:         "DRY_RUN=true with safe path",
			dryRunValue:  "true",
			args:         []string{"/home/user/file.txt"},
			expectedCode: 0,
		},
		{
			name:         "DRY_RUN=false with safe path",
			dryRunValue:  "false",
			args:         []string{"/nonexistent/file.txt"},
			expectedCode: 1, // Will fail because file doesn't exist, but that's ok
		},
		{
			name:         "DRY_RUN empty with dangerous path",
			dryRunValue:  "",
			args:         []string{"/etc"},
			expectedCode: 1, // Blocked by safety check
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpStdout, err := os.CreateTemp("", "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Remove(tmpStdout.Name()) }() //nolint:errcheck // cleanup in tests
			defer func() { _ = tmpStdout.Close() }()           //nolint:errcheck // cleanup in tests

			tmpStderr, err := os.CreateTemp("", "stderr")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Remove(tmpStderr.Name()) }() //nolint:errcheck // cleanup in tests
			defer func() { _ = tmpStderr.Close() }()           //nolint:errcheck // cleanup in tests

			t.Setenv("DRY_RUN", test.dryRunValue)

			exitCode := run(test.args, tmpStdout, tmpStderr)

			if exitCode != test.expectedCode {
				t.Errorf("Expected exit code %d, got %d", test.expectedCode, exitCode)
			}
		})
	}
}

// TestDoubleDashStopParsingOptions tests double dash handling.
func TestDoubleDashStopParsingOptions(t *testing.T) {
	// Test that -- properly stops option parsing
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "-- followed by /etc",
			args:        []string{"--", "/etc"},
			expectError: true, // /etc is a top-level path
		},
		{
			name:        "-- followed by safe path",
			args:        []string{"--", "/home/user/file"},
			expectError: false,
		},
		{
			name:        "-- followed by flag-like filename",
			args:        []string{"--", "-filename"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkArgs(test.args)
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
