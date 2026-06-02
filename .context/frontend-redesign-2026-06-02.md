# Frontend Redesign — Preserve & Elevate (2026-06-02)

Linear: SIN-23. PR: #6 (merged to `main`). Branch: `Vadim-Denisovich/redesign-frontend`.

## Goal

"Completely redesign" the panel frontend. Resolved with the user to **preserve the
existing charcoal/ChatGPT identity** (`#171717` canvas, `#212121` surface, `#2f2f2f`
elevated, brand `#10a37f`, Inter + JetBrains Mono) and **raise execution quality**,
not swap the palette. Applied the `design-taste-frontend` skill's discipline
(anti-default color/type/motion, audit-first, pre-flight: em-dash ban, contrast,
reduced-motion, theme lock) adapted to a dashboard; its landing-page/hero rules do
not apply. Sidebar was **locked** by the user (always-collapsed 64px rail, untouched).

## Process / workflow established

- Per-task branch `<github-username>/<task-slug>` off `main` (matches `4444urka/*`).
- A Linear issue per task in the `Sing-box-pannel` team, referenced in the PR.
- Both rules codified in `AGENTS.md` (`## Task Workflow`).

## Changes shipped

### Accessibility
- `App.tsx`: wrapped the tree in `<MotionConfig reducedMotion="user">`.
- `index.css`: global keyboard `:focus-visible` brand ring + `prefers-reduced-motion`
  CSS fallback.

### Shell
- A status `TopBar` was added then removed at the user's request; the shell keeps the
  floating mobile-menu button in `PanelLayout.tsx` and the locked sidebar. The
  dead/unused `topbar.tsx` was deleted (it had referenced non-existent `text-muted`
  classes).

### Core lifecycle (real, not faked)
- The start/stop control used to flip local React state only and show a fake "started"
  toast that the 3s metrics poll immediately reverted to "Stopped".
- Added store actions `startCore`/`stopCore` → `POST /api/core/start|stop` + refetch.
  `glass-strip.tsx` now `await`s them, surfaces the backend error on failure, disables
  the control while busy. Status reflects the backend poll.
- Removed the engine "plaque" (status dot + version) from the stop-confirm modal.
- Root cause of "core won't start" on dev machines: the **sing-box binary was not
  installed** (`exec: "sing-box": executable file not found in $PATH`). Installing it
  (`brew install sing-box`, or set `sing_box.binary_path`) makes the core run.

### Inbound creation fix
- The form defaulted to `protocol=naive + tls=none`, which the backend rejects
  (`naive requires tls`), and the error was swallowed (store `catch{}` + a
  fire-and-forget `addInbound`), so the UI showed a fake success while nothing
  persisted.
- `main` independently reworked the inbound form to be per-protocol and spec-correct
  (`buildPayload` forces `tls:"tls"` for naive/hysteria2; `setProtocol` normalizes TLS
  via `tlsForProtocol`). During the merge we kept main's version and re-added the one
  missing piece: a **save-failure error toast** in `useInboundForm.handleSave`
  (extract `ApiError.body.error`, fallback to `inbounds.saveFailed`).

### Copy / i18n
- Removed every em-dash from user-visible output (`format.ts` invalid-date fallback,
  `clients-table` inbound fallback, `inbound-form-modal` key placeholders).
- Cleaned stale "mock"/"demo"/"build" strings in `i18n.tsx` (en + ru): core stop body,
  `settings.readOnly`, `settings.changePasswordHint`, `inbounds.deleteBody`,
  `settings.description`; dropped the "Demo: admin / admin" login hint usage.
- Added keys: `inbounds.saveFailed`, `core.startFailed`, `core.stopFailed` (en + ru).

### Other
- Fixed Disk card `NaN%` (divide-by-zero on empty `diskSegments`).
- De-capped table column headers (charcoal spec forbids ALL-CAPS for non-technical
  text); protocol/transport chips and log levels stay uppercase.
- Docs: `AGENTS.md` task workflow + corrected dev run command
  (`go run ./cmd`, not `./cmd/main.go` which skips `embed.go`); "true black" → charcoal;
  `frontend-rules.md` mock-store path → real `lib/store.tsx`.

## Merge notes (`main` had diverged 16 commits)

`main` advanced (superpanel nodes + CLI, per-protocol inbounds, `HashRouter`,
install.sh rework) while the redesign was based on an older tree. Conflict resolution:
- `App.tsx`: keep `MotionConfig` (a11y) + main's `HashRouter` + `NodesPage` route.
- `useInboundForm.ts` / `inbound-form-modal.tsx`: take main's per-protocol form, re-add
  the save-error toast.
- All other overlaps auto-merged (i18n/store kept both my keys/actions and main's new
  ones).

## Verification

`pnpm typecheck` clean · `pnpm test` (10) pass · `pnpm build` clean · `go build ./...`
OK. Verified in-browser (Go backend + Vite, admin/admin): all pages render, no console
errors, inbound create persists, core starts to **Active** once sing-box is installed.
CI on PR #6: backend / frontend / shell checks all green.

## Follow-ups / not done

- E2E (Playwright) selectors not exercised this round.
- `lib/format.ts` and `lib/utils.ts` both define `formatBytes`/`formatSpeed`
  (duplication, different unit labels) — candidate for consolidation.
- `dashboard.trafficSubtitle` is still a descriptive subtitle (mild deviation from the
  "titles only" copy rule), left as a chart axis hint.
