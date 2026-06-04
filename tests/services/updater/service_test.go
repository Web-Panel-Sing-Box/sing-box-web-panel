package updater_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sing-box-web-panel/internal/services/updater"
)

type fakeReleaseClient struct {
	release updater.Release
	err     error
}

func (c fakeReleaseClient) Latest(context.Context, string) (updater.Release, error) {
	return c.release, c.err
}

type blockingRunner struct {
	started chan struct{}
	release chan struct{}
}

func (r blockingRunner) Run(context.Context, string, time.Duration) ([]byte, error) {
	close(r.started)
	<-r.release
	return []byte("updated"), nil
}

func TestVersionStatusDetectsAvailableUpdate(t *testing.T) {
	svc := updater.New(updater.Config{
		Repo:           "owner/repo",
		CurrentVersion: "1.0.0",
		CheckCacheTTL:  time.Minute,
	}, fakeReleaseClient{release: updater.Release{Version: "v1.1.0", URL: "https://example.test/release"}}, nil, discardLogger())

	status, err := svc.VersionStatus(context.Background())
	if err != nil {
		t.Fatalf("VersionStatus error = %v", err)
	}
	if !status.UpdateAvailable {
		t.Fatal("UpdateAvailable = false, want true")
	}
	if status.LatestVersion != "1.1.0" {
		t.Fatalf("LatestVersion = %q, want 1.1.0", status.LatestVersion)
	}
	if status.Status != "update_available" {
		t.Fatalf("Status = %q, want update_available", status.Status)
	}
}

func TestVersionStatusKeepsDevelopmentBuildNotUpgradeable(t *testing.T) {
	svc := updater.New(updater.Config{
		Repo:           "owner/repo",
		CurrentVersion: "dev-abcd123",
	}, fakeReleaseClient{release: updater.Release{Version: "v9.0.0"}}, nil, discardLogger())

	status, err := svc.VersionStatus(context.Background())
	if err != nil {
		t.Fatalf("VersionStatus error = %v", err)
	}
	if status.UpdateAvailable {
		t.Fatal("UpdateAvailable = true, want false")
	}
	if status.Status != "development" {
		t.Fatalf("Status = %q, want development", status.Status)
	}
}

func TestStartUpdateRejectsMissingHelper(t *testing.T) {
	svc := updater.New(updater.Config{
		ScriptPath:     filepath.Join(t.TempDir(), "missing"),
		CurrentVersion: "1.0.0",
	}, fakeReleaseClient{}, nil, discardLogger())

	_, err := svc.StartUpdate(context.Background())
	if !errors.Is(err, updater.ErrNotConfigured) {
		t.Fatalf("StartUpdate error = %v, want ErrNotConfigured", err)
	}
}

func TestStartUpdateRejectsConcurrentRun(t *testing.T) {
	script := filepath.Join(t.TempDir(), "update")
	if err := os.WriteFile(script, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	runner := blockingRunner{started: make(chan struct{}), release: make(chan struct{})}
	svc := updater.New(updater.Config{
		ScriptPath:     script,
		CurrentVersion: "1.0.0",
	}, fakeReleaseClient{}, runner, discardLogger())

	if _, err := svc.StartUpdate(context.Background()); err != nil {
		t.Fatalf("StartUpdate error = %v", err)
	}
	_, err := svc.StartUpdate(context.Background())
	if !errors.Is(err, updater.ErrAlreadyRunning) {
		t.Fatalf("concurrent StartUpdate error = %v, want ErrAlreadyRunning", err)
	}
	close(runner.release)
	<-runner.started
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
