//go:build linux

package singbox

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func externalProcessStatus(ctx context.Context, cfg ProcessConfig) (externalProcess, bool) {
	if strings.TrimSpace(cfg.Binary) == "" || strings.TrimSpace(cfg.ConfigPath) == "" {
		return externalProcess{}, false
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return externalProcess{}, false
	}
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return externalProcess{}, false
		default:
		}
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid == os.Getpid() {
			continue
		}
		cmdline, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
		if err != nil || len(cmdline) == 0 {
			continue
		}
		args := splitProcCmdline(cmdline)
		if matchesSingBoxRun(args, cfg) {
			return externalProcess{PID: pid}, true
		}
	}
	return externalProcess{}, false
}

func splitProcCmdline(data []byte) []string {
	parts := bytes.Split(bytes.TrimRight(data, "\x00"), []byte{0})
	args := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) > 0 {
			args = append(args, string(part))
		}
	}
	return args
}

func matchesSingBoxRun(args []string, cfg ProcessConfig) bool {
	if len(args) == 0 || !sameExecutable(args[0], cfg.Binary) {
		return false
	}
	hasRun := false
	for _, arg := range args[1:] {
		if arg == "run" {
			hasRun = true
			break
		}
	}
	return hasRun && hasConfigArg(args, cfg.ConfigPath)
}

func sameExecutable(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return false
	}
	if actual == expected || filepath.Clean(actual) == filepath.Clean(expected) {
		return true
	}
	return filepath.Base(actual) == filepath.Base(expected)
}

func hasConfigArg(args []string, expected string) bool {
	for i, arg := range args {
		switch {
		case arg == "-c" || arg == "--config":
			if i+1 < len(args) && samePath(args[i+1], expected) {
				return true
			}
		case strings.HasPrefix(arg, "-c="):
			if samePath(strings.TrimPrefix(arg, "-c="), expected) {
				return true
			}
		case strings.HasPrefix(arg, "--config="):
			if samePath(strings.TrimPrefix(arg, "--config="), expected) {
				return true
			}
		}
	}
	return false
}

func samePath(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return false
	}
	if actual == expected || filepath.Clean(actual) == filepath.Clean(expected) {
		return true
	}
	actualAbs, actualErr := filepath.Abs(actual)
	expectedAbs, expectedErr := filepath.Abs(expected)
	return actualErr == nil && expectedErr == nil && actualAbs == expectedAbs
}
