# Shilka · Sing-box Web Panel

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)
[![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](./LICENSE)

[Русский](./README_RU.md) | [English](./README.md)

---

### What is Shilka?

Shilka is a **local-first web panel** for managing a `sing-box` proxy server. It runs as a **single binary** with an embedded React frontend — no external web server, no Node runtime on the host, no Docker required. Built for VPS deployments with minimal resources (~30 MB RAM for the panel itself).

- Manage multiple **VLESS**, **Hysteria2** and **Naive** inbounds
- Provision clients with quotas, expiry, and individual subscription tokens
- Monitor traffic, system load, and sing-box core in real time
- TOTP 2FA with recovery codes for admin login
- Public subscription endpoint for automated client config import

---

### Features

| Area             | Details                                                                      |
| ---------------- | ---------------------------------------------------------------------------- |
| **Protocols**    | VLESS (Reality + Flow), Hysteria2 (H3 ALPN), Naive                           |
| **Transports**   | TCP, WebSocket, gRPC                                                         |
| **TLS**          | None, TLS, Reality (x25519 keypair) — per inbound                            |
| **Auth**         | Argon2id passwords, JWT, TOTP 2FA + recovery codes                           |
| **Traffic**      | Per-client byte counters, quota enforcement, expiry-based deactivation       |
| **Subscription** | Individual tokens, plain / base64 / JSON export, share-link QR on server     |
| **UI**           | Dark minimal panel, real-time dashboard, i18n (EN + RU), keyboard-accessible |
| **Config**       | YAML file + env variable overrides, validated at startup                     |

---

### Quick Start (VPS)

One command. Handles everything: user, directories, sing-box binary, TLS cert, systemd unit.

```bash
bash <(curl -Ls https://raw.githubusercontent.com/Web-Panel-Sing-Box/sing-box-web-panel/main/scripts/install.sh)
```

The script will ask:

- Domain or IP address
- Whether to use Let's Encrypt (acme.sh) — works for **domains** and **bare IPs**
- Panel port (random by default, 10000–65535)
- Web path prefix (random hex for obscurity)
- Admin username and password (auto-generated or custom)

After install, you get a `shilka` CLI command for management:

```bash
shilka           # interactive menu (start/stop/restart/reset-password/backup/etc.)
systemctl status shilka   # check service
journalctl -u shilka -f   # follow logs
```

---

### Build from Source

Requirements: **Go 1.26**, **Node 20+**, **pnpm**.

```bash
# 1. Build frontend
cd frontend
pnpm install && pnpm build

# 2. Embed frontend into Go binary
cd ..
rsync -a frontend/dist/ cmd/frontend/dist/

# 3. Build the single binary
go build -ldflags="-s -w" -o shilka ./cmd/

# 4. Run (dev config)
SHILKA_CONFIG_PATH=config/dev.yaml ./shilka
```

For cross-compilation (e.g., build for a VPS from macOS):

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o shilka-linux-amd64 ./cmd/
```

The resulting binary is fully self-contained — includes the frontend SPA, SQLite driver, and migrations.

---

### Configuration

Panel is configured via **YAML config** (`config/dev.yaml` or `config/prod.yaml`) + environment variable overrides. See `config/prod.yaml` for the production template.

Key sections:

```yaml
auth:
  jwt_secret: "" # required in production
  admin_user: "admin"
  admin_password: "" # required in production

http:
  address: ":443" # listen address; use ":" for all interfaces

tls:
  mode: "file" # off | file | self_signed | acme

sing_box:
  binary_path: "/opt/shilka/bin/sing-box"
  config_path: "/etc/shilka/config.json"

subscription:
  public_url: "https://panel.example.com"
```

Environment overrides (examples):

```bash
export SHILKA_AUTH_JWT_SECRET="your-secret"
export SHILKA_AUTH_ADMIN_PASSWORD="your-password"
export SHILKA_SING_BOX_API_SECRET="your-clash-secret"
```

The first admin is auto-seeded from these values. The sing-box config is generated from the database, validated with `sing-box check`, and applied atomically.

---

### Development

```bash
# Backend (API + Swagger)
go run ./cmd/
# http://127.0.0.1:8080
# Swagger: http://127.0.0.1:8080/swagger/

# Frontend (dev server, proxies /api to :8080)
cd frontend
pnpm install
pnpm dev
# Dev server at http://127.0.0.1:3000, proxies /api to :8080
```

### Build (embedded)

```bash
cd frontend && pnpm build
cd .. && rsync -a frontend/dist/ cmd/frontend/dist/
go build -o shilka ./cmd/
```

### Tests

```bash
go test ./tests/...           # Backend unit + integration
cd frontend && pnpm test      # Frontend unit
```

## Safety

sing-box management APIs (`Clash API`, `V2Ray API`) bind to `127.0.0.1`. Only the web panel port is exposed.

# http://127.0.0.1:3000

# Tests

go test ./tests/... # backend (11 packages)
cd frontend && pnpm test # frontend (8 files / 10 tests)

# Lint

cd frontend && pnpm typecheck
go vet ./...

```

---

### CLI Management Tools

After installing via [Quick Start](#quick-start-vps), or for a binary configured with the right paths, the `shilka` binary exposes these subcommands:

```

shilka run Start the server (used by systemd)
shilka admin reset-password Reset admin password interactively
shilka setting -port PORT Change the panel port
shilka setting -domain DOM Set domain and public URL
shilka api-token create Create an API token for nodes
shilka cert set-files Set custom TLS certificate files
shilka cert reset Disable panel TLS
shilka core reload Reload sing-box config

```

---

### Supported Protocols

| Protocol      | TLS Options        | Authentication      | Transport     |
| ------------- | ------------------ | ------------------- | ------------- |
| **VLESS**     | Reality, TLS, None | UUID                | TCP, WS, gRPC |
| **Hysteria2** | TLS (required)     | Password            | —             |
| **Naive**     | TLS (required)     | Username + Password | —             |

---

### Security

- All sing-box management APIs bind `127.0.0.1` — never exposed
- Passwords hashed with Argon2id (64 MiB memory, 3 iterations)
- TOTP 2FA with 8 one-time recovery codes (Argon2id-hashed)
- JWT with HS256, configurable expiry
- Per-IP login rate limiting (5/min) and general API rate limit (100/s)
- Request body capped at 16 KiB
- TLS config supports 4 modes: off, file (pre-generated certs), self-signed, acme (auto-cert + auto-renew)
```
