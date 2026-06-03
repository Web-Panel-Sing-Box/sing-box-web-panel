# Superpanel Nodes, API Tokens, CLI, and Installer - 2026-06-02

## Status

- Linear issue: `SIN-24`, completed.
- Task branch: `Vadim-Denisovich/superpanel-nodes-cli-installer`.
- Main implementation commit: `6a8a6b2 feat(superpanel): add nodes and CLI`.
- Pull request: https://github.com/Web-Panel-Sing-Box/shilka-web-panel/pull/5.
- Merge result: PR #5 merged into `main`; GitHub merge commit `0a6e95d7688bd167a52e5c1c01274d8eb5608fc0`.
- Scope: Shilka-to-Shilka superpanel support. 3x-ui was used as an implementation reference, not as a compatibility target.

## Reference Findings

3x-ui recently added a useful node-management shape: API tokens, a Nodes page, remote runtime dispatch, heartbeat workers, traffic sync workers, and a shell menu/installer. The Shilka implementation follows that architecture where it fits the local-first model, but deliberately tightens the security model:

- Local API tokens are not stored in plaintext. Only a token hash and short prefix are stored, and the raw token is shown once at creation time.
- Panel-to-panel API access is separated from browser-admin JWT access. API tokens are accepted only by the node API surface.
- Remote node bearer tokens are stored only because outbound node calls require them. They are masked in UI/API responses and must never be logged.
- Node HTTP calls use an SSRF guard by default. Loopback, private, link-local, and unspecified addresses are rejected unless the node explicitly enables `allow_private_address`.
- Password reset updates the admin password hash in SQLite. It must never delete `panel.db`.
- CLI and installer use the Shilka binary as the source of truth; shell menu actions delegate to binary subcommands.

## Backend And Database Changes

- Added migration `internal/repo/migrator/migrations/000008_create_superpanel.sql`.
- Added `api_tokens` and `nodes` tables.
- Rebuilt inbound and client tables to support remote ownership metadata:
  - `node_id`
  - `remote_id`
  - `last_synced_at`
- Local resources are represented by `node_id IS NULL`.
- Remote cached resources are represented by `node_id` plus `remote_id`.
- Local uniqueness remains strict for local rows, while remote rows can duplicate names/ports across different nodes.
- The sing-box generator uses local enabled rows only, so remote cache rows do not alter the master's local sing-box config.

New or updated domain/repository files include:

- `internal/domain/api_token.go`
- `internal/domain/node.go`
- `internal/domain/inbound.go`
- `internal/domain/client.go`
- `internal/repo/sqlite/api_token_repo.go`
- `internal/repo/sqlite/node_repo.go`
- `internal/repo/sqlite/inbound_repo.go`
- `internal/repo/sqlite/client_repo.go`

## API Tokens And Node Services

- Added `internal/services/apitoken/service.go` for token creation, hash matching, prefix display, enable/disable, revoke, and `last_used_at` updates.
- Added `internal/services/node/` for node management, remote HTTP calls, status probes, imports, syncs, heartbeat, and background worker orchestration.
- Node client behavior:
  - normalizes panel URLs;
  - sends `Authorization: Bearer <token>`;
  - uses short timeouts for status/sync paths and longer write timeouts;
  - applies SSRF checks before outbound requests;
  - never logs bearer token values.

## HTTP Handlers And Middleware

- Added API token management routes through `internal/transport/handler/api_token_handler.go`.
- Added node management and target-side node API routes through `internal/transport/handler/node_handler.go`.
- Updated auth middleware so regular UI/API endpoints remain JWT-only, while `/api/node/v1/*` accepts scoped API tokens.
- Added management endpoints under `/api/nodes` for list, create, update, delete, toggle, probe, import, sync, and status.
- Added panel-to-panel endpoints under `/api/node/v1/*` for status, snapshot, inbound operations, client operations, and core reload.

## CLI And Installer

- Split startup so the binary supports subcommands cleanly:
  - `shilka run`
  - `shilka setting ...`
  - `shilka admin reset-password`
  - `shilka api-token ...`
  - `shilka node ...`
  - `shilka core ...`
  - `shilka cert ...`
- Implemented CLI in `cmd/cli.go` with Go stdlib `flag`, not Cobra/Survey/Kardianos/service.
- Config writes use YAML parsing with backup/temp/rename flow instead of `sed`.
- Admin password reset updates the SQLite admin hash and leaves the database intact.
- Updated installer/menu behavior so systemd starts `shilka run`, and `/usr/local/bin/shilka` delegates binary subcommands instead of duplicating application logic.
- The installer retains the current mainline TLS/domain flow while preserving the important safety fix: no password reset path deletes `panel.db`.

## Frontend Changes

- Added node/API-token client APIs:
  - `frontend/src/api/nodes.ts`
  - `frontend/src/api/apiTokens.ts`
- Added the Nodes page:
  - `frontend/src/pages/NodesPage.tsx`
- Added `/nodes` route, sidebar entry, translations, node badges, and node filters for cached remote resources.
- Added an API-token creation modal pattern where the raw token is visible only immediately after creation.
- Fixed `frontend/pnpm-workspace.yaml` so `pnpm` workspace commands run reliably.

## Verification

The implementation was verified with:

- `go test -count=1 ./...`
- `go vet ./...`
- `go build ./...`
- `cd frontend && pnpm typecheck`
- `cd frontend && pnpm test`
- `cd frontend && pnpm build`
- `bash -n scripts/install.sh`
- `git diff --check`
- GitHub PR checks for backend, frontend, and shell on PR #5.

## Current Boundaries And Follow-ups

- v1 supports Shilka nodes only. It does not import or operate 3x-ui nodes directly.
- The master panel stores a cache of remote resources, but each node remains the source of truth for its own sing-box config.
- The target-side node API and node sync/import flows are in place. Before calling the remote mutation UX complete, verify and extend existing Inbounds/Clients create/edit forms so selected-node mutations dispatch remotely end-to-end.
- Keep node sync writes bounded and batched. Do not add per-poll SQLite writes in background jobs.
- Keep API-token auth isolated to `/api/node/v1/*`; browser UI endpoints should continue to reject API-token bearer credentials.
