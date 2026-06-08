package singbox_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sing-box-web-panel/internal/services/singbox"
)

func TestProcessManager_StartReportsImmediateExit(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryPath := writeFakeSingBox(t, dir, `#!/bin/sh
if [ "$1" = "version" ]; then echo "sing-box version 1.2.3"; exit 0; fi
echo "clash-api: bind 127.0.0.1:9090: address already in use" >&2
exit 1
`)

	pm := singbox.NewProcessManager(singbox.ProcessConfig{
		Mode:       "subprocess",
		Binary:     binaryPath,
		ConfigPath: cfgPath,
		WorkingDir: dir,
	}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))

	err := pm.Start(context.Background())
	if err == nil {
		t.Fatal("Start should fail when sing-box exits immediately")
	}
	if !strings.Contains(err.Error(), "address already in use") {
		t.Fatalf("Start error = %q, want bind detail", err.Error())
	}

	st, err := pm.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Running {
		t.Fatal("core should not be running after immediate exit")
	}
	if !strings.Contains(st.LastError, "address already in use") {
		t.Fatalf("LastError = %q, want bind detail", st.LastError)
	}
}

func TestProcessManager_StartLongRunningFakeCore(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryPath := writeFakeSingBox(t, dir, `#!/bin/sh
if [ "$1" = "version" ]; then echo "sing-box version 1.2.3"; exit 0; fi
trap 'exit 0' TERM
while true; do sleep 1; done
`)

	pm := singbox.NewProcessManager(singbox.ProcessConfig{
		Mode:       "subprocess",
		Binary:     binaryPath,
		ConfigPath: cfgPath,
		WorkingDir: dir,
	}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))

	if err := pm.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer pm.Stop(context.Background())

	st, err := pm.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !st.Running {
		t.Fatal("core should be running")
	}
	if st.PID == 0 {
		t.Fatal("PID should be non-zero")
	}
	if st.LastError != "" {
		t.Fatalf("LastError = %q, want empty", st.LastError)
	}
}

func writeFakeSingBox(t *testing.T, dir, script string) string {
	t.Helper()
	path := filepath.Join(dir, "sing-box")
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}
