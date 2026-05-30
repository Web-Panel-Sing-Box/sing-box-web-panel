# Backend Implementation — sing-box Control Plane (2026-05-31)

## Context

The repository already shipped a Go backend with **authentication only** (JWT cookie auth, Argon2id,
TOTP 2FA, recovery codes, bootstrap admin, Swagger, CORS/Logger/Auth middleware, embedded SQLite
migrations) plus a React/Vite **mock UI** whose data shapes were "ready for real API integration".
The `.context/sing-box-web-panel-plan.md` "Future" section enumerated the work: config generation,
process management, traffic stats + quota enforcement, inbound/client CRUD, subscriptions, dashboard
metrics, logs, settings, and TLS.

This session implemented all of that. Everything was built incrementally with `go build` / `go vet`
/ `go test` after each phase, and the critical paths were validated against a **real sing-box
binary** built from source (with `with_v2ray_api`, `with_acme`, `with_quic`, `with_clash_api`,
`with_gvisor`, `with_utls`) using the actual `sing-box check` and a live gRPC stats endpoint.

## Reference research and spec corrections

`SagerNet/sing-box` (v1.14.0-alpha) and `MHSanaei/3x-ui` were shallow-cloned into a gitignored
`.refs/` folder and used to confirm exact config field names (`option/` structs), the Clash API
connection schema (`experimental/clashapi`), and the V2Ray stats service (`experimental/v2rayapi`).
Two findings corrected the original brief (both recorded in `backend-description.md`):

1. **sing-box reload is not connection-preserving.** SIGHUP / `systemctl reload` re-reads the config
   but resets active connections (SagerNet/sing-box#3731), unlike Xray/v2ray. The supported flow is
   `sing-box check` → reload, accepting a brief reset; an unvalidated config is never applied.
2. **Per-user traffic is not exposed by the Clash API** (the `/connections` JSON carries the inbound
   tag but no user identity) and **the V2Ray API is not built into official binaries** (`with_v2ray_api`
   only). Therefore the Clash REST API drives aggregate dashboard metrics, and an **opt-in V2Ray gRPC
   adapter** drives accurate per-user quota accounting.

A third detail surfaced during verification: sing-box renames its gRPC stats service in `init()` to
the standard `v2ray.core.app.stats.command.StatsService` (proto package is `experimental.v2rayapi`),
so that — not the proto package name — is the method path used on the wire.

## Confirmed product decisions

- **Traffic source:** both. Clash REST is the default (works with the stock binary); the V2Ray gRPC
  adapter is auto-selected when `stats.source = v2ray`.
- **Protocols (v1):** VLESS (TLS/Reality; tcp/ws/grpc), Naive, Hysteria2 — exactly what the frontend models.
- **TLS (v1):** panel cert files, self-signed-on-bare-IP, and Let's Encrypt (panel via autocert;
  inbounds via emitted `tls.acme`).
- **JSON contract:** new resource endpoints return **camelCase JSON matching the frontend mock types**
  so swapping the mock store for `fetch` is a drop-in. Existing auth endpoints keep their snake_case.

## What was built

### Data layer
- Domain models: `internal/domain/{inbound,client,setting,config_revision,metrics}.go`
  (`Protocol`, `Transmission`, `TLSMode`, `InboundSettings`, `ClientStatus`, `UserTraffic`,
  `TrafficDelta`, `SystemMetrics`, etc.).
- Migrations `000003_create_inbounds` … `000007_create_traffic_rollup` (inbounds, clients, settings,
  config_revisions, traffic_rollup) via the existing embedded migrator.
- SQLite repos: `inbound_repo`, `client_repo` (incl. batched `AddTraffic`, `CountByInbound`,
  `SumTraffic`), `setting_repo`, `config_revision_repo`, `traffic_repo` — following the existing
  raw-SQL + scan-helper pattern. Added a shared `rowScanner` helper and made `sqlite.New` create the
  DB's parent directory.

### Inbound / Client management
- `internal/services/inbound/service.go`: CRUD + validation (port range, protocol set, Reality
  requires VLESS + dest + sni, Naive requires TLS), server-side Reality x25519 keypair + short_id +
  `xtls-rprx-vision` flow (tcp-only), transport defaults (ws path / grpc service name), toggle, clone.
- `internal/services/client/service.go`: CRUD, credential generation (UUID for VLESS, password for
  Naive/Hysteria2), subscription-token generation, status/quota transitions.
- `internal/lib/keys/keys.go`: crypto/rand-based generators — Reality keypair (`base64.RawURLEncoding`
  of x25519, matching `sing-box generate reality-keypair`), short_id, UUIDv4, tokens, passwords.
- Handlers (`inbound_handler.go`, `client_handler.go`, `common.go`) return frontend-shaped DTOs;
  numeric IDs are serialized as strings to match the TS `string` id type. Reality private key is never
  exposed (only public key / short_id / flow surface).

### sing-box config generation + lifecycle (`internal/services/singbox`)
- `schema.go`: minimal typed structs whose JSON tags match the sing-box option structs.
- `generator.go`: renders `config.json` from the DB. Emits `log`, dynamic `inbounds` (VLESS/Naive/
  Hysteria2 with their users, TLS/Reality/ACME, ws/grpc transport), a `direct` outbound, a `route`
  with a `bittorrent → reject` rule, and `experimental.clash_api` (always) + `cache_file`.
  `experimental.v2ray_api` is emitted **only** when `stats.source = v2ray`. **Targets sing-box
  1.11–1.14:** the DNS block is intentionally omitted (no DNS form is valid across both the legacy
  1.11 and the new 1.14 schemas) and blocking uses the 1.11+ `reject` rule action instead of the
  removed `block` outbound. Inactive clients are excluded from emitted users.
- `checker.go`: `sing-box check -c <tmp>` wrapper with timeout; returns the core's stderr on failure.
- `process.go`: `ProcessManager` interface with **systemd** (`systemctl start|stop|restart|reload|
  is-active|show`) and **subprocess** (`sing-box run -c`, SIGHUP reload, SIGTERM stop, stdout/stderr →
  log buffer) adapters, auto-detected (systemd when the unit is installed, else subprocess). All host
  commands use argument arrays, never shell strings.
- `apply.go`: orchestrates render → temp write → `sing-box check` → atomic rename → reload (only if
  running) → record a `config_revision` (sha256, ok, error). Keeps the last 5 config backups. Exposes
  a **debounced `Trigger()`** (the `ConfigTrigger` consumed by the inbound/client services) and a
  synchronous `Apply()` (the `/api/core/reload` endpoint). A failed check leaves the live config
  untouched and returns the error to the UI.

### Subscriptions (`internal/services/sublink`)
- `builder.go`: `vless://`, `hysteria2://`, `naive+https://` URI builders (Reality params: pbk/sid/
  fp/flow/sni; ws path+host; grpc serviceName).
- `subscription.go`: assembles **base64 / plain / sing-box-JSON**; the JSON format renders a minimal
  client config whose proxy outbound mirrors the inbound.
- `subscription_handler.go`: public `GET /sub/{token}` and `/api/subscription/{token}` (token is the
  credential; returns 403 for disabled/expired clients), plus authed `GET /api/clients/{id}/links`.
  Host resolution: `inbound_host` setting → configured default → request Host.

### Stats + quota worker (`internal/services/stats`)
- `source.go`: `LiveSource` (aggregate dashboard metrics) and `UserSource` (per-user deltas)
  interfaces; `LiveHolder` storing the latest sample plus a 60-point throughput history.
- `clash.go`: Clash REST adapter; reads `/connections` and derives throughput from the delta of the
  cumulative totals (avoids the streaming `/traffic` websocket); reports online connection count.
- `v2ray.go`: V2Ray gRPC `StatsService.QueryStats` adapter (pattern `user>>>`, reset=true → deltas),
  parsing `user>>>NAME>>>traffic>>>uplink|downlink`. Implemented with a **hand-rolled protowire
  codec** (no protoc, no generated code); grpc + protobuf are the only added deps and are isolated to
  this file.
- `worker.go`: live-sampling loop (dashboard), expiry/quota enforcement loop (disables clients →
  triggers a debounced re-apply), and an optional per-user accounting loop (batched counter writes +
  daily rollup, active only when a `UserSource` is configured). `start_after_first_use` records
  `first_used_at` on first observed traffic.

### Dashboard, logs, system metrics
- `internal/services/sysstat`: Linux reader (`/proc/stat` CPU delta, `/proc/meminfo`, `/proc/uptime`,
  `syscall.Statfs`) with a build-tagged non-Linux stub (zeros) so dev builds work; zero dependencies.
- `internal/services/logbuf`: bounded ring buffer fed by both an **slog tee handler** (panel logs) and
  an `io.Writer` line splitter (core subprocess stdout), with level detection and substring/level
  filtering.
- `dashboard_handler.go`: assembles the full `Metrics` shape (sysstat + live throughput + DB counts +
  rollup + process status) and a traffic-history endpoint. `logs_handler.go`: recent lines with
  `level`/`q`/`limit` filters.

### TLS (`internal/services/tlsmgr`)
- Panel modes: `off`, `file` (load cert/key), `self_signed` (ECDSA P-256 x509 with IP SANs → HTTPS on
  a bare IP, cached to disk), `acme` (autocert, TLS-ALPN-01 for configured domains). Wired into the
  server via `TLSConfig` + `ListenAndServeTLS`.
- Inbound TLS material (ACME domain/email or cert/key paths) is configurable through the inbound API
  and emitted by the generator.

### Security
- `internal/transport/middleware/ratelimit.go`: per-IP token-bucket limiter. A general API limit
  (`auth.api_rate_limit`, default 100/s) plus a stricter login limit (`auth.login_rate_limit`,
  default 5/m) on `/api/auth/login*` for brute-force protection; over-limit → 429 with `Retry-After`.
- Auth middleware extended with public prefixes (`/swagger`, `/sub/`, `/api/subscription/`).

### Wiring, config, docs
- `cmd/main.go`: instantiates all repos/services, resolves sing-box paths to absolute (fixes a
  subprocess working-dir + relative `-c` path bug), runs the apply loop and stats worker on a
  cancellable root context (stopped on shutdown), tees logs into the ring buffer, and mounts the TLS +
  rate-limit middleware.
- `internal/config/config.go` + `config/dev.yaml`: added `stats` (source, v2ray address), `tls`
  (mode/cert/acme/self-signed), and `sing_box.process_mode` / `service_name`.
- `AGENTS.md` updated (endpoint table, project layout, middleware stack, dependencies); Swagger
  regenerated (`docs/`).

## Verification

- Generated configs for VLESS/Reality, VLESS/WS, Hysteria2, and Naive all **pass `sing-box check`**
  (sing-box 1.14-alpha).
- Full live run through the panel: create inbound + client → `/api/core/reload` (render + check +
  write) → `/api/core/start` → Clash API comes up → **real core logs captured** in the ring →
  dashboard reflects `coreRunning`/inbound/user/online counts → `/api/clients/{id}/links` and all three
  subscription formats work, and the **JSON subscription validates with `sing-box check`** → deleting a
  client regenerated the live config to 0 users (apply trigger).
- V2Ray gRPC codec validated **against the live sing-box stats endpoint** (after fixing the service
  path) and with a deterministic in-process fake-server unit test (alice up=1500/down=3000, bob up=42).
- Login brute-force limit confirmed (5×401 then 429).
- Self-signed HTTPS on a bare IP confirmed (cert SAN includes the IP; HTTPS `/health` returns ok).
- `go build ./...`, `go vet ./...`, `go test ./tests/...` all pass. New tests added: sublink builder,
  generator (Clash vs V2Ray emission, inactive-client exclusion), rate limiter, inbound validation,
  V2Ray codec. A skipped live test (`SING_GROK_LIVE_V2RAY`) is kept for manual integration checks.

## Dependencies added (minimal, isolated)

- `golang.org/x/crypto/acme/autocert` — panel Let's Encrypt (the `x/crypto` module was already present).
- `google.golang.org/grpc` + `google.golang.org/protobuf` — V2Ray stats adapter only.
- No new deps for: Reality keypair (`x/crypto/curve25519`), self-signed cert (`crypto/x509`), system
  metrics (`/proc` + `syscall`), Clash client (`net/http`), rate limiting, or logging.

## Known v1 limitations (intentional, documented)

- The DNS block is omitted for cross-version safety; per-version DNS customization is future work.
- Per-user traffic quotas require the opt-in V2Ray source; Clash-only deployments enforce expiry and
  show aggregate (not per-user) traffic.
- `start_after_first_use` is recorded but only basic deferred-start handling is implemented.
- A toolchain note: Go was not installed on the dev machine; Go 1.26 was installed via Homebrew to
  build/verify (the project targets Go 1.26 per CI).
