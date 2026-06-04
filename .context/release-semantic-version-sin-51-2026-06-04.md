# Release Semantic Version Fix (SIN-51) - 2026-06-04

## What changed

- Moved release binary asset generation into the semantic-release `prepare` phase through `@semantic-release/exec`.
- Added `scripts/build-release-assets.sh` to build frontend assets, refresh the embedded frontend directory, compile Linux amd64/arm64 binaries, stamp `internal/version.Version` with the semantic tag, and write release checksums.
- Removed the old pre-semantic-release binary build from `.github/workflows/release.yml`, which stamped branch builds as `dev-<sha>`.
- Added regression coverage for build-time version injection, `/api` version output, and updater comparisons when the current version keeps the `v` prefix.

## Why

Release workflow previously built binaries before semantic-release created and exposed the release version. Since the workflow was running on a `main` branch push, it treated the build as a branch build and stamped `internal/version.Version` with `dev-${GITHUB_SHA::7}`. Installed release binaries then reported `dev-<sha>` from `/api`, node status, panel update status, and `shilka version`.

The new flow builds release assets only after semantic-release has computed `nextRelease.version`, so the exact tag value, for example `v1.9.2`, is passed to Go with `-ldflags`.

## Files touched

- `.github/workflows/release.yml`
- `.releaserc`
- `scripts/build-release-assets.sh`
- `tests/cmd/version_test.go`
- `tests/transport/handler/health_handler_test.go`
- `tests/services/updater/service_test.go`

## Verification

- `bash -n scripts/build-release-assets.sh`
- `go test ./tests/...`
- `go build ./...`
- `go vet ./...`
- `shellcheck scripts/build-release-assets.sh scripts/install.sh scripts/update.sh scripts/uninstall.sh`

The CLI regression test builds a temporary binary with `-X sing-box-web-panel/internal/version.Version=v9.9.9` and verifies `shilka version` returns `shilka v9.9.9`.
