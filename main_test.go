package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsVulnerable(t *testing.T) {

	tests := []struct {
		name     string
		args     string
		expected string
	}{
		{
			name:     "classic 1",
			args:     "-rf --no-preserve-root /",
			expected: "known top level match: /",
		},
		{
			name:     "classic 2",
			args:     "-rf /",
			expected: "known top level match: /",
		},
		{
			name:     "glob not protected",
			args:     "-rf /*",
			expected: "known top level match: /bin",
		},
		{
			name:     "home directory can be deleted 😈",
			args:     "/home/user",
			expected: "",
		},
		{
			name:     "wildcard in user bin",
			args:     "/usr/bin/*",
			expected: "known top level direct child match, all contents match: /usr/bin/*",
		},
		{
			name:     "direct file in /usr/bin",
			args:     "/usr/bin/bash",
			expected: "",
		},
		{
			name:     "direct dir within /usr/bin",
			args:     "/usr/bin/go",
			expected: "",
		},
		{
			name:     "direct dir within /usr/bin with wildcard",
			args:     "/usr/bin/go/*",
			expected: "",
		},
		{
			name:     "wildcard in /etc",
			args:     "/etc/*",
			expected: "known top level direct child match, all contents match: /etc/*",
		},
		{
			name:     "wildcard in /tmp",
			args:     "/tmp/*",
			expected: "",
		},
		{
			name:     "direct /tmp",
			args:     "/tmp/",
			expected: "known top level match: /tmp",
		},
		{
			name:     "hidden files in root",
			args:     "/.*",
			expected: "",
		},
		{
			name:     "wildcard in /mnt",
			args:     "/mnt/*",
			expected: "",
		},
		{
			name:     "super wildcard!",
			args:     "/**/*",
			expected: "known top level match: /usr/bin",
		},
		{
			name:     "explicit file",
			args:     "/etc/passwd",
			expected: "",
		},
		{
			name:     "file starts with -",
			args:     "-rf -- -foo bar",
			expected: "",
		},
		{
			name:     "file starts with - at top level",
			args:     "-rf -- /-foo",
			expected: "",
		},
		{
			name:     "relative path",
			args:     "./go.mod",
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			args := []string{"run", "."}
			for _, arg := range strings.SplitN(test.args, " ", -1) {
				if strings.HasPrefix(arg, "-") {
					continue
				}
				if isGlob(arg) {
					matches, err := filepath.Glob(arg)
					if err != nil {
						t.Fatal(err)
					}
					if len(matches) > 0 {
						args = append(args, matches...)
						continue
					}
				}
				args = append(args, arg)
			}

			cmd := exec.Command("go", args...)

			var buffer bytes.Buffer

			cmd.Stdout = &buffer
			cmd.Stderr = &buffer

			_ = cmd.Run()

			if len(test.expected) == 0 && buffer.Len() != 0 {
				t.Error("- ", test.expected)
				t.Error("+ ", buffer.String())
				return
			}

			result := strings.Replace(buffer.String(), "exit status 1", "", -1)
			result = strings.TrimSpace(result)
			if test.expected != result {
				t.Error("- ", test.expected)
				t.Error("+ ", result)
			}

		})
	}
}

func TestIsGlob(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Simple wildcard", "/usr/bin/*", true},
		{"Single char wildcard", "/home/?ser", true},
		{"Character set wildcard", "/tmp/[a-z]*", true},
		{"No wildcard", "/etc/passwd", false},
		{"Plain directory", "/var/log", false},
		{"Escaped asterisk", "/home/user/\\*", true}, // TODO fix
		{"Double wildcard", "/**/*", true},
		{"Hidden glob", "/.*", true},
		{"Relative path no glob", "docs/index.html", false},
		{"Trailing wildcard only", "*", true},
		{"Wildcard and literal", "file[1-9].txt", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isGlob(tc.input)
			if result != tc.expected {
				t.Errorf("isGlob(%q) = %v; want %v", tc.input, result, tc.expected)
			}
		})
	}
}
