# Frontend Fixes Summary — 2026-05-29

## Round 2 — Late-evening polish (same day)

- Switched the `+ New configuration` and `+ Add client` CTAs from the primary green back to a white pill with `text-canvas` foreground so the buttons match the panel theme but stay highly contrastive.
- Fixed the dev-only bug where Tailwind 4 + Vite plugin would not regenerate `.text-canvas` on HMR. Wrapped all base resets (font-inherit on form controls, body color, etc.) inside `@layer base` and component utilities inside `@layer components` so Tailwind's `@layer utilities` rules win the cascade — the dark text on the white CTA now renders both in dev and in production.
- Added a dev-mode fallback `@layer utilities { .text-canvas { color: #171717 } .text-surface { color: #212121 } }` block inside `src/index.css` so brand-critical color classes always exist even before Tailwind finishes its first scan.
- Brought the Clients and Inbounds table cards closer to the bottom of the viewport: `max-h-[calc(100dvh-170px)]` / `max-h-[calc(100dvh-150px)]` (down from `-260px` / `-240px`). Dropped the `min-h-[…px]` floors entirely so the cards collapse to content height when filters return only a few rows.
- Stretched the Logs page further: page wrapper is now `h-[calc(100dvh-72px)] lg:h-[calc(100dvh-48px)] min-h-[420px]`, so on laptops and tablets the log viewer fills almost to the bottom edge.
- Added a `size` prop to `components/ui/toggle.tsx` (`sm | md | lg`). The "Start after first use" switch in both the inbound creation modal and the Add client modal now uses `size="lg"` (28×48 track, 24×24 knob).
- Removed the descriptive subtitles below the accordion section headers in the inbound creation/edit modal (Basics, Transport & security, User template) — only the bold section titles remain.
- Removed the descriptive subtitle below the inbound modal title (was "Configure protocol, transport, security, and a starter client." / "Change the remark before saving the cloned inbound.").
- Replaced the textual `Clone` and `Delete` buttons in the inbound edit modal footer with square icon-only buttons (`Copy` and `Trash2` from lucide). The Delete icon turns red on hover; both have proper `title` and `aria-label`.
- Added a GitHub repository link at the bottom-left of the sidebar. Uses a hand-rolled `GithubMark` SVG component (the installed `lucide-react@1.17.0` does not export `Github`). Opens `https://github.com/Web-Panel-Sing-Box/sing-box-web-panel` in a new tab with `rel="noopener noreferrer"`. The icon is collapsed-only on desktop (matching the icons-only rail) and shows the `GitHub` label in the mobile drawer.

## Files Touched (Round 2)

- `frontend/src/components/ui/button.tsx` — `white` variant restored to `bg-white text-canvas`.
- `frontend/src/pages/InboundsPage.tsx`, `frontend/src/pages/ClientsPage.tsx` — CTAs back to `variant="white"`.
- `frontend/src/index.css` — wrapped base resets in `@layer base`, component helpers in `@layer components`, added the dev-mode `text-canvas` / `text-surface` fallback under `@layer utilities`.
- `frontend/src/components/clients/clients-table.tsx`, `frontend/src/components/inbounds/inbounds-list.tsx` — tighter `max-h-[calc(100dvh-…)]`, no `min-h`.
- `frontend/src/pages/LogsPage.tsx` — viewer wrapper height calc tightened.
- `frontend/src/components/ui/toggle.tsx` — added `size` prop with `sm | md | lg` and per-size knob travel.
- `frontend/src/components/inbounds/inbound-form-modal.tsx` — removed `subtitle` from `ModalHeader`, dropped `description` from all three Accordion sections, switched Clone/Delete to icon buttons via a new local `IconActionButton`, and the User-template Start-after-first-use Toggle now uses `size="lg"`.
- `frontend/src/components/clients/add-client-modal.tsx` — Start-after-first-use Toggle uses `size="lg"`.
- `frontend/src/components/shell/sidebar.tsx` — added a `GithubMark` SVG component and the bottom GitHub link.

## Verification (Round 2)

- `cd frontend && npm run typecheck` — passed.
- `cd frontend && npm run build` — `2753 modules transformed`, 816 kB / 250 kB gzip.
- `cd frontend && npm run dev` — Vite boots on `http://127.0.0.1:3000/`.
- Chrome DevTools MCP at 1440×900 — Inbounds CTA renders as a white pill with a dark plus icon and "New configuration" text; Add client CTA likewise; Edit-inbound modal shows clean accordion titles, big Start-after-first-use toggle, icon-only Clone/Delete; GitHub icon visible at the bottom of the sidebar rail.

## Known Notes (Round 2)

- `lucide-react@1.17.0` (pinned by the user) does not export `Github`. We inlined a 24-line SVG instead of bumping the dep.
- Tailwind 4's HMR sometimes skips arbitrary value classes like `text-[#171717]` until a Vite restart. We avoid arbitrary color values for brand tokens; `text-canvas` (defined via `@config`) is the canonical class, plus the dev-mode fallback added to `index.css`.

---

## Round 1 — Earlier the same day

### Implemented Changes

- Reworked the Clients table column distribution so Name and Data usage are clustered on the left and Inbound and Status are clustered on the right. Expiry sits in a wide centre column with text-center alignment so its content visually floats away from both clusters. Inbound and Status cells use text-right and justify-end so they hug the right edge of the card.
- Added a `max-h-[calc(100dvh-260px)] min-h-[360px]` vertical scroll container to the Clients table card with a sticky `bg-surface` header row, so long client lists scroll inside the card without pushing the page below the fold.
- Added a `max-h-[calc(100dvh-240px)] min-h-[320px]` vertical scroll container to the Inbounds list card with a sticky `bg-surface` header row, matching the Clients pattern.
- Removed the leading `:` from the Inbounds port column so ports render as plain numbers (`44321` instead of `:44321`).
- Defaulted the "Start after first use" toggle to off in the Inbound creation modal, the Add client modal, and the mock store `addClient` action. The toggle is also reset to off whenever the Inbound modal reopens.
- Made the Logs page stretch to the visible window height. The page wrapper is now `h-[calc(100dvh-128px)] lg:h-[calc(100dvh-96px)] min-h-[420px] flex flex-col`, and the LogViewer is `flex-1 min-h-0` so the scrollable log area fills the remaining vertical space on laptops and tablets without leaving large empty space below.

### Files Touched

- `frontend/src/components/clients/clients-table.tsx` — new shared `GRID` constant, sticky header, scroll wrapper, centred Expiry, right-aligned Inbound/Status.
- `frontend/src/components/inbounds/inbounds-list.tsx` — scroll wrapper, sticky header.
- `frontend/src/components/inbounds/inbound-row.tsx` — dropped the `:` prefix from `{inbound.port}`.
- `frontend/src/components/inbounds/inbound-form-modal.tsx` — `startAfterFirstUse` default = `false`, also reset in the modal-open `useEffect`.
- `frontend/src/components/clients/add-client-modal.tsx` — `startAfterFirstUse` default = `false`, also reset on open.
- `frontend/src/lib/mock/store.tsx` — `addClient` falls back to `false` instead of `true` when `startAfterFirstUse` is omitted.
- `frontend/src/components/logs/log-viewer.tsx` — outer wrapper is `flex min-h-0 flex-1 overflow-hidden`; inner scroller is `min-h-0 w-full flex-1 overflow-y-auto`.
- `frontend/src/pages/LogsPage.tsx` — page wrapper now `h-[calc(100dvh-128px)] lg:h-[calc(100dvh-96px)] min-h-[420px] flex flex-col`.

### Verification Commands Run

- `cd frontend && npm run typecheck` — passed.
- `cd frontend && npm run build` — `2753 modules transformed`, 814 kB / 249 kB gzip, no errors.
- `cd frontend && npm run dev` — boots on `http://127.0.0.1:3000/`.
- Manual Chrome DevTools MCP walkthrough at 1440×900:
  - `/clients` — NAME and DATA USAGE clustered on the left, EXPIRY centred, INBOUND and STATUS clustered on the right; vertical scrollbar appears on the right edge of the card when the 24 mock clients overflow.
  - `/inbounds` — port column renders without colon (`44321`, `51005`, …); all 6 rows fit inside the new scroll container.
  - `/logs` — log viewer fills the remaining viewport height; live log lines stream in at the bottom.

### Known Notes

- Vite still reports the existing large-bundle warning for the Framer Motion / Recharts-heavy bundle. Unchanged by these fixes.
- The Clients table inner min-width was bumped from `min-w-[920px]` to `min-w-[960px]` to accommodate the new column distribution; on viewports narrower than 960px the table scrolls horizontally inside the card (vertical scroll continues to work).
- The Logs page calc uses `100dvh` (dynamic viewport) so the layout is stable when mobile/tablet browser chrome collapses on scroll. Falls back to standard viewport behaviour on browsers without `dvh` support.
