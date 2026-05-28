# Sing-Box Grok Web Panel Implementation Plan

## Summary

Build a local-first control panel for `sing-box` with FastAPI, SQLite, Next.js, Tailwind, Framer Motion, and systemd/Bash automation. The panel is not tied to any external infrastructure and controls only the local `sing-box` process through `127.0.0.1`.

## Repository Bootstrap

- Create a monorepo with `backend/`, `frontend/`, `scripts/`, `systemd/`, `.context/`, `README.md`, and `AGENTS.md`.
- Initialize git on `main`, add `origin` as `https://github.com/Web-Panel-Sing-Box/sing-box-web-panel.git`, commit the scaffold, and push when GitHub credentials are available.
- Keep `README.md` minimal until the implementation stabilizes.

## Backend

- Implement FastAPI with async SQLAlchemy and SQLite.
- Tables: `admins`, `users`, `inbounds`, `traffic_ledger`, `settings`, `config_revisions`, `audit_logs`, `active_ips`, `subscriptions`.
- Provide JWT cookie auth, Argon2id password hashing, login rate limiting, and audit logs.
- Implement CRUD for users and inbounds, subscription links, QR generation, log reads, dashboard metrics, and core process actions.
- Implement `LocalConfigGenerator` that reads SQLite, builds full `config.json`, validates with `sing-box check`, writes atomically, and records config revisions.
- Implement `ProcessManager` adapters for systemd and direct subprocess mode.
- Implement `TrafficBackgroundWorker` with live speed polling, per-user source adapters, quota enforcement, and batched SQLite writes.

## Frontend

- Implement Next.js App Router UI with true black background, mono typography, restrained panels, neon cyan/lime accents, Recharts, Framer Motion, and lucide icons.
- Screens: login, dashboard, metrics, clients table, inbounds list, core actions, logs, connection modal with QR and subscription links.
- Fetch API through same-origin `/api/*` rewrites to local FastAPI.

## Install and CLI

- `scripts/install.sh` detects OS/arch, installs prerequisites, downloads latest stable sing-box, creates local service user and runtime directories, builds backend/frontend, creates secrets, writes initial config, installs systemd units, and starts services.
- `scripts/sing-grok` provides terminal menu: start, stop, restart/reload, reset admin password, change panel port, status, logs.
- Systemd units run backend, frontend, and local sing-box under a dedicated `sing-grok` user.

## Security and Disk I/O

- sing-box Clash/V2Ray APIs must bind only to `127.0.0.1`.
- Generated secrets live in `/etc/sing-grok/*.env` with `0640` permissions.
- SQLite uses WAL, `synchronous=NORMAL`, `busy_timeout`, foreign keys, and batched traffic writes.
- Shell commands must avoid interpolating untrusted user input.

## Acceptance Checks

- Backend syntax and tests pass.
- Frontend typecheck/build pass.
- Shell scripts pass shellcheck.
- Installer can deploy on Ubuntu/Debian VPS with one command.
- Panel can create an inbound/client, generate a valid link/QR, validate and apply sing-box config, restart local core, stream logs, and enforce traffic limits.
