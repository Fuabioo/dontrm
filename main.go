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

// Define a set of known top-level system paths
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
	ErrTopLevelPath             = errors.New("known top level match")
	ErrTopLevelChildAllContents = errors.New("known top level direct child match, all contents match")
)

var version = "dev"

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "version" {
			println(lipgloss.NewStyle().Bold(true).Render("DON'T rm!"), version)
			return
		}
	}

	dryRun := os.Getenv("DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "1"
	err := checkArgs(args)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	if dryRun {
		os.Exit(0)
	}
	cmd := exec.Command("/usr/bin/rm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
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

// isGlob returns true if the path contains typical globbing characters
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

	// any potentially destructive path e.g. /usr/bin/*
	// - add a 🤡 each time you've fallen for that specific one -
	// 🤡🤡
	evaluated, match := evaluatePotentiallyDestructiveActions(sanitize(tail))
	if match {
		return fmt.Errorf("%w: %s", ErrTopLevelChildAllContents, evaluated)
	}

	return nil
}
