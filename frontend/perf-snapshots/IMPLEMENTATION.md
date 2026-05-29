# Frontend Performance Optimization — Implementation Report

Companion to [buzzing-launching-widget.md](buzzing-launching-widget.md) (the original plan, in Russian). This document records what was actually built, in English, for future contributors.

## Scope

Optimize the `sing-grok-frontend` SPA (Vite 8 + React 19 + TypeScript strict + pnpm + Tailwind 4 + Framer Motion 12 + Recharts 3 + react-router 7) along eight axes requested by the user:

1. Install a bundle visualizer to compare library sizes across builds.
2. Extract component logic into custom hooks.
3. Memoize functions.
4. Lazy-load heavy components.
5. Replace `motion` with `LazyMotion` (Framer Motion).
6. Audit imports — only destructured, no whole-library imports.
7. Wrap child-bound callbacks in `useCallback`.
8. Wrap stable-props chart components in `React.memo`.

The plan also identified a non-obvious blocker that had to be addressed before items 3, 7, 8 could pay off (see Step 0 below).

## Step 0 — Split `MockStoreProvider` into six contexts (the blocker)

**Problem.** `frontend/src/lib/mock/store.tsx` exposed a single `useStore()` returning a memoized `value = { metrics, history, inbounds, clients, logs, paused, ...actions }`. A `setInterval(1000ms)` mutates `metrics` and `history` every second, which re-creates `value`, which re-renders **every** consumer — including `InboundsPage` / `ClientsPage` / `LogsPage`, which never read metrics. Without this fix, `React.memo` and `useCallback` downstream are decorative.

**Change.** Split the single Context into six narrow contexts, each consumed via a dedicated hook:

| Context             | Hook                | Returns                       | Re-render frequency |
|---------------------|---------------------|-------------------------------|---------------------|
| `MetricsContext`    | `useMetrics()`      | `{ metrics, history }`        | Every 1 s (tick)    |
| `InboundsContext`   | `useInbounds()`     | `Inbound[]`                   | Only on mutation    |
| `ClientsContext`    | `useClients()`      | `Client[]`                    | Only on mutation    |
| `LogsContext`       | `useLogs()`         | `LogEntry[]`                  | ~every 3 s          |
| `RuntimeContext`    | `useRuntime()`      | `{ paused }`                  | Rarely              |
| `ActionsContext`    | `useStoreActions()` | All mutators (stable identity)| Never               |

`MockStoreProvider` now composes six nested providers in one file. The interval reads current `inbounds`/`clients` through `useRef`s, so adding/removing inbounds no longer reinstalls the interval. The legacy `useStore()` was removed and all 18 callers were migrated to the narrow hooks.

**Effect.** Only `useMetrics()` consumers (the four dashboard cards and `TrafficChart`) rerender on the tick. The other pages are no longer caught in the cascade.

## Step 1 — Bundle visualizer

- Added `rollup-plugin-visualizer@^7.0.1` as a devDependency.
- `frontend/vite.config.ts` conditionally enables it when `ANALYZE=true`, writing `dist/stats.html` (treemap, gzip + brotli sizes).
- Production sourcemaps moved behind the same flag (`sourcemap: analyze`). The previous default was `sourcemap: true`, which inflated the dist by ~4 MB and skewed size comparisons.
- New script `pnpm build:analyze` in `package.json` runs the full pipeline with `ANALYZE=true`.
- Baseline (`perf-snapshots/baseline-stats.html`) was captured before any other change.

## Step 2 — Custom hooks

Created `frontend/src/hooks/`:

- `useDisclosure(initial?)` — `{ isOpen, open, close, toggle }` for boolean state used by modals.
- `useCopyToClipboard(resetMs?)` — `{ copied, copy, reset }` wrapping `navigator.clipboard.writeText` with a temporary "copied" flag.
- `useLogFilter(logs, filter)` — memoized filter for the logs viewer.
- `useClientFilter(clients, filter)` — memoized filter for the clients table.
- `useInboundForm({ open, mode, inbound, onClose })` — extracts the 20+ `useState`s, the reset `useEffect`, and the save/delete/randomize handlers from the 517-line `inbound-form-modal.tsx`. Returns a flat object with values, setters, refs, and memoized callbacks. The modal now reads everything off `const f = useInboundForm(...)`.

Random utilities moved out of the modal into `frontend/src/lib/random.ts`: `randomPort()`, `randomHex(length)`, `makeUuid()`. They were inline-functions sitting in the lazy-loaded modal chunk; moving them keeps the helpers tree-shakable and reusable.

> `useAddClientForm` was deliberately skipped — that modal is small enough that a hook extraction would add boilerplate without a real win.

## Step 3 — Memoization

- `metric-card.tsx`: the `format` callback `(n) => \`${n.toFixed(1)}%\`` was instantiated three times per render across `CpuCard` / `RamCard` / `DiskCard`. Replaced with a single module-level constant `formatPercentValue`.
- `metric-card.tsx`: `TrafficSplitCard`'s `spark = history.slice(-20).map(...)` is now wrapped in `useMemo([history])`.
- `traffic-chart.tsx`: `fmtTime` / `fmtAxis` were already module-level; `ChartTooltip` is a stable function reference. Left as-is.
- `clients-table.tsx`: `inboundMap` and `rows` use `useMemo`. `rows` now flows through `useClientFilter`.
- `log-viewer.tsx`: `filtered` now flows through `useLogFilter`.

Layout components (`Card`, `Modal`, `Accordion`) intentionally not memoized — their `children` prop changes every render, defeating `memo`.

## Step 4 — Lazy loading

**Pages.** All five route components in `App.tsx` are wrapped in `React.lazy`:

```tsx
const DashboardPage = lazy(() => import('@/pages/DashboardPage').then(m => ({ default: m.DashboardPage })));
// + InboundsPage, ClientsPage, SettingsPage, LogsPage
```

A single `<Suspense fallback={<RouteFallback />}>` lives in `PanelLayout.tsx`, above `<Outlet />`, so the sidebar / topbar do not flash during route transitions. The fallback skeleton matches the expected card dimensions.

**Modals.** Four heavy modals load on demand:

- `inbound-form-modal.tsx` (~16 KB raw, 4.65 KB gzip) — lazy from `InboundsPage`.
- `add-client-modal.tsx` and `client-detail-modal.tsx` — lazy from `ClientsPage`.
- `qr-modal.tsx` already loads transitively inside `client-detail-modal.tsx` (acceptable; it lives in the same on-demand chunk).

Each lazy import is paired with a **prefetch on hover / focus** of its trigger button (e.g. `onMouseEnter={prefetchInboundForm}` on the "+ New inbound" CTA). Without this, the first open of a modal would visibly stall for 50–200 ms and clip the `AnimatePresence` enter animation.

Lazy modals are gated by `{ openState ? <Suspense fallback={null}><Modal /></Suspense> : null }` so unopened modals do not even mount the lazy chunk.

**Recharts** stays eager. The dashboard is the landing page and Recharts' `ResponsiveContainer` behaves poorly when mounted under `Suspense` (it can size to 0×0 for a frame).

## Step 5 — LazyMotion

- `App.tsx` wraps the tree in `<LazyMotion features={domMax} strict>`. `domMax` (not `domAnimation`) is required because `tabs.tsx` uses `motion.span layoutId="segmented-active"` — a shared layout animation.
- All 12 component files were rewritten from `import { motion } from 'framer-motion'` to `import { m } from 'framer-motion'`, and from `<motion.*>` to `<m.*>`. `AnimatePresence` is unchanged.
- `frontend/src/test/test-utils.tsx` was updated: `renderWithProviders` now wraps subjects in `<LazyMotion features={domMax} strict>` and a `<Suspense fallback={null}>`. Without this, components using `m.*` would fail under `strict`, and tests rendering lazy modals would not resolve.

## Step 6 — Import audit

A grep audit confirmed there were no whole-module imports (`import * as`, `import Foo from 'lodash'`, etc.) in `frontend/src/`. All `recharts`, `lucide-react`, and `framer-motion` imports were already destructured. No code change was needed for this item.

## Step 7 — `useCallback` on child-bound callbacks

- `InboundsPage.tsx`: `openCreate`, `openEdit`, `closeModal`, `openClone` are now `useCallback`s. `<InboundsList>` and `<InboundFormModal>` receive stable references.
- `ClientsPage.tsx`: `updateFilter`, `closeDetail` are `useCallback`s. `addModal` uses `useDisclosure` (returns stable callbacks).
- `inbound-row.tsx`: `open` and `onToggle` are `useCallback`s with proper dependency arrays.
- `clients-table.tsx`: `Row` receives `onSelect: (client) => void` and internally calls `useCallback(() => onSelect(client), [onSelect, client])` for the click handler, plus `useCallback`s for hover/leave.

## Step 8 — `React.memo`

- `InboundRow` in `inbound-row.tsx` is exported as `memo(InboundRowImpl)`. Combined with the `useCallback` `onEdit` from `InboundsPage`, rows skip re-render when their `inbound` reference is unchanged.
- The internal `Row` in `clients-table.tsx` is wrapped in `memo(function Row(...))`. The parent passes a stable `onSelect`, so memoization actually fires.

The four dashboard cards and `TrafficChart` are **intentionally not memoized**. Their props are derived from `useMetrics()`, which mutates every tick. Memoization would not save anything; the win is that they no longer drag unrelated pages along through the shared Context (Step 0).

## Test setup changes

`frontend/src/test/test-utils.tsx`:

- Adds `<LazyMotion features={domMax} strict>` and `<Suspense fallback={null}>` so production semantics are mirrored.
- This required two `InboundsPage.test.tsx` assertions to be switched from `getByText` to `findByText` / `findByRole`, because the modal is now async.

All 10 unit tests pass (`vitest run`).

## Files touched

**Modified.**

- `frontend/package.json` — added `rollup-plugin-visualizer` devDep + `build:analyze` script.
- `frontend/vite.config.ts` — conditional visualizer plugin + sourcemap toggle.
- `frontend/src/App.tsx` — `LazyMotion features={domMax} strict` + `React.lazy` for all 5 pages.
- `frontend/src/lib/mock/store.tsx` — six contexts replacing the single `useStore()`.
- `frontend/src/pages/PanelLayout.tsx` — `Suspense` with skeleton fallback above `<Outlet />`.
- `frontend/src/pages/InboundsPage.tsx`, `frontend/src/pages/ClientsPage.tsx` — `useCallback`, prefetch, lazy modals.
- `frontend/src/components/inbounds/inbound-form-modal.tsx` — slimmed down via `useInboundForm` + LazyMotion.
- `frontend/src/components/inbounds/inbound-row.tsx` — `memo` + `useCallback`s.
- `frontend/src/components/clients/clients-table.tsx`, `client-filter-bar.tsx`, `client-detail-modal.tsx`, `add-client-modal.tsx` — narrow-hook migration, memoized Row.
- `frontend/src/components/dashboard/{metric-card,traffic-chart,connections-panel,glass-strip,quick-links}.tsx` — narrow hooks + format constants + `useMemo` on derived arrays.
- `frontend/src/components/logs/log-viewer.tsx`, `log-filter-bar.tsx` — `useLogFilter` + narrow hooks.
- `frontend/src/components/shell/{topbar,sidebar,page-transition}.tsx` — narrow hooks + LazyMotion.
- All 12 framer-motion-using files (`motion.*` → `m.*`).
- `frontend/src/pages/InboundsPage.test.tsx` — async `findByText`.
- `frontend/src/test/test-utils.tsx` — `LazyMotion` + `Suspense`.

**Added.**

- `frontend/src/hooks/` — `useDisclosure.ts`, `useCopyToClipboard.ts`, `useLogFilter.ts`, `useClientFilter.ts`, `useInboundForm.ts`.
- `frontend/src/lib/random.ts` — `randomPort`, `randomHex`, `makeUuid`.
- `frontend/perf-snapshots/` — `baseline-stats.html`, `final-stats.html`, `dashboard-after.png`, `dashboard-cpu4x-after.json`, `report.html`, `report.json`, `README.md` (results table).

## Measurements

Production builds via `ANALYZE=true ./node_modules/.bin/vite build`.

| Metric                              | Baseline       | After          | Δ        |
|-------------------------------------|----------------|----------------|----------|
| Initial JS (raw)                    | 815.97 KB      | 424.68 KB      | **−48 %**|
| Initial JS (gzip)                   | 249.74 KB      | 136.85 KB      | **−45 %**|
| Initial chunks                      | 1              | code-split     | n/a      |
| First load on `/inbounds` (gzip)    | 249.74 KB      | ~139 KB        | **−44 %**|
| First load on `/settings` (gzip)    | 249.74 KB      | ~138 KB        | **−45 %**|
| First load on `/logs` (gzip)        | 249.74 KB      | ~138 KB        | **−45 %**|
| First load on `/` Dashboard (gzip)  | 249.74 KB      | 136.85 + 102.27 ≈ 239 KB (parallel) | −4 %, parallel |
| `inbound-form-modal` in initial     | yes            | no (4.65 KB gzip, lazy + prefetch) | extracted |
| `add-client-modal` in initial       | yes            | no (1.17 KB gzip, lazy) | extracted |
| `client-detail-modal` in initial    | yes            | no (2.28 KB gzip, lazy) | extracted |
| Sourcemap in prod                   | 4 MB           | 0 (only with `ANALYZE`) | −4 MB    |
| Vitest unit suite                   | 10/10          | 10/10          | parity   |
| LCP (DevTools trace, CPU 4×)        | n/a            | **465 ms**     | new metric |
| INP                                 | n/a            | **1 ms**       | new metric |
| CLS                                 | n/a            | **0.00**       | new metric |
| Lighthouse Accessibility / BP / SEO | n/a            | 96 / 100 / 82  | new metric |

The biggest wins:

1. **Three of the five pages load ~45 % less JS on first hit** thanks to route-level code splitting.
2. **`InboundsPage` / `ClientsPage` / `LogsPage` no longer rerender every second** because the metrics tick no longer reaches them. This is the qualitative improvement that translates directly into smoother interactions on low-end devices.
3. **Heavy modals are deferred** with prefetch on hover, so UX feels identical while shipping less to the initial render.

## How to reproduce the measurements

```bash
cd frontend
pnpm install
ANALYZE=true ./node_modules/.bin/vite build   # writes dist/stats.html
./node_modules/.bin/vite                      # dev server at http://127.0.0.1:3000/
./node_modules/.bin/vitest run                # 10 tests
```

For Lighthouse / CPU-throttled traces, open Chrome DevTools → Performance, enable CPU throttling 4×, and reload the dev page.
