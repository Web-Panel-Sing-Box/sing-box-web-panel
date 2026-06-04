# Backend VPS Smoke Follow-ups - 2026-06-03

## Linear

- Issue: `SIN-37` - Fix VPS smoke follow-up regressions
- Source smoke report: `.context/vps-install-smoke-execution-2026-06-03.md`

## Why

The no-downtime VPS smoke passed the main install/API/UI scenarios, but exposed several follow-up defects that affect installed panels and node-to-node management:

- `shilka core status` reported `Running: false` from a fresh CLI process even while the panel-managed sing-box subprocess was running.
- Self-node probe/sync could not connect to a high-port panel using self-signed TLS because node HTTP clients had no explicit certificate verification bypass option.
- Backend release responses returned client share links as `link`, while the frontend DTO expected `shareLink`.
- Panel version metadata used placeholders such as `0.0.0`, `shilka`, and `shilka development`.

## What Changed

### Node Self-Signed TLS Opt-in

- Added `Node.SkipTLSVerify` and service input support.
- Added migration `000010_add_node_skip_tls_verify.sql` with `skip_tls_verify BOOLEAN NOT NULL DEFAULT 0`.
- Persisted the flag in node create, update, list, and scan paths.
- Exposed the flag through the node REST DTO as `skipTlsVerify`.
- Added `shilka node add -skip-tls-verify` for CLI-created nodes.
- Wired the flag into `node.HTTPClient` as an explicit `tls.Config{InsecureSkipVerify: true}` only when enabled.
- Added a Nodes UI checkbox using localized `nodes.skipTlsVerify` copy in English and Russian.

### CLI Core Status Fallback

- Added a Linux-only `/proc` fallback for subprocess mode.
- When no in-memory child process is tracked, the fallback searches for an external process matching:
  - the configured sing-box executable,
  - `run`,
  - the configured `-c`/`--config` path.
- This matches the installed `shilka.service` model, where the panel process owns the running sing-box child but a separate CLI invocation has no in-memory process handle.
- Non-Linux builds return no fallback and keep existing behavior.

### Link DTO Compatibility

- Backend client links now include both `link` and `shareLink`.
- Frontend `getClientLinks()` accepts either field and normalizes both names for callers.
- Swagger generated schema was updated to document `shareLink`.

### Version Metadata

- Added `internal/version` as the shared panel version source.
- `/api`, `/api/node/v1/status`, and CLI `shilka version` now use the shared version.
- Release builds inject `internal/version.Version` through `-ldflags`.
- Branch builds use `dev-<sha>` instead of incorrectly exposing `main`.

## Files Touched

- `.github/workflows/release.yml`
- `cmd/cli.go`
- `cmd/main.go`
- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`
- `frontend/src/api/clients.ts`
- `frontend/src/api/nodes.ts`
- `frontend/src/lib/i18n.tsx`
- `frontend/src/pages/NodesPage.tsx`
- `internal/domain/node.go`
- `internal/repo/migrator/migrations/000010_add_node_skip_tls_verify.sql`
- `internal/repo/sqlite/node_repo.go`
- `internal/services/node/client.go`
- `internal/services/node/service.go`
- `internal/services/node/types.go`
- `internal/services/singbox/process.go`
- `internal/services/singbox/process_linux.go`
- `internal/services/singbox/process_other.go`
- `internal/transport/handler/health_handler.go`
- `internal/transport/handler/node_handler.go`
- `internal/transport/handler/subscription_handler.go`
- `internal/version/version.go`
- `tests/services/node/client_test.go`
- `tests/services/singbox/process_linux_test.go`
- `tests/transport/handler/health_handler_test.go`

## Verification

Passed:

- `go test ./tests/services/node ./tests/services/singbox ./tests/transport/handler`
- `go test ./tests/...`
- `go build ./...`
- `go vet ./...`
- `cd frontend && pnpm typecheck`
- `cd frontend && pnpm build`
- `cd frontend && pnpm test`

## Notes

- The TLS bypass is intentionally per-node and disabled by default.
- The CLI process fallback is Linux-only because installed panel behavior is Linux/systemd-oriented and `/proc` provides a low-cost deterministic check there.
- The release workflow still relies on semantic-release to publish assets. Exact semver injection can be tightened further if the workflow is later changed to build inside semantic-release's version-aware prepare phase.
