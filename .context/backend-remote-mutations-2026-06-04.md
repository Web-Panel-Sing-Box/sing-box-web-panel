# Backend Remote Mutations - 2026-06-04

## Task

- Linear issue: `SIN-57`.
- Branch: `vadigofficial/sin-57-remote-mutacii-ne-vistavleni-v-apiui-mastera-tolko`.
- Goal: expose remote inbound and client mutations through the master panel API and UI instead of limiting the master to read-only remote cache views.

## What Changed

- Added master-side remote dispatch for inbounds and clients.
  - `POST /api/inbounds` and `POST /api/clients` now accept optional `nodeId`.
  - `POST /api/nodes/{id}/inbounds` and `POST /api/nodes/{id}/clients` are supported aliases.
  - Existing update, delete, toggle, reset traffic, and status endpoints detect cached remote rows and call the target node API.
- Extended the node HTTP client with write operations for `/api/node/v1/*`.
  - Inbounds: create, update, delete, toggle.
  - Clients: create, update, delete, reset traffic, set status.
  - Bearer token auth, configured base path, SSRF guard, and TLS verification options are preserved.
- Added cache mapping and refresh logic in the node service.
  - Master local IDs are mapped to target `remoteId` values before outbound node calls.
  - Returned remote resources are upserted back into the local cache.
  - Remote 404 on delete is treated as a stale cache row and removes the cache entry.
  - Cross-node and local-to-remote moves are rejected with validation errors.
- Preserved SIN-52 TLS fields in remote inbound payloads and cache updates.
  - `acmeDomain`, `acmeEmail`, `certPath`, `keyPath`, and `allowInsecure` are sent to nodes and retained in cached remote rows.
- Guarded local services so cached remote rows cannot be mutated as if they were local sing-box resources.

## Frontend

- Added node selectors to inbound and client create flows.
- Edit flows show node ownership as read-only and filter client inbound choices to the same node.
- Store and API request types now pass optional `nodeId`.
- Touched mutation flows now await API calls and show backend error messages through toasts.
- Added missing node-related i18n strings for English and Russian.

## Main Files Touched

- Backend:
  - `internal/services/node/client.go`
  - `internal/services/node/service.go`
  - `internal/transport/handler/inbound_handler.go`
  - `internal/transport/handler/client_handler.go`
  - `internal/repo/sqlite/inbound_repo.go`
  - `internal/repo/sqlite/client_repo.go`
- Frontend:
  - `frontend/src/hooks/useInboundForm.ts`
  - `frontend/src/components/inbounds/inbound-form-modal.tsx`
  - `frontend/src/components/clients/add-client-modal.tsx`
  - `frontend/src/components/clients/client-detail-modal.tsx`
  - `frontend/src/lib/store.tsx`
  - `frontend/src/lib/i18n.tsx`
- Tests:
  - `tests/services/node/client_test.go`
  - `tests/services/node/service_remote_test.go`
  - `tests/transport/handler/inbound_handler_test.go`
  - `frontend/src/components/inbounds/inbound-form-modal.test.tsx`

## Verification

- `go test ./tests/...`
- `go vet ./...`
- `go build ./...`
- `cd frontend && pnpm typecheck`
- `cd frontend && pnpm test`
- `cd frontend && pnpm build`
- `git diff --check`

