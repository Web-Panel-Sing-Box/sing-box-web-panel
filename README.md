# Sing Grok Web Panel

Status: Planning and early scaffold.

Sing Grok is a local-first web panel for managing a local `sing-box` process through `127.0.0.1`. It is designed to run on a user's own machine or VPS without a required external control plane.

## Stack

- Backend: Python 3.11+, FastAPI, asyncio, SQLite, SQLAlchemy Async.
- Frontend: Next.js App Router, TypeScript, Tailwind CSS, Framer Motion, Recharts.
- Core: local `sing-box` binary and local `config.json`.
- Ops: Bash installer, systemd units, local console menu.

## Safety Default

The panel must keep sing-box management APIs bound to `127.0.0.1`. Public exposure, if needed, belongs only to the web panel and must be configured explicitly by the operator.
