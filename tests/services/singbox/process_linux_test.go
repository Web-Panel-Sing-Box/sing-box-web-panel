//go:build linux

package singbox_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"sing-box-web-panel/internal/services/singbox"
)

func TestSubprocessStatusDetectsExternalSingBoxRun(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(dir, "sing-box")
	script := "#!/bin/sh\nif [ \"$1\" = \"version\" ]; then echo \"sing-box version 1.2.3\"; exit 0; fi\nsleep 30\n"
	if err := os.WriteFile(binaryPath, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}

	proc := exec.Command(binaryPath, "run", "-c", configPath)
	if err := proc.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = proc.Process.Kill()
		_ = proc.Wait()
	}()

	pm := singbox.NewProcessManager(singbox.ProcessConfig{
		Mode:       "subprocess",
		Binary:     binaryPath,
		ConfigPath: configPath,
	}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st, err := pm.Status(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if st.Running && st.PID == proc.Process.Pid {
			if st.Version != "sing-box 1.2.3" {
				t.Fatalf("version = %q, want sing-box 1.2.3", st.Version)
			}
			if st.Uptime <= 0 {
				t.Fatalf("uptime = %v, want > 0 for a detected external process", st.Uptime)
			}
			if st.Uptime > time.Minute {
				t.Fatalf("uptime = %v, implausibly large for a just-started process", st.Uptime)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	st, err := pm.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Fatalf("status = %+v, want running external process pid %d", st, proc.Process.Pid)
}
