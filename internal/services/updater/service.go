package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrAlreadyRunning = errors.New("update already running")
	ErrNotConfigured  = errors.New("update helper is not configured")
)

type Config struct {
	Repo           string
	ScriptPath     string
	CurrentVersion string
	CheckCacheTTL  time.Duration
	CommandTimeout time.Duration
}

type Status struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	ReleaseURL      string
	CheckedAt       time.Time
	Status          string
}

type Release struct {
	Version string
	URL     string
}

type ReleaseClient interface {
	Latest(ctx context.Context, repo string) (Release, error)
}

type CommandRunner interface {
	Run(ctx context.Context, scriptPath string, timeout time.Duration) ([]byte, error)
}

type Service struct {
	cfg    Config
	client ReleaseClient
	runner CommandRunner
	log    *slog.Logger

	mu       sync.Mutex
	cached   Status
	running  bool
	lastRun  string
	lastErr  string
	lastDone time.Time
}

func New(cfg Config, client ReleaseClient, runner CommandRunner, log *slog.Logger) *Service {
	if cfg.Repo == "" {
		cfg.Repo = "Web-Panel-Sing-Box/shilka-web-panel"
	}
	if cfg.CheckCacheTTL <= 0 {
		cfg.CheckCacheTTL = 10 * time.Minute
	}
	if cfg.CommandTimeout <= 0 {
		cfg.CommandTimeout = 10 * time.Minute
	}
	if client == nil {
		client = HTTPReleaseClient{Client: http.DefaultClient}
	}
	if runner == nil {
		runner = ShellRunner{}
	}
	return &Service{cfg: cfg, client: client, runner: runner, log: log}
}

func (s *Service) VersionStatus(ctx context.Context) (Status, error) {
	s.mu.Lock()
	if s.running {
		status := s.cached
		status.CurrentVersion = s.cfg.CurrentVersion
		status.Status = "running"
		s.mu.Unlock()
		return status, nil
	}
	if !s.cached.CheckedAt.IsZero() && time.Since(s.cached.CheckedAt) < s.cfg.CheckCacheTTL {
		status := s.cached
		if s.lastRun != "" {
			status.Status = s.lastRun
		}
		s.mu.Unlock()
		return status, nil
	}
	s.mu.Unlock()

	release, err := s.client.Latest(ctx, s.cfg.Repo)
	now := time.Now().UTC()
	status := Status{
		CurrentVersion: s.cfg.CurrentVersion,
		CheckedAt:      now,
		Status:         "up_to_date",
	}
	if err != nil {
		status.Status = "check_failed"
		s.mu.Lock()
		s.cached = status
		s.lastErr = err.Error()
		s.mu.Unlock()
		return status, nil
	}
	status.LatestVersion = normalizeVersion(release.Version)
	status.ReleaseURL = release.URL
	if isDevelopmentVersion(s.cfg.CurrentVersion) {
		status.Status = "development"
	} else if versionGreater(status.LatestVersion, s.cfg.CurrentVersion) {
		status.UpdateAvailable = true
		status.Status = "update_available"
	}

	s.mu.Lock()
	s.cached = status
	if s.lastRun != "" && s.lastRun != "failed" {
		s.lastRun = ""
	}
	s.mu.Unlock()
	return status, nil
}

func (s *Service) StartUpdate(ctx context.Context) (Status, error) {
	if err := s.helperReady(); err != nil {
		return Status{CurrentVersion: s.cfg.CurrentVersion, Status: "not_configured"}, err
	}
	s.mu.Lock()
	if s.running {
		status := s.cached
		status.CurrentVersion = s.cfg.CurrentVersion
		status.Status = "running"
		s.mu.Unlock()
		return status, ErrAlreadyRunning
	}
	s.running = true
	s.lastRun = "running"
	s.lastErr = ""
	status := s.cached
	status.CurrentVersion = s.cfg.CurrentVersion
	status.Status = "running"
	s.mu.Unlock()

	go s.runUpdate()
	return status, nil
}

func (s *Service) runUpdate() {
	ctx := context.Background()
	out, err := s.runner.Run(ctx, s.cfg.ScriptPath, s.cfg.CommandTimeout)
	if err != nil {
		s.log.Error("panel update failed",
			slog.String("error", err.Error()),
			slog.String("output", trimOutput(out)),
		)
		s.mu.Lock()
		s.running = false
		s.lastRun = "failed"
		s.lastErr = err.Error()
		s.lastDone = time.Now().UTC()
		s.mu.Unlock()
		return
	}
	s.log.Info("panel update started",
		slog.String("output", trimOutput(out)),
	)
	s.mu.Lock()
	s.running = false
	s.lastRun = "updated"
	s.lastDone = time.Now().UTC()
	s.cached.CheckedAt = time.Time{}
	s.mu.Unlock()
}

func (s *Service) helperReady() error {
	if s.cfg.ScriptPath == "" {
		return ErrNotConfigured
	}
	st, err := os.Stat(s.cfg.ScriptPath)
	if err != nil {
		return ErrNotConfigured
	}
	if st.IsDir() || st.Mode()&0o111 == 0 {
		return ErrNotConfigured
	}
	return nil
}

type HTTPReleaseClient struct {
	Client *http.Client
}

func (c HTTPReleaseClient) Latest(ctx context.Context, repo string) (Release, error) {
	repo = strings.Trim(repo, "/")
	if repo == "" || strings.Contains(repo, "..") || strings.Count(repo, "/") != 1 {
		return Release{}, fmt.Errorf("invalid release repo")
	}
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "shilka-panel-updater")
	res, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("github latest release: %s", res.Status)
	}
	var body struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return Release{}, err
	}
	if body.TagName == "" {
		return Release{}, fmt.Errorf("github latest release missing tag")
	}
	return Release{Version: body.TagName, URL: body.HTMLURL}, nil
}

type ShellRunner struct{}

func (ShellRunner) Run(ctx context.Context, scriptPath string, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var cmd *exec.Cmd
	if os.Geteuid() == 0 {
		cmd = exec.CommandContext(runCtx, scriptPath)
	} else {
		cmd = exec.CommandContext(runCtx, "sudo", "-n", scriptPath)
	}
	out, err := cmd.CombinedOutput()
	if runCtx.Err() != nil {
		return out, runCtx.Err()
	}
	return out, err
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}

func isDevelopmentVersion(v string) bool {
	v = normalizeVersion(v)
	return v == "" || v == "dev" || strings.HasPrefix(v, "dev-")
}

func versionGreater(a, b string) bool {
	aa := versionParts(a)
	bb := versionParts(b)
	if len(aa) == 0 || len(bb) == 0 {
		return normalizeVersion(a) != "" && normalizeVersion(a) != normalizeVersion(b)
	}
	for i := 0; i < len(aa) || i < len(bb); i++ {
		var av, bv int
		if i < len(aa) {
			av = aa[i]
		}
		if i < len(bb) {
			bv = bb[i]
		}
		if av > bv {
			return true
		}
		if av < bv {
			return false
		}
	}
	return false
}

func versionParts(v string) []int {
	v = normalizeVersion(v)
	main := strings.SplitN(v, "-", 2)[0]
	parts := strings.Split(main, ".")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil
		}
		out = append(out, n)
	}
	return out
}

func trimOutput(out []byte) string {
	s := strings.TrimSpace(string(out))
	if len(s) > 2000 {
		return s[len(s)-2000:]
	}
	return s
}
