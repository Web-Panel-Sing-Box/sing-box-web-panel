package logbuf_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sing-box-web-panel/internal/services/logbuf"
)

func TestTailFileSeedsAndFollowsCoreLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sing-box.log")
	if err := os.WriteFile(path, []byte("INFO started\nERROR failed once\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	buf := logbuf.New(10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go buf.TailFile(ctx, path, logbuf.SourceCore, 10*time.Millisecond, nil)

	waitForLog(t, buf, "failed once")
	if err := appendFile(path, "WARN reconnecting\n"); err != nil {
		t.Fatal(err)
	}
	waitForLog(t, buf, "reconnecting")

	got := buf.Recent(10, "warn", "core", "reconnecting")
	if len(got) != 1 {
		t.Fatalf("warn core entries = %d, want 1", len(got))
	}
}

func waitForLog(t *testing.T, buf *logbuf.Buffer, query string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got := buf.Recent(10, "all", "core", query); len(got) > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("log containing %q was not tailed", query)
}

func appendFile(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}
