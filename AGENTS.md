# AGENTS.md

This repository implements a local-first web panel for `sing-box`. Keep changes scoped, secure by default, and aligned with the existing stack.

## Stack

- **Backend** — Go 1.26, `cleanenv` (YAML config + env override), `modernc.org/sqlite` (pure-Go SQLite driver), `log/slog` (structured logging), `github.com/fatih/color` (dev-mode colored output).
- **Frontend** — Vite 8 + React 19 + TypeScript (strict) SPA. Tailwind CSS 4, Framer Motion 12 (via LazyMotion), Recharts 3, React Router DOM 7. Lives in `frontend/`.
- **Database** — single-file SQLite with WAL mode, embedded `//go:embed` migrations in `internal/repo/migrator/`, idempotent versioned SQL files.

## Project Layout

```
cmd/main.go                              # entrypoint: config load, DB init, wiring, HTTP server, graceful shutdown
docs/                                    # generated Swagger 2.0 spec (docs.go, swagger.json, swagger.yaml)
internal/
  config/config.go                       # Config struct, MustLoad() via cleanenv
  domain/                                # domain models (Admin, RecoveryCode, etc.)
    admin.go
  lib/
    auth/
      argon2.go                          # Argon2id hashing (PHC format)
      jwt.go                             # JWT create/validate + TOTP-pending tokens
      totp.go                            # TOTP secret generation, validation, recovery codes
    sl/slog.go                           # shared slog helpers (Error attr, SetupPrettySlog)
  repo/
    storage.go                           # sentinel errors (ErrNotFound, ErrExist)
    migrator/
      migrator.go                        # embed-based migration runner
      migrations/*.sql                   # versioned SQL migration files
    sqlite/
      sqlite.go                          # OpenDB with pragmas + migration call
      admin_repo.go                      # AdminRepository impl
      recovery_repo.go                   # RecoveryCodeRepository impl
  services/
    auth/
      auth.go                            # AuthService: Login, SetupTOTP, ConfirmTOTP, ChangePassword, etc.
      errors.go                          # domain errors (ErrInvalidCredentials, etc.)
      otp_adapter.go                     # TOTPProvider adapter for TOTPManager
      time.go                            # timeNow = time.Now (testability)
  transport/
    handler/
      auth_handler.go                    # /api/auth/* HTTP handlers
      health_handler.go                  # GET /, GET /health
    middleware/
      auth.go                            # JWT auth middleware (cookie + Bearer)
      cors.go                            # CORS middleware (OPTIONS preflight, origin validation)
      logger.go                          # structured request logging (method, path, status, duration)
frontend/                                # Vite SPA (see frontend/package.json)
tests/                                   # mirrored project structure, external test packages
  lib/auth/                              # Argon2, JWT, TOTP unit tests
  services/auth/                         # AuthService tests with mocks
  transport/middleware/                  # CORS, Logger, Auth middleware tests
  transport/handler/                     # Health, Auth handler tests
config/dev.yaml                          # development YAML config
```

## Development Rules

- Do not bind sing-box Clash/V2Ray/management APIs to `0.0.0.0`; they must listen on `127.0.0.1`.
- Always validate generated sing-box configs with `sing-box check` before applying them.
- Do not log JWT secrets, admin passwords, sing-box API secrets, UUID lists, or generated private keys.
- Use subprocess argument arrays for host commands; do not build shell strings from user input.
- Keep SQLite writes batched for traffic updates. Avoid per-poll disk writes from background workers.
- Frontend style is true black, minimal, mono, and kinetic. Do not copy proprietary Grok or 3x-ui assets.
- All `.context/` and `AGENTS.md` content must be written in English.
- Tests live in `tests/` mirroring the `internal/` structure. Use external test packages (`package foo_test`).

## API Endpoints

| Method | Path | Auth | Purpose |
|--------|------|:----:|---------|
| GET | `/` | — | Panel name and version |
| GET | `/health` | — | Health check |
| GET | `/swagger/*` | — | Swagger UI |
| POST | `/api/auth/login` | — | Login (username + password, optional TOTP) |
| POST | `/api/auth/login/recovery` | — | Login via recovery code |
| GET | `/api/auth/me` | JWT | Current admin profile |
| POST | `/api/auth/logout` | — | Clear cookie |
| POST | `/api/auth/totp/setup` | JWT | Generate TOTP secret + QR URI |
| POST | `/api/auth/totp/confirm` | JWT | Confirm TOTP, get recovery codes |
| POST | `/api/auth/totp/disable` | JWT | Disable TOTP (requires code) |
| POST | `/api/auth/change-password` | JWT | Change password |

## Middleware Stack (outer to inner)

```
Logger → CORS → Auth → Mux
```

- **Logger**: logs method, path, status, size, duration. 4xx→WARN, 5xx→ERROR.
- **CORS**: handles OPTIONS preflight (204), validates Origin on non-OPTIONS cross-domain requests. Same-origin always allowed. Allowed origins: localhost:3000, 127.0.0.1:3000.
- **Auth**: public paths skip auth (`/`, `/health`, `/swagger/*`, `/api/auth/login*`, `/api/auth/logout`). Protected paths require JWT via `Authorization: Bearer <token>` header or `token` cookie.

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

## Auth Implementation

- **Passwords**: Argon2id with per-password salt, stored as PHC string in `admins.password_hash`.
- **JWT**: HS256, configurable expiry (default 24h). Claims: `sub` (adminID), `iat`, `exp`.
- **TOTP**: SHA1, 6 digits, 30s period. Secret stored base32 in `admins.totp_secret`. Three states: not set up (empty secret), pending confirmation (secret set, not confirmed), active (confirmed).
- **TOTP pending tokens**: short-lived (5 min) JWT with `totp_pending: true` claim. Cannot be used to access protected endpoints.
- **Recovery codes**: 8 per admin, format `XXXX-XXXX`, hashed with Argon2id. One-time use. Regenerated on TOTP confirmation.
- **Bootstrap**: if `admins` table is empty on startup, an admin is created from `auth.admin_user` / `auth.admin_password` config values.

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
- On signal: HTTP server shuts down with `ShutdownTimeout`, database is closed, then process exits.

## Checks

- Backend build: `go build ./...`
- Backend vet: `go vet ./...`
- Backend tests: `go test ./tests/...`
- Swagger gen: `swag init -g cmd/main.go -o docs --parseDependency --parseInternal`
- Frontend typecheck: `cd frontend && pnpm typecheck`
- Frontend build: `cd frontend && pnpm build`
- Frontend bundle analysis: `cd frontend && pnpm build:analyze` (writes `dist/stats.html`)
- Frontend tests: `cd frontend && pnpm test`
- Note: this repo uses `pnpm`. `pnpm install` from the `frontend/` folder may need `--ignore-workspace` because the root `pnpm-workspace.yaml` does not declare packages.

## Git Safety

- Never stage unrelated user changes.
- Prefer explicit paths for `git add`.
- Keep generated runtime data, secrets, databases, logs, and build outputs out of git.
- The `docs/` folder (Swagger) is committed because `docs.go` compiles into the binary.

## Dependencies

Direct (in go.mod):
- `github.com/ilyakaznacheev/cleanenv` — config
- `modernc.org/sqlite` — pure-Go SQLite
- `github.com/fatih/color` — dev-mode colored output
- `golang.org/x/crypto` — Argon2id
- `github.com/golang-jwt/jwt/v5` — JWT
- `github.com/pquerna/otp` — TOTP
- `github.com/swaggo/http-swagger/v2` — Swagger UI
- `github.com/swaggo/swag` — Swagger spec generation
