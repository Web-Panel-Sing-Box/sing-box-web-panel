# Shilka – Sing-box Web Panel

Shilka is a local-first web panel for managing a local `sing-box` process. Single binary, single port, single systemd unit. Designed for VPS deployment with minimal resources.

## Stack

- **Backend**: Go 1.26, embedded SQLite (`modernc.org/sqlite`), JWT + TOTP 2FA, Argon2id passwords.
- **Frontend**: Vite + React 19 SPA, Tailwind CSS 4, Framer Motion, Recharts, embedded into binary.
- **Core**: local `sing-box` binary, config generated from DB, validated with `sing-box check`.
- **Ops**: one-command install script, single systemd unit.

## Quick Start (VPS)

```bash
curl -fsSL https://raw.githubusercontent.com/Web-Panel-Sing-Box/shilka-web-panel/main/scripts/install.sh | bash
```

## Local Development

### Backend

```bash
go run ./cmd
# API at http://127.0.0.1:8080
# Swagger at http://127.0.0.1:8080/swagger/
```

### Frontend

```bash
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

sing-box management APIs (`Clash API`, `V2Ray API`) bind  to `127.0.0.1`. Only the web panel port is exposed.
