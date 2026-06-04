# Installer Hardening And Naive Inbound Fix - 2026-06-02

## Scope

Implemented the working package for Linear issues `SIN-27`, `SIN-28`,
`SIN-29`, `SIN-30`, `SIN-31`, and `SIN-33` on branch
`Vadim-Denisovich/installer-vps-hardening-naive-fix`.

This pass focused on code and documentation changes that can be verified
locally without mutating the production VPS. The production no-downtime VPS
smoke task remains represented by `SIN-27`; the saved runbook is
`.context/vps-install-smoke-plan-2026-06-02.md`.

## What Changed

### Installer

- Updated the default release repository to the canonical
  `Web-Panel-Sing-Box/shilka-web-panel`.
- Added `port_in_use()` and reused it for random-port selection and selected
  panel port validation.
- Added an explicit install exposure choice:
  - direct high-port public listener;
  - reverse-proxy mode, binding Shilka to `127.0.0.1:<port>`.
- In reverse-proxy mode, generated config now uses `tls.mode: "off"` and a
  public URL without the internal port.
- In direct mode, generated `subscription.public_url` now includes the generated
  `PANEL_PATH`, so subscription links are emitted under the real panel prefix.
- Added warnings when standalone Let's Encrypt is selected while port `80` is
  already occupied.
- Added release asset checksum verification for downloaded Shilka binaries:
  `checksums.txt` is fetched from the selected release and the selected
  `shilka-linux-{arch}` asset must match before installation.

### Naive Inbounds

- Added frontend conversion helpers:
  - API empty/unknown network values map to UI `both`;
  - UI `both` is omitted from API payloads.
- Updated the inbound form to stop sending `naiveNetwork: "both"`.
- Updated inbound row rendering so empty Naive network values display as
  `TCP+UDP`.
- Updated backend inbound service to tolerate legacy `NaiveNetwork: "both"` and
  normalize it to the empty auto-mode value before persistence.
- Added focused frontend and backend regression coverage.

### README

- Updated the VPS quick-start installer URL to the canonical
  `shilka-web-panel` repository.
- Fixed the backend development command to `go run ./cmd`, which includes
  `cmd/embed.go`.

## Files Touched

- `scripts/install.sh`
- `README.md`
- `frontend/src/api/types.ts`
- `frontend/src/api/types.test.ts`
- `frontend/src/hooks/useInboundForm.ts`
- `frontend/src/components/inbounds/inbound-row.tsx`
- `internal/services/inbound/service.go`
- `tests/services/inbound/service_test.go`

## Verification

Passed:

- `bash -n scripts/install.sh`
- `shellcheck scripts/install.sh`
- live checksum parser check against Shilka release `v1.7.0`
- `go test ./tests/services/inbound`
- `go build ./...`
- `go vet ./...`
- `go test ./tests/...`
- `cd frontend && pnpm test -- types.test.ts InboundsPage.test.tsx`
- `cd frontend && pnpm typecheck`
- `cd frontend && pnpm test`
- `cd frontend && pnpm build`
- `git diff --check`

## Notes

- The production VPS install smoke was not run as part of this code pass. It is
  intentionally tracked separately by `SIN-27` because it mutates systemd,
  `/opt/shilka`, `/etc/shilka`, `/var/lib/shilka`, and potentially firewall
  state on the server.
- The saved VPS runbook remains the source of truth for the no-downtime
  production smoke sequence.
