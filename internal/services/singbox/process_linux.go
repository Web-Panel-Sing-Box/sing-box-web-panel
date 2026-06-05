//go:build linux

package singbox

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
			return externalProcess{PID: pid, Uptime: processUptime(entry.Name())}, true
		}
	}
	return externalProcess{}, false
}

// processUptime returns how long the process has been running, derived from
// /proc/<pid>/stat field 22 (starttime in clock ticks since boot) measured
// against /proc/uptime. Any parse failure yields 0 so that Running/PID
// detection is never broken by an uptime hiccup.
func processUptime(pidDir string) time.Duration {
	statRaw, err := os.ReadFile(filepath.Join("/proc", pidDir, "stat"))
	if err != nil {
		return 0
	}
	// comm (field 2) is wrapped in parentheses and may contain spaces or
	// parentheses, so split after the last ')'. The first token after it is
	// state (field 3), which makes starttime (field 22) index 19.
	idx := bytes.LastIndexByte(statRaw, ')')
	if idx < 0 || idx+2 >= len(statRaw) {
		return 0
	}
	fields := strings.Fields(string(statRaw[idx+2:]))
	if len(fields) < 20 {
		return 0
	}
	startTicks, err := strconv.ParseInt(fields[19], 10, 64)
	if err != nil {
		return 0
	}
	upRaw, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	upFields := strings.Fields(string(upRaw))
	if len(upFields) == 0 {
		return 0
	}
	sysUptime, err := strconv.ParseFloat(upFields[0], 64)
	if err != nil {
		return 0
	}
	const userHZ = 100 // sysconf(_SC_CLK_TCK); 100 on Linux in practice
	startSec := float64(startTicks) / userHZ
	if d := sysUptime - startSec; d > 0 {
		return time.Duration(d * float64(time.Second))
	}
	return 0
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
	if len(args) == 0 || !sameProcessExecutable(args, cfg.Binary) {
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

func sameProcessExecutable(args []string, expected string) bool {
	if sameExecutable(args[0], expected) {
		return true
	}
	// Test fixtures and some admin wrappers can launch an executable script via
	// an interpreter, making /proc cmdline look like `/bin/sh <script> ...`.
	return len(args) > 1 && sameExecutable(args[1], expected)
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
