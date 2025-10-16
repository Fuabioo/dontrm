// Package main implements dontrm, a safe wrapper around the rm command that prevents catastrophic system deletions.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// systemPaths defines a set of known top-level system paths that should be protected.
var systemPaths = map[string]string{
	"/":         "/",
	"/bin":      "/bin",
	"/boot":     "/boot",
	"/dev":      "/dev",
	"/etc":      "/etc",
	"/home":     "/home",
	"/lib":      "/lib",
	"/lib64":    "/lib64",
	"/media":    "/media",
	"/mnt":      "/mnt",
	"/opt":      "/opt",
	"/proc":     "/proc",
	"/root":     "/root",
	"/run":      "/run",
	"/sbin":     "/sbin",
	"/srv":      "/srv",
	"/sys":      "/sys",
	"/tmp":      "/tmp",
	"/usr":      "/usr",
	"/usr/bin":  "/usr/bin",
	"/usr/sbin": "/usr/sbin",
	"/var":      "/var",
}

var (
	// ErrTopLevelPath indicates that a top-level system path was matched.
	ErrTopLevelPath = errors.New("⛔ Blocked dangerous operation: Cannot delete system directory")
	// ErrTopLevelChildAllContents indicates that all contents of a top-level directory were matched.
	ErrTopLevelChildAllContents = errors.New("⛔ Blocked dangerous operation: Cannot delete all contents of system directory")
)

var version = "dev"

func main() {
	exitCode := run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(exitCode)
}

// run contains the main application logic and returns an exit code.
// This function is extracted to be testable without side effects.
func run(args []string, stdout, stderr *os.File) int {
	// Handle version command
	if len(args) > 0 && args[0] == "version" {
		_, _ = fmt.Fprintln(stdout, lipgloss.NewStyle().Bold(true).Render("DON'T rm!"), version)
		return 0
	}

	// Check if dry run mode is enabled
	dryRun := os.Getenv("DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "1"

	// Validate arguments for safety
	if err := checkArgs(args); err != nil {
		_, _ = fmt.Fprintln(stderr, err.Error())
		return 1
	}

	// In dry run mode, exit successfully without executing rm
	if dryRun {
		return 0
	}

	// Execute the actual rm command
	cmd := exec.Command("/usr/bin/rm", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return 1
	}

	return 0
}

func isTopLevelSystemPath(path string) (string, bool) {
	cleanPath := filepath.Clean(path)
	value, ok := systemPaths[cleanPath]
	return value, ok
}

func sanitize(values []string) string {
	// Sort the values for deterministic output
	sort.Strings(values)
	return strings.Join(values, " ")
}

// isGlob returns true if the path contains typical globbing characters.
func isGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func echoGlob(pattern string) ([]string, error) {
	if !isGlob(pattern) {
		return []string{pattern}, nil
	}

	// Use filepath.Glob to expand the pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func evaluatePotentiallyDestructiveActions(tail string) (string, bool) {
	for sysPath := range systemPaths {
		// evaluate sysPath/*
		evaluated := filepath.Join(sysPath, "*")
		sysPathTail, err := echoGlob(evaluated)
		if err != nil {
			log.Println(err)
			continue
		}
		if sanitize(sysPathTail) == tail {
			return evaluated, true
		}
	}

	return "", false
}

func checkArgs(args []string) error {
	tail := make([]string, 0, len(args))
	stopParsingOptions := false
	for _, arg := range args {
		if arg == "--" {
			stopParsingOptions = true
		}

		// ignore flags
		if !stopParsingOptions && strings.HasPrefix(arg, "-") {
			continue
		}

		// any known top level e.g. /usr/bin or /usr/bin/
		evaluated, match := isTopLevelSystemPath(arg)
		if match {
			return fmt.Errorf("%w: %s", ErrTopLevelPath, evaluated)
		}

		tail = append(tail, arg)
	}

	// If tail is empty (no files specified), skip destructive action check
	// The actual rm command will handle empty args appropriately
	if len(tail) == 0 {
		return nil
	}

	// any potentially destructive path e.g. /usr/bin/*
	// - add a 🤡 each time you've fallen for that specific one -
	// 🤡🤡
	evaluated, match := evaluatePotentiallyDestructiveActions(sanitize(tail))
	if match {
		return fmt.Errorf("%w: %s", ErrTopLevelChildAllContents, evaluated)
	}

	return nil
}
