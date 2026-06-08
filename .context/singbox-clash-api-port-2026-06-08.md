# SIN-74 - Clash API Port Conflict

## Summary

SIN-74 fixed the install-time and runtime diagnostics path for sing-box Clash API port conflicts. The installer no longer blindly writes `127.0.0.1:9090` when that port is already taken, and the process manager now keeps the most recent core startup or unexpected-exit error for `/api/core/status` and `shilka core status`.

## What Changed

- Added installer support for `--clash-api-addr` and `SHILKA_CLASH_API_ADDR`, with fallback to the existing `SHILKA_SING_BOX_API_ADDRESS`.
- Added local-only Clash API validation: only `127.0.0.1:<port>` is accepted, occupied ports are rejected, and the panel port cannot be reused.
- Added automatic Clash API address selection: keep `127.0.0.1:9090` when free, otherwise choose a free high local port.
- Extended `singbox.Status` with `LastError` and exposed it as optional `lastError` in API, CLI, and Swagger docs.
- Captured a bounded tail of sing-box subprocess output while still streaming logs, so immediate bind failures are returned from `Start`.
- Added systemd startup polling and failure detail capture for systemd-managed core mode.

## Files Touched

- `scripts/install.sh`
- `internal/services/singbox/process.go`
- `internal/transport/handler/core_handler.go`
- `cmd/cli.go`
- `frontend/src/api/core.ts`
- `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`
- `tests/services/singbox/integration_test.go`
- `tests/services/singbox/process_test.go`
- `tests/transport/handler/core_handler_test.go`

## Verification

- `bash -n scripts/install.sh`
- `shellcheck scripts/install.sh scripts/update.sh scripts/uninstall.sh`
- `go test ./tests/services/singbox`
- `go test ./tests/transport/handler`
- `go test ./tests/...`
- `go build ./...`
- `go vet ./...`
- `cd frontend && pnpm typecheck`
- `cd frontend && pnpm build`

## Notes

- Existing untracked files were left untouched.
- Server verification should deploy only the built binary and reuse the existing config and TLS material. The installer must not be rerun for this check.
