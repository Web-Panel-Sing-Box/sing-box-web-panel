package cmd_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIVersionUsesInjectedBuildVersion(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	binary := filepath.Join(t.TempDir(), "shilka")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	build := exec.Command(
		"go",
		"build",
		"-ldflags",
		"-X sing-box-web-panel/internal/version.Version=v9.9.9",
		"-o",
		binary,
		"./cmd/",
	)
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build error = %v\n%s", err, out)
	}

	run := exec.Command(binary, "version")
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("shilka version error = %v\n%s", err, out)
	}
	if got, want := strings.TrimSpace(string(out)), "shilka v9.9.9"; got != want {
		t.Fatalf("shilka version = %q, want %q", got, want)
	}
}
