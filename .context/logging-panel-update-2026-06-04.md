# Logging And Panel Update - 2026-06-04

## Summary

Implemented SIN-49 and SIN-36. The panel now has structured in-memory logs for backend, core, and frontend sources, plus request IDs, redacted fields, frontend runtime error ingestion, and a richer Logs UI. The dashboard now includes a panel version widget backed by protected update APIs and a root-owned update helper script for VPS installs.

## What Changed

- Extended backend logging to honor `logging.level`, `logging.format`, `logging.file_path`, file size, and backup settings.
- Added request ID generation in the HTTP logger, returned through `X-Request-ID` and stored in panel log entries.
- Extended `logbuf` entries with `source`, `requestId`, and redacted structured `fields`.
- Added a `core_log_path` tailer so sing-box file logs are mirrored into `/api/logs` as `source=core`, including the last existing lines on startup and appended lines afterward.
- Added `GET /api/logs?level=&source=&q=&limit=` and `POST /api/logs/frontend`.
- Added updater config and `GET /api/panel/version`, `POST /api/panel/update`.
- Added `scripts/update.sh` for locked, checksum-verified, atomic binary updates and systemd restart.
- Updated installer and uninstaller to manage `/usr/local/sbin/shilka-update` and `/etc/sudoers.d/shilka-update`.
- Added a dashboard panel version card and source-aware Logs UI filters.
- Added frontend runtime error reporting for authenticated sessions.
- Regenerated Swagger docs for the new API routes.

## Files Touched

- Backend: `cmd/main.go`, `internal/config/config.go`, `internal/services/logbuf/logbuf.go`, `internal/services/logbuf/tailer.go`, `internal/services/updater/service.go`, `internal/transport/handler/logs_handler.go`, `internal/transport/handler/panel_handler.go`, `internal/transport/middleware/logger.go`.
- Frontend: `frontend/src/api/logs.ts`, `frontend/src/api/panel.ts`, `frontend/src/components/dashboard/panel-version-card.tsx`, `frontend/src/components/logs/*`, `frontend/src/hooks/useLogFilter.ts`, `frontend/src/lib/i18n.tsx`, `frontend/src/lib/store.tsx`, dashboard and logs pages.
- Scripts/config/docs: `scripts/update.sh`, `scripts/install.sh`, `scripts/uninstall.sh`, `config/dev.yaml`, `config/prod.yaml`, `docs/*`, `go.sum`.
- Tests: new log buffer, updater, handler, and frontend widget/filter tests, plus request ID middleware coverage.

## Verification

- `go test ./tests/...`
- `go build ./...`
- `go vet ./...`
- `bash -n scripts/update.sh scripts/install.sh scripts/uninstall.sh`
- `cd frontend && pnpm test`
- `cd frontend && pnpm typecheck`
- `cd frontend && pnpm build`
- `go run -mod=mod github.com/swaggo/swag/cmd/swag init -g cmd/main.go -o docs --parseDependency --parseInternal`

Browser-level visual smoke was attempted with Vite on `127.0.0.1:3000`; the server started, but Playwright could not launch a browser in this environment because the bundled browser was missing and system Chrome exited with `SIGABRT`/`kill EPERM`.
