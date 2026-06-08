package singbox

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Status is a snapshot of the sing-box process state for the dashboard.
type Status struct {
	Running bool
	PID     int
	Version string
	Uptime  time.Duration
	// LastError is a short diagnostic for the most recent failed start or
	// unexpected exit. It is intentionally human-readable for API/CLI display.
	LastError string
}

type externalProcess struct {
	PID    int
	Uptime time.Duration
}

// ProcessManager controls the sing-box lifecycle. Reload sends SIGHUP; note
// that sing-box reload re-reads the config but resets active connections.
type ProcessManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Restart(ctx context.Context) error
	Reload(ctx context.Context) error
	Status(ctx context.Context) (Status, error)
}

// ProcessConfig configures the process manager.
type ProcessConfig struct {
	Mode         string // auto | systemd | subprocess
	Binary       string
	ConfigPath   string
	WorkingDir   string
	ServiceName  string
	RestartDelay time.Duration
}

// NewProcessManager picks a systemd or subprocess manager. In "auto" mode it
// uses systemd when the unit is installed (Linux), otherwise a subprocess
// (typical for local development on macOS).
func NewProcessManager(cfg ProcessConfig, logSink io.Writer, log *slog.Logger) ProcessManager {
	if logSink == nil {
		logSink = io.Discard
	}
	switch cfg.Mode {
	case "systemd":
		return &systemdManager{service: cfg.ServiceName, binary: cfg.Binary}
	case "subprocess":
		return newSubprocessManager(cfg, logSink, log)
	default:
		if systemdAvailable(cfg.ServiceName) {
			return &systemdManager{service: cfg.ServiceName, binary: cfg.Binary}
		}
		return newSubprocessManager(cfg, logSink, log)
	}
}

// coreVersion runs `sing-box version` and extracts the version string.
func coreVersion(ctx context.Context, binary string) string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, binary, "version").Output()
	if err != nil {
		return ""
	}
	// "sing-box version 1.11.0 ..." -> "sing-box 1.11.0"
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "version" && i+1 < len(fields) {
			return "sing-box " + fields[i+1]
		}
	}
	return strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
}

// --- systemd ---

func systemdAvailable(service string) bool {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return false
	}
	if service == "" {
		return false
	}
	// `systemctl cat <unit>` exits 0 only if the unit file exists.
	return exec.Command("systemctl", "cat", service).Run() == nil
}

type systemdManager struct {
	service string
	binary  string

	mu        sync.Mutex
	lastError string
}

func (m *systemdManager) systemctl(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "systemctl", append(args, m.service)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}

func (m *systemdManager) Start(ctx context.Context) error {
	if err := m.systemctl(ctx, "start"); err != nil {
		return m.recordSystemdError(ctx, err)
	}
	return m.waitActive(ctx, "start")
}

func (m *systemdManager) Stop(ctx context.Context) error {
	if err := m.systemctl(ctx, "stop"); err != nil {
		return m.recordSystemdError(ctx, err)
	}
	m.setLastError("")
	return nil
}

func (m *systemdManager) Restart(ctx context.Context) error {
	if err := m.systemctl(ctx, "restart"); err != nil {
		return m.recordSystemdError(ctx, err)
	}
	return m.waitActive(ctx, "restart")
}

func (m *systemdManager) Reload(ctx context.Context) error {
	if err := m.systemctl(ctx, "reload"); err != nil {
		return m.recordSystemdError(ctx, err)
	}
	return m.waitActive(ctx, "reload")
}

func (m *systemdManager) Status(ctx context.Context) (Status, error) {
	st := Status{Version: coreVersion(ctx, m.binary)}

	active, _ := exec.CommandContext(ctx, "systemctl", "is-active", m.service).Output()
	st.Running = strings.TrimSpace(string(active)) == "active"

	if props, err := exec.CommandContext(ctx, "systemctl", "show", m.service,
		"--property=MainPID,ActiveEnterTimestamp").Output(); err == nil {
		for _, line := range strings.Split(string(props), "\n") {
			key, val, ok := strings.Cut(strings.TrimSpace(line), "=")
			if !ok {
				continue
			}
			switch key {
			case "MainPID":
				st.PID, _ = strconv.Atoi(val)
			case "ActiveEnterTimestamp":
				if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", val); err == nil {
					st.Uptime = time.Since(t)
				}
			}
		}
	}
	if !st.Running {
		st.LastError = m.getLastError()
		if st.LastError == "" {
			st.LastError = m.systemdFailureDetails(ctx)
		}
	}
	return st, nil
}

func (m *systemdManager) waitActive(ctx context.Context, op string) error {
	deadline := time.Now().Add(2 * time.Second)
	for {
		active, _ := exec.CommandContext(ctx, "systemctl", "is-active", m.service).Output()
		switch strings.TrimSpace(string(active)) {
		case "active":
			m.setLastError("")
			return nil
		case "failed":
			msg := m.systemdFailureDetails(ctx)
			if msg == "" {
				msg = "systemd service failed"
			}
			m.setLastError(msg)
			return fmt.Errorf("%s sing-box: %s", op, msg)
		}
		if time.Now().After(deadline) {
			msg := m.systemdFailureDetails(ctx)
			if msg == "" {
				msg = fmt.Sprintf("systemd service is %s", strings.TrimSpace(string(active)))
			}
			m.setLastError(msg)
			return fmt.Errorf("%s sing-box: %s", op, msg)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *systemdManager) recordSystemdError(ctx context.Context, err error) error {
	msg := err.Error()
	if details := m.systemdFailureDetails(ctx); details != "" {
		msg = strings.TrimSpace(msg + ": " + details)
	}
	msg = trimDiagnostic(msg)
	m.setLastError(msg)
	return fmt.Errorf("%s", msg)
}

func (m *systemdManager) systemdFailureDetails(ctx context.Context) string {
	var parts []string
	if props, err := exec.CommandContext(ctx, "systemctl", "show", m.service,
		"--property=ActiveState,SubState,Result,ExecMainStatus,ExecMainCode").Output(); err == nil {
		propsText := strings.TrimSpace(string(props))
		if propsText != "" && !systemdPropsAreClean(propsText) {
			parts = append(parts, propsText)
		}
	}
	if len(parts) > 0 {
		journal, err := exec.CommandContext(ctx, "journalctl", "-u", m.service, "-n", "20", "--no-pager", "--output=cat").CombinedOutput()
		if err != nil {
			return trimDiagnostic(strings.Join(parts, "\n"))
		}
		journalText := strings.TrimSpace(string(journal))
		if journalText != "" {
			parts = append(parts, journalText)
		}
	}
	return trimDiagnostic(strings.Join(parts, "\n"))
}

func systemdPropsAreClean(props string) bool {
	values := map[string]string{}
	for _, line := range strings.Split(props, "\n") {
		key, val, ok := strings.Cut(strings.TrimSpace(line), "=")
		if ok {
			values[key] = val
		}
	}
	return values["ActiveState"] == "inactive" &&
		values["SubState"] == "dead" &&
		(values["Result"] == "" || values["Result"] == "success") &&
		(values["ExecMainStatus"] == "" || values["ExecMainStatus"] == "0")
}

func (m *systemdManager) setLastError(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastError = trimDiagnostic(msg)
}

func (m *systemdManager) getLastError() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastError
}

// --- subprocess ---

type subprocessManager struct {
	cfg     ProcessConfig
	logSink io.Writer
	log     *slog.Logger

	mu        sync.Mutex
	cmd       *exec.Cmd
	startedAt time.Time
	lastError string
}

func newSubprocessManager(cfg ProcessConfig, logSink io.Writer, log *slog.Logger) *subprocessManager {
	return &subprocessManager{cfg: cfg, logSink: logSink, log: log}
}

func (m *subprocessManager) Start(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running() {
		return nil
	}
	cmd := exec.Command(m.cfg.Binary, "run", "-c", m.cfg.ConfigPath)
	if m.cfg.WorkingDir != "" {
		cmd.Dir = m.cfg.WorkingDir
	}
	tail := newBoundedLogWriter(m.logSink, 8192)
	cmd.Stdout = tail
	cmd.Stderr = tail
	if err := cmd.Start(); err != nil {
		msg := fmt.Sprintf("start sing-box: %v", err)
		m.lastError = msg
		return fmt.Errorf("%s", msg)
	}
	m.cmd = cmd
	m.startedAt = time.Now()
	m.lastError = ""

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case err := <-waitCh:
		msg := subprocessExitMessage(err, tail.String())
		m.cmd = nil
		m.lastError = msg
		if m.log != nil {
			m.log.Error("core exited during start", slog.String("error", msg))
		}
		return fmt.Errorf("start sing-box: %s", msg)
	case <-timer.C:
		go m.reap(cmd, waitCh, tail)
	}
	return nil
}

func (m *subprocessManager) Stop(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running() {
		return nil
	}
	if err := m.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal sing-box: %w", err)
	}
	m.cmd = nil
	m.lastError = ""
	return nil
}

func (m *subprocessManager) Restart(ctx context.Context) error {
	if err := m.Stop(ctx); err != nil {
		return err
	}
	if d := m.cfg.RestartDelay; d > 0 {
		time.Sleep(d)
	}
	return m.Start(ctx)
}

func (m *subprocessManager) Reload(ctx context.Context) error {
	// SIGHUP delivery is unreliable across platforms. For subprocess mode
	// we restart the core atomically — the config is already written by
	// the applier before calling Reload.
	return m.Restart(ctx)
}

func (m *subprocessManager) Status(ctx context.Context) (Status, error) {
	m.mu.Lock()
	st := Status{Running: m.running()}
	if st.Running {
		st.PID = m.cmd.Process.Pid
		st.Uptime = time.Since(m.startedAt)
	}
	st.LastError = m.lastError
	m.mu.Unlock()
	if !st.Running {
		if external, ok := externalProcessStatus(ctx, m.cfg); ok {
			st.Running = true
			st.PID = external.PID
			st.Uptime = external.Uptime
			st.LastError = ""
		}
	}
	st.Version = coreVersion(ctx, m.cfg.Binary)
	return st, nil
}

func (m *subprocessManager) reap(cmd *exec.Cmd, waitCh <-chan error, tail *boundedLogWriter) {
	err := <-waitCh
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cmd != cmd {
		return
	}
	msg := subprocessExitMessage(err, tail.String())
	m.cmd = nil
	m.lastError = msg
	if m.log != nil {
		m.log.Error("core exited", slog.String("error", msg))
	}
}

// running reports whether a managed process is currently tracked. Callers must
// hold m.mu.
func (m *subprocessManager) running() bool {
	return m.cmd != nil && m.cmd.Process != nil
}

type boundedLogWriter struct {
	mu  sync.Mutex
	dst io.Writer
	max int
	buf []byte
}

func newBoundedLogWriter(dst io.Writer, max int) *boundedLogWriter {
	if max <= 0 {
		max = 8192
	}
	return &boundedLogWriter{dst: dst, max: max}
}

func (w *boundedLogWriter) Write(p []byte) (int, error) {
	if w.dst != nil {
		_, _ = w.dst.Write(p)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	if len(w.buf) > w.max {
		w.buf = append([]byte(nil), w.buf[len(w.buf)-w.max:]...)
	}
	return len(p), nil
}

func (w *boundedLogWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return trimDiagnostic(string(w.buf))
}

func subprocessExitMessage(err error, output string) string {
	msg := "sing-box exited"
	if err != nil {
		msg = fmt.Sprintf("sing-box exited: %v", err)
	}
	if output != "" {
		msg += ": " + output
	}
	return trimDiagnostic(msg)
}

func trimDiagnostic(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 4096 {
		return s
	}
	return strings.TrimSpace(s[len(s)-4096:])
}
