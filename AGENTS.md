# AGENTS.md

This repository implements a local-first web panel for `sing-box`. Keep changes scoped, secure by default, and aligned with the existing stack.

## Stack

- **Backend** — Go 1.26, `cleanenv` (YAML config + env override), `modernc.org/sqlite` (pure-Go SQLite driver), `log/slog` (structured logging), `github.com/fatih/color` (dev-mode colored output).
- **Frontend** — Vite + React 18 + TypeScript SPA. Tailwind CSS, Framer Motion, Recharts, React Router DOM. Lives in `frontend/`.
- **Database** — single-file SQLite with WAL mode, embedded `//go:embed` migrations in `internal/repo/migrator/`, idempotent versioned SQL files.

## Project Layout

```
cmd/main.go                           # entrypoint: config load, DB init, signals, graceful shutdown
internal/
  config/config.go                    # Config struct, MustLoad() via cleanenv
  lib/sl/slog.go                      # shared slog helpers (Error attr, SetupPrettySlog)
  repo/
    storage.go                        # sentinel errors (ErrNotFound, ErrExist)
    migrator/
      migrator.go                     # embed-based migration runner
      migrations/*.sql                # versioned SQL migration files
    sqlite/
      sqlite.go                       # OpenDB with pragmas + migration call
frontend/                             # Vite SPA (see frontend/package.json)
config/dev.yaml                       # development YAML config
```

## Development Rules

- Do not bind sing-box Clash/V2Ray/management APIs to `0.0.0.0`; they must listen on `127.0.0.1`.
- Always validate generated sing-box configs with `sing-box check` before applying them.
- Do not log JWT secrets, admin passwords, sing-box API secrets, UUID lists, or generated private keys.
- Use subprocess argument arrays for host commands; do not build shell strings from user input.
- Keep SQLite writes batched for traffic updates. Avoid per-poll disk writes from background workers.
- Frontend style is true black, minimal, mono, and kinetic. Do not copy proprietary Grok or 3x-ui assets.
- All `.context/` and `AGENTS.md` content must be written in English.

## Config

- Primary config is `config/dev.yaml` (YAML). Override secrets via environment variables at runtime.
- `SING_GROK_CONFIG_PATH` env var can point to a different config file.
- `cleanenv.ReadConfig()` reads YAML first, then overrides matching fields from env vars.
- All config struct fields have `env-default` tags for sensible defaults on weak VDS.

## Migrations

- SQL files embedded via `//go:embed` in `internal/repo/migrator/migrator.go`.
- Files named `NNNNNN_description.sql`, sorted lexicographically, applied in version order.
- Each migration runs in a transaction. Applied versions tracked in `schema_migrations` table.
- Migrations run automatically on startup via `sqlite.New()`.
- All `CREATE TABLE` statements use `IF NOT EXISTS` for idempotency.

## Graceful Shutdown

- `signal.NotifyContext` catches `SIGINT` and `SIGTERM`.
- On signal: database is closed, resources released, then process exits.
- Future HTTP server shutdown will use `cfg.HTTP.ShutdownTimeout`.

## Checks

- Backend vet: `go vet ./...`
- Backend build: `go build ./cmd/`
- Frontend typecheck: `cd frontend && npm run typecheck`
- Frontend build: `cd frontend && npm run build`
- Frontend tests: `cd frontend && npm test`

## Git Safety

- Never stage unrelated user changes.
- Prefer explicit paths for `git add`.
- Keep generated runtime data, secrets, databases, logs, and build outputs out of git.

## Dependencies

Direct (in go.mod): `github.com/ilyakaznacheev/cleanenv`, `modernc.org/sqlite`, `github.com/fatih/color`.
