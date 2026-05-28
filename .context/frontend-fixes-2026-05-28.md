# Frontend Fixes Summary — 2026-05-28

## Implemented Changes

- Removed the global page title/core-status top bar and kept a compact mobile menu button.
- Reworked the Dashboard glass strip so core status, version, uptime, total sent, and total received fill the strip evenly.
- Moved the clickable core status pill into the Dashboard strip. Stopping an active core opens a confirmation modal; starting a stopped core happens immediately.
- Removed the Dashboard quick-link cards.
- Aligned Traffic with Inbounds and Clients cards, added sidebar-style icons to Inbounds and Clients, and made those cards navigable.
- Made Dashboard protocol chips link to filtered Inbounds and reveal their protocol color on hover.
- Fixed Memory and Disk card layout so swap is half-width and numeric labels do not overlap bars or directory labels.
- Centered collapsed sidebar icon highlights and pin/menu controls.
- Reworked Settings to use one page-level Save button and removed per-section Save buttons.
- Added persisted English/Russian i18n with English as the default language.
- Added success toast styling with a green textual check mark after successful-action messages.
- Changed custom dropdowns to render through a high-z-index portal so they appear above modals and clipped containers.
- Reworked Clients table column distribution, compact status badges, URL inbound filtering, and centered the "Start after first use" toggle.
- Reworked Inbounds table to open the edit modal directly, removed row expansion, added Transport, aligned the final columns, made client counts link to filtered Clients, and moved Clone/Delete into the modal with delete confirmation.
- Added Clone mode for inbounds that pre-fills settings and focuses the remark field with a copy suffix.
- Fixed a shellcheck warning in `scripts/install.sh` by replacing unquoted env expansion with an argument array.

## Tests And CI Added

- Added Vitest, Testing Library, jsdom, and focused unit/component/integration tests.
- Added Playwright smoke tests for Dashboard, Inbounds, Clients URL filtering, dropdown layering, and Russian language switching.
- Added GitHub Actions CI on pushes to `main` and pull requests into `main`.
- CI jobs cover backend compile/tests, frontend typecheck/tests/build/smoke, and shellcheck.

## Verification Commands Run

- `cd frontend && npm run typecheck`
- `cd frontend && npm test`
- `cd frontend && npm run build`
- `cd frontend && PLAYWRIGHT_CHANNEL=chrome npm run test:smoke`
- `python -m compileall backend/app`
- `uv --cache-dir /private/tmp/uv-cache run --project backend --extra dev pytest`
- `shellcheck scripts/install.sh scripts/sing-grok`

## Known Notes

- The default Playwright Chromium download was slow and timed out locally, so the local smoke run used installed Google Chrome through `PLAYWRIGHT_CHANNEL=chrome`. CI still installs Playwright Chromium explicitly.
- Vite still reports the existing large bundle warning for the Framer Motion/Recharts-heavy bundle.
- During smoke tests, Recharts logs a non-blocking container-size warning while pages initialize in headless Chrome; smoke assertions still pass.
