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
- Graceful shutdown via `signal.NotifyContext` (SIGINT/SIGTERM).
- Tables: `admins`, `inbounds`, `users`, `traffic_ledger`, `settings`, `config_revisions`, `subscriptions`.
- Future: JWT cookie auth, Argon2id password hashing, login rate limiting, audit logs.
- Future: `LocalConfigGenerator` that reads SQLite, builds full `config.json`, validates with `sing-box check`, writes atomically, records revisions.
- Future: `ProcessManager` adapters for systemd and direct subprocess mode.
- Future: `TrafficBackgroundWorker` with live speed polling, per-user source adapters, quota enforcement, batched SQLite writes.
- Future: CRUD for users and inbounds, subscription links, QR generation, dashboard metrics, log reads, core process actions.

## Frontend

- Vite + React 18 + TypeScript SPA. Tailwind CSS, Framer Motion, Recharts, React Router DOM.
- True black background (`#171717`), mono typography (Inter + JetBrains Mono), restrained panels, neon accents.
- Screens: Dashboard, Inbounds, Clients, Settings, Logs.
- Dev server proxies `/api/*` to `127.0.0.1:8080`.
- Current state: fully functional **mock UI** with in-memory store. All data shapes defined, ready for real API integration.

## Config

- Primary: `config/dev.yaml` (YAML). Secrets overridden via environment variables.
- `SING_GROK_CONFIG_PATH` env var to point to a different config file.
- `cleanenv.ReadConfig()` reads YAML first, then overrides from env vars.
- Sections: runtime (GoMemLimit, GoGC), database (SQLite pragmas), http, frontend, auth, sing_box, metrics, logging, subscription.

## Check Commands

```
go vet ./...
go build ./cmd/
cd frontend && npm run typecheck
cd frontend && npm run build
cd frontend && npm test
```

## Security and Disk I/O

- sing-box Clash/V2Ray APIs must bind only to `127.0.0.1`.
- Secrets must not be logged, committed to git, or hardcoded in YAML.
- SQLite uses WAL, `synchronous=NORMAL`, `busy_timeout`, foreign keys, and batched traffic writes.
- Shell commands must avoid interpolating untrusted user input.

## Acceptance Goals

- Installer can deploy on Ubuntu/Debian VPS with one command.
- Panel can create an inbound/client, generate a valid link/QR, validate and apply sing-box config, restart local core, stream logs, and enforce traffic limits.
- Optimized for weak VDS (256–512 MB RAM, single vCPU).
