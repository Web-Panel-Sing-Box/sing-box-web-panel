# AGENTS.md

This repository implements a local-first web panel for `sing-box`. Keep changes scoped, secure by default, and aligned with the existing stack.

## Stack

- **Backend** — Go 1.26, `cleanenv` (YAML config + env override), `modernc.org/sqlite` (pure-Go SQLite driver), `log/slog` (structured logging), `github.com/fatih/color` (dev-mode colored output).
- **Frontend** — Vite 8 + React 19 + TypeScript (strict) SPA. Tailwind CSS 4, Framer Motion 12 (via LazyMotion), Recharts 3, React Router DOM 7. Lives in `frontend/`.
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

## Frontend Conventions

- Data lives in `frontend/src/lib/mock/store.tsx` (React context, in-memory). Real `/api/*` wiring runs through the Vite proxy in `vite.config.ts`; do not call `fetch` from components.
- Sidebar is permanently collapsed to a 64 px icons-only rail on desktop (`lg:`) and slides in as a 260 px drawer on smaller widths — no hover-expand or pin.
- Long tables (Clients, Inbounds) live inside a Card with `max-h-[calc(100dvh-NNNpx)] min-h-[…]`, an inner `flex-1 overflow-auto` wrapper, and a `sticky top-0 z-10 bg-surface` header. The page itself never gains a second scrollbar.
- The Logs page wrapper is `flex h-[calc(100dvh-…)] flex-col`; the viewer is `flex-1 min-h-0` so the log surface fills the window without overflowing the page.
- Modals render through `components/ui/modal.tsx`; they are flat (no header/footer dividers) and use the brand-green primary CTA. Toggles default to off unless the spec says otherwise.

## Frontend Performance Rules

- The mock store in `frontend/src/lib/mock/store.tsx` is split into six narrow contexts. Consume only what you need via `useMetrics()` / `useInbounds()` / `useClients()` / `useLogs()` / `useRuntime()` / `useStoreActions()`. Do **not** reintroduce a single `useStore()` that bundles ticking metrics with rarely-changing slices — it caused cascading re-renders across every page.
- Framer Motion is wrapped in `<LazyMotion features={domMax} strict>` at the App root. Import `{ m }` (not `{ motion }`) and use `<m.div>` etc. `strict` will throw at runtime if `motion.*` slips back in. `domMax` is required because `tabs.tsx` uses `layoutId` (shared layout animation).
- Routes are lazy-loaded in `App.tsx` (`React.lazy` + a single `<Suspense>` in `PanelLayout.tsx`). Heavy modals (`inbound-form-modal`, `add-client-modal`, `client-detail-modal`) are also lazy with a `prefetch on hover/focus` of their trigger button. Recharts stays eager on the dashboard — `ResponsiveContainer` mis-sizes when mounted under Suspense.
- Reusable hooks live in `frontend/src/hooks/` (`useDisclosure`, `useCopyToClipboard`, `useLogFilter`, `useClientFilter`, `useInboundForm`). Reach for these before recreating local state machines in components.
- Random helpers (`randomPort`, `randomHex`, `makeUuid`) live in `frontend/src/lib/random.ts`. Don't inline crypto-based randomness in components.
- Sourcemaps are off in regular prod builds. Use `pnpm build:analyze` (sets `ANALYZE=true`) to get sourcemaps **and** `dist/stats.html` from `rollup-plugin-visualizer`. Save snapshots in `frontend/perf-snapshots/`.
- Test setup (`frontend/src/test/test-utils.tsx`) wraps subjects in `LazyMotion` + `Suspense`. Mirror this if you write a new render helper, or `m.*` and lazy modals will fail.

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
- Frontend typecheck: `cd frontend && pnpm typecheck`
- Frontend build: `cd frontend && pnpm build`
- Frontend bundle analysis: `cd frontend && pnpm build:analyze` (writes `dist/stats.html`)
- Frontend tests: `cd frontend && pnpm test`
- Note: this repo uses `pnpm`. `pnpm install` from the `frontend/` folder may need `--ignore-workspace` because the root `pnpm-workspace.yaml` does not declare packages.

## Git Safety

- Never stage unrelated user changes.
- Prefer explicit paths for `git add`.
- Keep generated runtime data, secrets, databases, logs, and build outputs out of git.

## Dependencies

Direct (in go.mod): `github.com/ilyakaznacheev/cleanenv`, `modernc.org/sqlite`, `github.com/fatih/color`.
