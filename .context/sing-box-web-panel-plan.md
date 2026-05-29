# Sing-Box Web Panel Implementation Plan

## Summary

Build a local-first control panel for `sing-box` with Go, SQLite, Vite+React SPA, Tailwind, Framer Motion, and systemd/Bash automation. The panel is not tied to any external infrastructure and controls only the local `sing-box` process through `127.0.0.1`.

## Repository

Monorepo with `cmd/`, `internal/`, `frontend/`, `config/`, `scripts/`, `systemd/`, `.context/`, `README.md`, and `AGENTS.md`.

## Backend (Go)

- Go 1.26.2 with `cleanenv` for YAML config + env var override.
- SQLite via `modernc.org/sqlite` (pure Go, no CGo). WAL mode, `synchronous=NORMAL`, mmap, batched writes.
- `log/slog` for structured logging, `github.com/fatih/color` for dev-mode pretty output.
- Embedded `//go:embed` migrations with version tracking in `schema_migrations`. Applied automatically on startup via `sqlite.New()`.
- Graceful shutdown via `signal.NotifyContext` (SIGINT/SIGTERM) with HTTP `ShutdownTimeout`.
- Tables: `admins`, `admin_recovery_codes`, `inbounds`, `users`, `user_inbounds`, `traffic_ledger`, `settings`, `config_revisions`, `subscriptions`.

### Implemented

- **Auth**: JWT cookie auth (HS256), Argon2id password hashing (PHC format), TOTP 2FA (SHA1, 6 digits, 30s), recovery codes (Argon2id-hashed, one-time use).
- **Bootstrap**: first admin auto-created from config if `admins` table is empty.
- **HTTP server**: with swagger, cors, auth, and request logging middleware.
- **API endpoints**: login, login/recovery, me, logout, totp/setup, totp/confirm, totp/disable, change-password, health, root.
- **Swagger UI**: served at `/swagger/`, spec generated via `swaggo/swag` annotations.

### Domain Models

- `Admin` — username, password_hash, totp_secret, is_totp_enabled, totp_confirmed_at
- `RecoveryCode` — admin_id, code_hash, is_used, used_at
- `Inbound` — remark, protocol (vless/hysteria2/naive), port, enabled, config_json (protocol-specific)
- `User` — name, traffic counters, quota, expiry, status
- `TrafficEntry`, `Setting`, `ConfigRevision`, `Subscription`

### Future

- `LocalConfigGenerator` that reads SQLite, builds full `config.json`, validates with `sing-box check`, writes atomically, records revisions.
- `ProcessManager` adapters for systemd and direct subprocess mode.
- `TrafficBackgroundWorker` with live speed polling, per-user source adapters, quota enforcement, batched SQLite writes.
- CRUD for users and inbounds, subscription links, QR generation, dashboard metrics, log reads, core process actions.

## Frontend

- Vite + React 19 + TypeScript SPA. Tailwind CSS, Framer Motion, Recharts, React Router DOM.
- True black background (`#171717`), mono typography (Inter + JetBrains Mono), restrained panels, neon accents.
- Screens: Dashboard, Inbounds, Clients, Settings, Logs.
- Dev server proxies `/api/*` to `127.0.0.1:8080`.
- Current state: fully functional **mock UI** with in-memory store. All data shapes defined, ready for real API integration.

## Tests

- Located in `tests/` mirroring `internal/` structure.
- External test packages (`package foo_test`).
- Tests for all implemented packages: Argon2, JWT, TOTP, AuthService (with mocks), CORS middleware, Auth middleware, Logger middleware, Health handler, Auth handler.
- Run: `go test ./tests/...`

## CI/CD

- GitHub Actions: Go build, vet, test (`go build ./...`, `go vet ./...`, `go test ./tests/...`).
- Frontend: typecheck, unit tests, build, Playwright smoke tests.
- Shell scripts: shellcheck lint.

## Config

- Primary: `config/dev.yaml` (YAML). Secrets overridden via environment variables.
- `SING_GROK_CONFIG_PATH` env var to point to a different config file.
- `cleanenv.ReadConfig()` reads YAML first, then overrides from env vars.
- Sections: runtime (GoMemLimit, GoGC), database (SQLite pragmas), http, frontend, auth, sing_box, metrics, logging, subscription.

## Check Commands

```
go build ./...
go vet ./...
go test ./tests/...
swag init -g cmd/main.go -o docs --parseDependency --parseInternal
cd frontend && pnpm typecheck
cd frontend && pnpm build
cd frontend && pnpm test
```

## Security and Disk I/O

- sing-box Clash/V2Ray APIs must bind only to `127.0.0.1`.
- Secrets must not be logged, committed to git, or hardcoded in YAML.
- SQLite uses WAL, `synchronous=NORMAL`, `busy_timeout`, foreign keys, and batched traffic writes.
- Shell commands must avoid interpolating untrusted user input.
- Recovery codes and passwords are Argon2id-hashed. TOTP secrets stored in DB (future: AES-encrypted).

## Acceptance Goals

- Installer can deploy on Ubuntu/Debian VPS with one command.
- Panel can create an inbound/client, generate a valid link/QR, validate and apply sing-box config, restart local core, stream logs, and enforce traffic limits.
- Optimized for weak VDS (256–512 MB RAM, single vCPU).