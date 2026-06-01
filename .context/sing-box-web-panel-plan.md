# Shilka Implementation Plan

## Summary

Build a local-first control panel for `sing-box` with Go, SQLite, Vite+React SPA, Tailwind, Framer Motion, and systemd/Bash automation. The panel is not tied to any external infrastructure and controls only the local `sing-box` process through `127.0.0.1`.

## Repository

Monorepo with `cmd/`, `internal/`, `frontend/`, `config/`, `scripts/`, `.context/`, `README.md`, and `AGENTS.md`.

## Backend (Go)

- Go 1.26 with `cleanenv` for YAML config + env var override.
- SQLite via `modernc.org/sqlite` (pure Go, no CGo). WAL mode, `synchronous=NORMAL`, mmap, batched writes.
- `log/slog` for structured logging, `github.com/fatih/color` for dev-mode pretty output.
- Embedded `//go:embed` migrations and frontend SPA.
- Graceful shutdown via `signal.NotifyContext` (SIGINT/SIGTERM) with HTTP `ShutdownTimeout`.
- Auto-start core on panel boot; core stopped on graceful shutdown.
- Tables: `admins`, `admin_recovery_codes`, `inbounds`, `clients`, `settings`, `config_revisions`, `traffic_ledger`, `subscriptions`.

### Implemented

- **Auth**: JWT cookie/Bearer auth (HS256), Argon2id password hashing (PHC format), TOTP 2FA (SHA1, 6 digits, 30s) with pending JWT flow, recovery codes (Argon2id-hashed, one-time use).
- **Bootstrap**: first admin auto-created from config if `admins` table is empty.
- **Core lifecycle**: ProcessManager (systemd + subprocess), generator (full config.json from DB), checker (`sing-box check`), applier (render→check→atomic write→restart, debounced).
- **Traffic stats**: Clash REST API polling (global throughput, online count), V2Ray gRPC (opt-in, per-user counters), batched DB writes, quota enforcement.
- **Full protocol support**: VLESS (TCP/WS/gRPC, TLS/Reality, flow, multiplex), Hysteria2 (obfs, masquerade, bandwidth, bbr_profile, brutal_debug, realm stubs), Naive (quic_congestion_control).
- **API endpoints**: auth (login, TOTP, recovery, change-password), inbound CRUD, client CRUD, core lifecycle, dashboard metrics, logs, subscriptions, share links.
- **Swagger UI**: served at `/swagger/` (dev only), spec via `swaggo/swag`.
- **Frontend serving**: embedded `//go:embed` (prod) or disk mode (dev). SPA fallback for all unknown paths.
- **Middleware**: Logger (structured, 4xx=WARN, 5xx=ERROR), CORS (same-origin + allowed list), Auth (JWT cookie/Bearer), RateLimit (per-IP token bucket).

### Config

- Primary: `config/dev.yaml` (YAML). Secrets overridden via environment variables.
- `SHILKA_CONFIG_PATH` env var to point to a different config file.
- `cleanenv.ReadConfig()` reads YAML first, then overrides from env vars.
- Prod config template: `config/prod.yaml`.

## Frontend

- Vite + React 19 + TypeScript SPA. Tailwind CSS, Framer Motion, Recharts, React Router DOM.
- True black background, mono typography, restrained panels, neon accents.
- Screens: Dashboard, Inbounds, Clients, Settings, Logs.
- **API layer**: `src/api/` with typed DTOs and fetch functions for all backend endpoints.
- **State**: `StoreProvider` polls backend every 3s when authenticated, skips when no token.
- Auth: real login via `POST /api/auth/login`, TOTP pending token flow.
- Routes: `/login` (public), `/dashboard`, `/inbounds`, `/clients`, `/settings`, `/logs` (protected).
- Dev server proxies `/api/*` to `127.0.0.1:8080`.

## Build & Deploy

- **Embedded binary**: `pnpm build` → `rsync frontend/dist cmd/frontend/dist` → `go build ./cmd/`
- **CI**: backend (build, vet, test), frontend (typecheck, test, build, smoke tests), shellcheck
- **CD**: push to `main` → semantic-release → build linux/amd64 + linux/arm64 binaries → GitHub Release with checksums
- **install.sh**: downloads sing-box + shilka binary from GitHub releases, writes prod config, installs systemd unit

## Tests

- `tests/` mirrors `internal/` structure, external test packages.
- Unit: Argon2, JWT, TOTP, AuthService (mocks), CORS, Auth middleware, Logger middleware, handlers.
- Integration: sing-box checker, process manager, applier, full pipeline (config→core→Clash API).
- Frontend unit: vitest + testing-library.
- Frontend e2e: Playwright smoke tests.

## Security

- sing-box Clash/V2Ray APIs bind only to `127.0.0.1`.
- Secrets never logged, committed, or hardcoded in YAML.
- Recovery codes and passwords Argon2id-hashed.
- Subprocess Reload uses full Restart (SIGHUP unreliable cross-platform).
- Rate limiting on login (5/m) and general API (100/s).

## Acceptance Goals

- One-command deployment on Ubuntu/Debian VPS.
- Panel can create inbounds/clients, generate links, validate and apply config, restart core, stream logs, enforce quotas.
- Optimized for weak VDS (256–512 MB RAM, single vCPU).
- Single binary, single port, single systemd unit.
