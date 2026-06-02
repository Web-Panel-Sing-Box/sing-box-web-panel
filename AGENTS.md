# AGENTS.md

This repository implements a local-first web panel for `sing-box`. Keep changes scoped, secure by default, and aligned with the existing stack.

## Stack

- **Backend** — Go 1.26, `cleanenv` (YAML config + env override), `modernc.org/sqlite` (pure-Go SQLite driver), `log/slog` (structured logging), `github.com/fatih/color` (dev-mode colored output).
- **Frontend** — Vite 8 + React 19 + TypeScript (strict) SPA. Tailwind CSS 4, Framer Motion 12 (via LazyMotion), Recharts 3, React Router DOM 7. Lives in `frontend/`.
- **Database** — single-file SQLite with WAL mode, embedded `//go:embed` migrations in `internal/repo/migrator/`, idempotent versioned SQL files.

## Task Workflow (mandatory for every new task)

Do this **before** writing any code for a task:

1. **Branch per task.** Never commit to `main` or `develop`. Branch from `main` using `<github-username>/<task-slug>` — e.g. `Vadim-Denisovich/redesign-frontend`. This matches the existing `4444urka/*` convention. Note: git refs cannot nest (a `Vadim-Denisovich` branch and a `Vadim-Denisovich/x` branch conflict), so the username is a path *prefix*, not a standalone branch.
2. **Linear issue.** Create a Linear issue for the task in the `Sing-box-pannel` team (id `2fa98408-cff3-4f63-80e9-b8fc46adf926`) before starting. Set it to `In Progress` while working. Reference the issue identifier (e.g. `SIN-23`) in the PR title/body.

## Project Layout

```
cmd/main.go                              # entrypoint: config load, DB init, wiring, HTTP server, graceful shutdown
cmd/embed.go                             # //go:embed frontend/dist for single-binary deploys
docs/                                    # generated Swagger 2.0 spec (docs.go, swagger.json, swagger.yaml)
internal/
  config/config.go                       # Config struct, MustLoad() via cleanenv
  domain/                                # domain models (Admin, RecoveryCode, Inbound, Client, etc.)
    admin.go
    inbound.go
  lib/
    auth/
      argon2.go                          # Argon2id hashing (PHC format)
      jwt.go                             # JWT create/validate + TOTP-pending tokens
      totp.go                            # TOTP secret generation, validation, recovery codes
    keys/keys.go                         # Reality x25519 keypair, short_id, UUID, tokens, passwords
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
    auth/                                # AuthService (login, TOTP, recovery, change-password)
    inbound/service.go                   # inbound CRUD + validation + Reality/transport defaults
    client/service.go                    # client CRUD, status/quota transitions, credential + sub-token gen
    singbox/                             # config generation + core lifecycle
      schema.go                          # sing-box config structs (VLESS, Hysteria2, Naive with full fields)
      generator.go                       # render config.json from the DB (targets sing-box 1.11–1.14)
      checker.go                         # `sing-box check` wrapper
      process.go                         # ProcessManager: systemd + subprocess adapters + auto-detect
      apply.go                           # render→check→atomic write→restart→record revision (debounced)
    stats/                               # traffic metrics + quota enforcement
    sublink/                             # vless/hysteria2/naive link + subscription
    sysstat/                             # host CPU/RAM/disk/uptime (linux /proc; stub elsewhere)
    tlsmgr/tlsmgr.go                     # panel TLS: file | self-signed | acme autocert
    logbuf/logbuf.go                     # in-memory log ring
  transport/
    handler/                             # auth, health, frontend, inbound, client, core, dashboard, logs, subscription
    middleware/
      auth.go                            # JWT auth (cookie + Bearer; public paths + frontend prefix whitelist)
      cors.go                            # CORS (OPTIONS preflight, same-origin detection, origin validation)
      logger.go                          # structured request logging (method, path, status, size, duration)
      ratelimit.go                       # per-IP token-bucket (login brute-force + general API limit)
frontend/
  src/api/                               # typed DTOs + fetch functions for all backend endpoints
  src/lib/store.tsx                       # React context provider polling backend every 3s
  src/pages/                             # Dashboard, Inbounds, Clients, Settings, Logs
tests/                                   # mirrored project structure, external test packages
  lib/auth/                              # Argon2, JWT, TOTP unit tests
  services/auth/                         # AuthService tests with mocks
  services/inbound/                      # Inbound service tests
  services/singbox/                      # Generator + integration tests against real sing-box
  transport/middleware/                  # CORS, Logger, Auth middleware tests
  transport/handler/                     # Health, Auth handler tests
config/dev.yaml                          # development YAML config
config/prod.yaml                         # production YAML config template
scripts/install.sh                       # one-command VPS installer
```

## Development Rules

- Do not bind sing-box Clash/V2Ray/management APIs to `0.0.0.0`; they must listen on `127.0.0.1`.
- Always validate generated sing-box configs with `sing-box check` before applying them.
- Do not log JWT secrets, admin passwords, sing-box API secrets, UUID lists, or generated private keys.
- Use subprocess argument arrays for host commands; do not build shell strings from user input.
- Keep SQLite writes batched for traffic updates. Avoid per-poll disk writes from background workers.
- Frontend style is charcoal (`#171717` canvas, never true black), minimal, kinetic, with mono reserved for technical data (tokens, UUIDs, IPs, logs). Do not copy proprietary assets.
- All `.context/` and `AGENTS.md` content must be written in English.
- After completing any task, always record the work as a detailed English Markdown summary in `.context/`, named `<area>-<topic>-<YYYY-MM-DD>.md` (e.g. `frontend-per-protocol-inbounds-2026-06-01.md`). Cover what changed, why, the files touched, and how it was verified.
- Tests live in `tests/` mirroring the `internal/` structure. Use external test packages (`package foo_test`).

## API Endpoints

| Method | Path | Auth | Purpose |
|--------|------|:----:|---------|
| GET | `/api` | — | Panel name and version |
| GET | `/api/health` | — | Health check |
| GET | `/swagger/*` | — | Swagger UI (dev only) |
| POST | `/api/auth/login` | — | Login (username + password) |
| POST | `/api/auth/login/totp` | — | Complete TOTP login with pending token |
| POST | `/api/auth/login/recovery` | — | Login via recovery code |
| GET | `/api/auth/me` | JWT | Current admin profile |
| POST | `/api/auth/logout` | — | Clear cookie |
| POST | `/api/auth/totp/setup` | JWT | Generate TOTP secret + QR URI |
| POST | `/api/auth/totp/confirm` | JWT | Confirm TOTP, get recovery codes |
| POST | `/api/auth/totp/disable` | JWT | Disable TOTP (requires code) |
| POST | `/api/auth/change-password` | JWT | Change password |
| GET/POST | `/api/inbounds` | JWT | List / create inbounds |
| GET/PUT/DELETE | `/api/inbounds/{id}` | JWT | Read / update / delete an inbound |
| POST | `/api/inbounds/{id}/toggle` | JWT | Enable/disable an inbound |
| POST | `/api/inbounds/{id}/clone` | JWT | Clone an inbound |
| GET/POST | `/api/clients` | JWT | List / create clients |
| GET/PUT/DELETE | `/api/clients/{id}` | JWT | Read / update / delete a client |
| POST | `/api/clients/{id}/reset-traffic` | JWT | Reset a client's counters |
| POST | `/api/clients/{id}/status` | JWT | Set client status |
| GET | `/api/core/status` | JWT | sing-box process status |
| POST | `/api/core/{start\|stop\|restart\|reload}` | JWT | Lifecycle management |
| GET | `/api/core/version` | JWT | Core version |
| GET | `/api/core/config` | JWT | Preview generated config |
| GET | `/api/core/logs` | JWT | Core log lines with pagination |
| GET | `/api/dashboard/metrics` | JWT | Dashboard metrics snapshot |
| GET | `/api/dashboard/traffic` | JWT | Throughput history |
| GET | `/api/logs` | JWT | Panel log lines |
| GET | `/sub/{token}` | — | Public subscription |
| GET | `/api/subscription/{token}` | — | Public subscription (alias) |

## Middleware Stack (outer to inner)

```
Logger → RateLimit (API) → RateLimit (Login) → CORS → Auth → Mux
```

- **Logger**: logs method, path, status, size, duration. 4xx→WARN, 5xx→ERROR.
- **RateLimit**: per-IP token bucket. Login limit (5/m) on `/api/auth/login*`. API limit (100/s) on everything else.
- **CORS**: handles OPTIONS preflight (204), validates Origin on cross-domain requests. Same-origin always allowed.
- **Auth**: public: `/api/auth/login*`, `/api/auth/logout`, `/api`, `/api/health`, `/swagger`, `/sub/`, frontend paths (any non-`/api` prefix). Protected: JWT via `Authorization: Bearer` or `token` cookie.

## Frontend Conventions

- API layer in `frontend/src/api/` with typed DTOs and fetch wrappers (Bearer token stored in localStorage).
- Data lives in `frontend/src/lib/store.tsx` (React context polling backend every 3s when authenticated).
- Sidebar: 64px icons-only rail on desktop, 260px slide-in drawer on mobile.
- Long tables: Card with `max-h-[calc(100dvh-NNNpx)]`, inner `overflow-auto`, sticky header.
- Modals via `components/ui/modal.tsx`, flat (no header/footer dividers).
- Framer Motion in `<LazyMotion features={domMax} strict>`. Import `{ m }` only.
- Routes lazy-loaded via `React.lazy` + `<Suspense>`. Router is `HashRouter`.
- Reusable hooks in `frontend/src/hooks/`.
- **Accessibility.** Tree wrapped in `<MotionConfig reducedMotion="user">`; honor `prefers-reduced-motion` (also via the CSS fallback in `index.css`). Keyboard `:focus-visible` ring is global in `index.css` — do not strip outlines. Animate only `transform`/`opacity`.
- **Copy.** No em-dashes (`—`/`–`) anywhere in user-visible strings — use a hyphen `-`. Titles only, no filler subtitles. Every visible string goes through `lib/i18n.tsx` in both `en` and `ru`. No stale "mock"/"demo" wording (the backend is wired).
- **Mutations surface errors.** Components call store actions (which call the API); always `await` them and show failures via a toast. Extract the backend message from `ApiError.body.error`, falling back to an i18n key. Never fire-and-forget a mutation (it silently lies on failure).
- **Core lifecycle is real.** Start/stop go through store `startCore`/`stopCore` → `POST /api/core/start|stop`; the status pill reflects the backend poll, never an optimistic local flip. sing-box must be installed (PATH or `sing_box.binary_path`) for the core to actually run.
- **Inbound TLS rules.** TLS is constrained per protocol in `useInboundForm` (`tlsForProtocol`/`buildPayload`): naive and hysteria2 require TLS, Reality is VLESS-only. Keep the form in sync with `internal/services/inbound/service.go` validation.
- **Casing.** Table column headers are sentence case. ALL-CAPS is reserved for technical identifiers only (protocol/transport chips, log levels).

## Auth Implementation

- **Passwords**: Argon2id with per-password salt, stored as PHC string in `admins.password_hash`.
- **JWT**: HS256, configurable expiry (default 24h). Claims: `sub` (adminID), `iat`, `exp`, `totp_pending` (bool).
- **TOTP**: SHA1, 6 digits, 30s period. Secret stored base32 in `admins.totp_secret`.
- **TOTP pending tokens**: short-lived (5 min) JWT with `totp_pending: true`. Issued after password verification when 2FA is enabled. Client submits pending token + TOTP code to `/api/auth/login/totp`.
- **Recovery codes**: 8 per admin, format `XXXX-XXXX`, Argon2id-hashed. Verified via `Verify()` not `Hash()` exact match.
- **Bootstrap**: if `admins` table is empty, admin created from config.

## Config

- Primary: `config/dev.yaml` (YAML). Production: `config/prod.yaml`.
- `SHILKA_CONFIG_PATH` env var for custom config path.
- `cleanenv.ReadConfig()` reads YAML, then overrides from env vars.
- All struct fields have `env-default` tags.

## Migrations

- SQL files in `internal/repo/migrator/migrations/`, embedded via `//go:embed`.
- Files named `NNNNNN_description.sql`, sorted lexicographically.
- Each migration in a transaction. Applied versions tracked in `schema_migrations`.
- Auto-run on startup via `sqlite.New()`.

## Build

- **Dev**: `go run ./cmd` (run the whole package — `go run ./cmd/main.go` alone fails because it skips `cmd/embed.go`)
- **Prod (embedded)**: `cd frontend && pnpm build && cd .. && rsync -a frontend/dist/ cmd/frontend/dist/ && go build -ldflags="-s -w" -o shilka ./cmd/`
- **Binary released via CD**: linux/amd64 + linux/arm64, attached to GitHub Release by semantic-release.
- **install.sh**: downloads pre-built binary + sing-box, writes config, installs systemd unit.

## Checks

- Backend build: `go build ./...`
- Backend vet: `go vet ./...`
- Backend tests: `go test ./tests/...`
- Swagger: `swag init -g cmd/main.go -o docs --parseDependency --parseInternal`
- Frontend typecheck: `cd frontend && pnpm typecheck`
- Frontend build: `cd frontend && pnpm build`
- Frontend tests: `cd frontend && pnpm test`

## Dependencies

Direct (in go.mod):
- `github.com/ilyakaznacheev/cleanenv` — config
- `modernc.org/sqlite` — pure-Go SQLite
- `github.com/fatih/color` — dev-mode colored output
- `golang.org/x/crypto` — Argon2id, x25519, ACME autocert
- `github.com/golang-jwt/jwt/v5` — JWT
- `github.com/pquerna/otp` — TOTP
- `github.com/swaggo/http-swagger/v2` + `github.com/swaggo/swag` — Swagger
- `google.golang.org/grpc` + `google.golang.org/protobuf` — V2Ray stats adapter (opt-in)
