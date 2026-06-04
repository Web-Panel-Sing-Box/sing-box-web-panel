package logbuf_test

import (
	"testing"

	"sing-box-web-panel/internal/services/logbuf"
)

func TestBufferRecentFiltersBySourceAndQuery(t *testing.T) {
	buf := logbuf.New(10)
	buf.AppendEntry(logbuf.Entry{Level: "info", Source: logbuf.SourcePanel, Message: "server started"})
	buf.AppendEntry(logbuf.Entry{Level: "error", Source: logbuf.SourceFrontend, Message: "render failed", Fields: map[string]string{"component": "Dashboard"}})

	got := buf.Recent(10, "error", "frontend", "dashboard")
	if len(got) != 1 {
		t.Fatalf("entries = %d, want 1", len(got))
	}
	if got[0].Source != logbuf.SourceFrontend {
		t.Fatalf("source = %q, want frontend", got[0].Source)
	}
}

func TestBufferRedactsSensitiveFields(t *testing.T) {
	buf := logbuf.New(10)
	buf.AppendEntry(logbuf.Entry{
		Level:   "warn",
		Source:  logbuf.SourcePanel,
		Message: "auth failed",
		Fields: map[string]string{
			"admin_password": "secret",
			"request_token":  "token",
			"path":           "/api/auth/login",
		},
	})

	got := buf.Recent(10, "all", "all", "")
	if len(got) != 1 {
		t.Fatalf("entries = %d, want 1", len(got))
	}
	if got[0].Fields["admin_password"] != "[redacted]" {
		t.Fatalf("password field = %q, want redacted", got[0].Fields["admin_password"])
	}
	if got[0].Fields["request_token"] != "[redacted]" {
		t.Fatalf("token field = %q, want redacted", got[0].Fields["request_token"])
	}
	if got[0].Fields["path"] != "/api/auth/login" {
		t.Fatalf("path field = %q", got[0].Fields["path"])
	}
}
