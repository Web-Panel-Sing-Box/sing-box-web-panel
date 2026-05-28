# AGENTS.md

This repository implements a local-first web panel for `sing-box`. Keep changes scoped, secure by default, and aligned with the existing stack.

## Development Rules

- Backend code lives in `backend/app` and uses Python 3.11+, FastAPI, asyncio, SQLAlchemy Async, and SQLite.
- Frontend code lives in `frontend` and uses Next.js App Router, TypeScript, Tailwind CSS, Framer Motion, and Recharts.
- Do not bind sing-box Clash/V2Ray/management APIs to `0.0.0.0`; they must listen on `127.0.0.1`.
- Always validate generated sing-box configs with `sing-box check` before applying them.
- Do not log JWT secrets, admin passwords, sing-box API secrets, UUID lists, or generated private keys.
- Use subprocess argument arrays for host commands; do not build shell strings from user input.
- Keep SQLite writes batched for traffic updates. Avoid per-poll disk writes from background workers.
- Frontend style is true black, minimal, mono, and kinetic. Do not copy proprietary Grok or 3x-ui assets.

## Checks

- Backend syntax: `python -m compileall backend/app`
- Backend tests: `cd backend && pytest`
- Frontend typecheck: `cd frontend && npm run typecheck`
- Frontend build: `cd frontend && npm run build`
- Shell scripts: `shellcheck scripts/install.sh scripts/sing-grok`

## Git Safety

- Never stage unrelated user changes.
- Prefer explicit paths for `git add`.
- Keep generated runtime data, secrets, databases, logs, and build outputs out of git.
