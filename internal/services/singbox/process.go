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
}

func (m *systemdManager) systemctl(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "systemctl", append(args, m.service)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}

func (m *systemdManager) Start(ctx context.Context) error   { return m.systemctl(ctx, "start") }
func (m *systemdManager) Stop(ctx context.Context) error    { return m.systemctl(ctx, "stop") }
func (m *systemdManager) Restart(ctx context.Context) error { return m.systemctl(ctx, "restart") }
func (m *systemdManager) Reload(ctx context.Context) error  { return m.systemctl(ctx, "reload") }

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
	return st, nil
}

// --- subprocess ---

type subprocessManager struct {
	cfg     ProcessConfig
	logSink io.Writer
	log     *slog.Logger

	mu        sync.Mutex
	cmd       *exec.Cmd
	startedAt time.Time
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
	cmd.Stdout = m.logSink
	cmd.Stderr = m.logSink
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start sing-box: %w", err)
	}
	m.cmd = cmd
	m.startedAt = time.Now()

	go func() {
		_ = cmd.Wait()
		m.mu.Lock()
		if m.cmd == cmd {
			m.cmd = nil
		}
		m.mu.Unlock()
	}()
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
	m.mu.Unlock()
	st.Version = coreVersion(ctx, m.cfg.Binary)
	return st, nil
}

// running reports whether a managed process is currently tracked. Callers must
// hold m.mu.
func (m *subprocessManager) running() bool {
	return m.cmd != nil && m.cmd.Process != nil
}
