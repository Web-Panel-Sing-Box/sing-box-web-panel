# Frontend Performance Optimization — 2026-05-29

Targeted optimization pass on `sing-grok-frontend` (Vite 8 + React 19 + TS strict + pnpm + Tailwind 4 + Framer Motion 12 + Recharts 3 + react-router 7). Goal: shrink initial JS, eliminate per-second cascading re-renders, and add a permanent bundle-comparison workflow.

Original request (in Russian) covered eight items: install a bundle visualizer, extract hooks, memoize functions, lazy-load heavy components, switch `motion` → `LazyMotion`, audit imports, wrap callbacks in `useCallback`, and apply `React.memo` to chart components. A non-obvious blocker (single Context with ticking metrics) was identified during planning and addressed as Step 0.

## Step 0 — Split `MockStoreProvider` into six contexts (the blocker)

**Why.** `frontend/src/lib/mock/store.tsx` exposed a single `useStore()` returning a memoized `value = { metrics, history, inbounds, clients, logs, paused, ...actions }`. A 1-second `setInterval` mutates `metrics`/`history`, which re-creates `value`, which re-renders every consumer — including `InboundsPage` / `ClientsPage` / `LogsPage`, which never read metrics. Without fixing this, `React.memo` and `useCallback` downstream are decorative.

**Change.** Six narrow contexts, each with a dedicated hook:

| Context             | Hook                | Returns                       | Re-render frequency |
|---------------------|---------------------|-------------------------------|---------------------|
| `MetricsContext`    | `useMetrics()`      | `{ metrics, history }`        | Every 1 s (tick)    |
| `InboundsContext`   | `useInbounds()`     | `Inbound[]`                   | Only on mutation    |
| `ClientsContext`    | `useClients()`      | `Client[]`                    | Only on mutation    |
| `LogsContext`       | `useLogs()`         | `LogEntry[]`                  | ~every 3 s          |
| `RuntimeContext`    | `useRuntime()`      | `{ paused }`                  | Rarely              |
| `ActionsContext`    | `useStoreActions()` | All mutators (stable identity)| Never               |

`MockStoreProvider` composes the six providers in one file. The tick interval reads current `inbounds`/`clients` through `useRef`s, so adding/removing inbounds no longer reinstalls it. The legacy `useStore()` was removed; all 18 callers were migrated to the narrow hooks.

**Effect.** Only `useMetrics()` consumers (four dashboard cards + `TrafficChart`) re-render on the tick. Inbounds / Clients / Logs pages are detached from the cascade.

## Step 1 — Bundle visualizer + sourcemap cleanup

- Added `rollup-plugin-visualizer@^7.0.1` as a devDependency.
- `frontend/vite.config.ts` conditionally enables the plugin when `ANALYZE=true`, writing `dist/stats.html` (treemap, gzip + brotli).
- Production sourcemaps moved behind the same flag (`sourcemap: analyze`). The previous default (`sourcemap: true`) added ~4 MB to the dist and skewed size comparisons.
- New script `pnpm build:analyze` in `package.json` runs the full pipeline with `ANALYZE=true`.
- Baseline (`frontend/perf-snapshots/baseline-stats.html`) was captured before any other change.

## Step 2 — Custom hooks

Created `frontend/src/hooks/`:

- `useDisclosure(initial?)` — `{ isOpen, open, close, toggle }` for modal boolean state.
- `useCopyToClipboard(resetMs?)` — `{ copied, copy, reset }` around `navigator.clipboard.writeText`.
- `useLogFilter(logs, filter)` — memoized log filter.
- `useClientFilter(clients, filter)` — memoized client filter.
- `useInboundForm({ open, mode, inbound, onClose })` — extracts the 20+ `useState`s, the reset `useEffect`, and save/delete/randomize handlers from the 517-line `inbound-form-modal.tsx`. Returns a flat object with values, setters, refs, and memoized callbacks; modal now reads everything off `const f = useInboundForm(...)`.

Random utilities moved to `frontend/src/lib/random.ts`: `randomPort()`, `randomHex(length)`, `makeUuid()`.

`useAddClientForm` was deliberately skipped — the modal is small enough that a hook extraction would add boilerplate without a real win.

## Step 3 — Memoization

- `metric-card.tsx`: the `format` callback `(n) => \`${n.toFixed(1)}%\`` was instantiated three times per render across `CpuCard` / `RamCard` / `DiskCard`. Replaced with a single module-level constant `formatPercentValue`.
- `metric-card.tsx`: `TrafficSplitCard`'s `spark = history.slice(-20).map(...)` is wrapped in `useMemo([history])`.
- `traffic-chart.tsx`: `fmtTime` / `fmtAxis` were already module-level; `ChartTooltip` is a stable function reference. Left as-is.
- `clients-table.tsx`: `inboundMap` keeps its `useMemo`; `rows` flows through `useClientFilter`.
- `log-viewer.tsx`: `filtered` flows through `useLogFilter`.

Layout components (`Card`, `Modal`, `Accordion`) intentionally not memoized — their `children` prop changes every render.

## Step 4 — Lazy loading

**Pages.** All five route components in `App.tsx` are wrapped in `React.lazy`:

```tsx
const DashboardPage = lazy(() => import('@/pages/DashboardPage').then(m => ({ default: m.DashboardPage })));
// + InboundsPage, ClientsPage, SettingsPage, LogsPage
```

A single `<Suspense fallback={<RouteFallback />}>` lives in `PanelLayout.tsx`, above `<Outlet />`, so the sidebar / topbar don't flash during route transitions. The fallback skeleton matches the expected card dimensions.

**Modals.** Four heavy modals load on demand:

- `inbound-form-modal.tsx` (~16 KB raw, 4.65 KB gzip) — lazy from `InboundsPage`.
- `add-client-modal.tsx` and `client-detail-modal.tsx` — lazy from `ClientsPage`.
- `qr-modal.tsx` rides inside `client-detail-modal.tsx`'s chunk (acceptable; not in the initial bundle).

Each lazy import is paired with a **prefetch on hover / focus** of its trigger button (e.g. `onMouseEnter={prefetchInboundForm}` on the "+ New inbound" CTA). Without prefetch, the first modal open would visibly stall for 50–200 ms and clip the `AnimatePresence` enter animation.

Lazy modals are gated by `{ openState ? <Suspense fallback={null}><Modal /></Suspense> : null }` so unopened modals don't mount the lazy chunk.

**Recharts** stays eager. Dashboard is the landing page and Recharts' `ResponsiveContainer` misbehaves under Suspense (sizes to 0×0 for a frame).

## Step 5 — LazyMotion

- `App.tsx` wraps the tree in `<LazyMotion features={domMax} strict>`. `domMax` (not `domAnimation`) is required because `tabs.tsx` uses `motion.span layoutId="segmented-active"` — a shared layout animation.
- All 12 component files rewritten from `import { motion } from 'framer-motion'` to `import { m } from 'framer-motion'`, and `<motion.*>` → `<m.*>`. `AnimatePresence` unchanged.
- `frontend/src/test/test-utils.tsx` updated: `renderWithProviders` wraps subjects in `<LazyMotion features={domMax} strict>` + `<Suspense fallback={null}>`. Mirror this in any new render helper.

## Step 6 — Import audit

A grep audit confirmed there were no whole-module imports (`import * as`, `import Foo from 'lodash'`, etc.) in `frontend/src/`. All `recharts`, `lucide-react`, and `framer-motion` imports were already destructured. No code change needed.

## Step 7 — `useCallback` on child-bound callbacks

- `InboundsPage.tsx`: `openCreate`, `openEdit`, `closeModal`, `openClone` are `useCallback`s. `<InboundsList>` and `<InboundFormModal>` receive stable references.
- `ClientsPage.tsx`: `updateFilter`, `closeDetail` are `useCallback`s. `addModal` uses `useDisclosure` (already stable).
- `inbound-row.tsx`: `open` and `onToggle` are `useCallback`s with proper dependency arrays.
- `clients-table.tsx`: `Row` receives `onSelect: (client) => void` and internally calls `useCallback(() => onSelect(client), [onSelect, client])`, plus `useCallback`s for hover/leave.

## Step 8 — `React.memo`

- `InboundRow` in `inbound-row.tsx` is exported as `memo(InboundRowImpl)`. Combined with `useCallback` `onEdit` from `InboundsPage`, rows skip re-render when their `inbound` reference is unchanged.
- The internal `Row` in `clients-table.tsx` is wrapped in `memo(function Row(...))`. The parent passes a stable `onSelect`, so memoization actually fires.

Dashboard cards and `TrafficChart` are **intentionally not memoized**. Their props derive from `useMetrics()`, which mutates every tick. The win was decoupling them from unrelated pages (Step 0), not skipping their own renders.

## Files Touched

**Modified.**

- `frontend/package.json` — added `rollup-plugin-visualizer` devDep + `build:analyze` script.
- `frontend/vite.config.ts` — conditional visualizer plugin + sourcemap toggle.
- `frontend/src/App.tsx` — `LazyMotion features={domMax} strict` + `React.lazy` for 5 pages.
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
- `frontend/src/pages/InboundsPage.test.tsx` — async `findByText` for lazy modal.
- `frontend/src/test/test-utils.tsx` — `LazyMotion` + `Suspense`.
- `AGENTS.md` — stack updated to React 19, new Frontend Performance Rules section, pnpm commands.

**Added.**

- `frontend/src/hooks/` — `useDisclosure.ts`, `useCopyToClipboard.ts`, `useLogFilter.ts`, `useClientFilter.ts`, `useInboundForm.ts`.
- `frontend/src/lib/random.ts` — `randomPort`, `randomHex`, `makeUuid`.
- `frontend/perf-snapshots/` — `baseline-stats.html`, `final-stats.html`, `dashboard-after.png`, `dashboard-cpu4x-after.json`, `report.html`, `report.json`, `README.md`, `IMPLEMENTATION.md`.

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
2. **`InboundsPage` / `ClientsPage` / `LogsPage` no longer re-render every second** because the metrics tick no longer reaches them. This is the qualitative improvement that translates directly into smoother interactions on low-end devices.
3. **Heavy modals are deferred** with prefetch on hover, so UX feels identical while shipping less to the initial render.

## Verification

- `pnpm typecheck` — clean.
- `pnpm test` — 10/10 passed (after wrapping test renderer in `LazyMotion` + `Suspense`, and switching two `getByText` to `findByText` for the now-async `inbound-form-modal`).
- `pnpm build:analyze` — `dist/stats.html` generated; visualizer treemap confirms route chunks.
- Chrome DevTools at 1440×900, CPU throttling 4× — dashboard renders cleanly, no LayoutShift, charts populate within one frame after Suspense resolves (skeleton placeholder sized to match preserves layout).

## Known Notes

- `pnpm` from inside `frontend/` may require `--ignore-workspace` because the root `pnpm-workspace.yaml` does not declare packages (only `allowBuilds`/`minimumReleaseAgeExclude`). Documented in `AGENTS.md`.
- `LazyMotion` uses `features={domMax}` (not the smaller `domAnimation`) because `tabs.tsx` uses `layoutId`. Switching to `domAnimation` would break the `Segmented` sliding indicator. If `Segmented` is ever rewritten without `layoutId`, downgrade to `domAnimation` to save ~10 KB gzip.
- `Recharts` console warns once at first paint (`width(-1) and height(-1) of chart should be greater than 0`). It's the `ResponsiveContainer` measuring during the Suspense → mount transition. Charts render correctly on the next frame. Not worth chasing.
- `frontend/perf-snapshots/baseline-stats.html` and `final-stats.html` are committed (~1.2 MB each) so future contributors can diff visually. If repo size becomes a concern, add them to `.gitignore` and regenerate via `pnpm build:analyze`.

## Commit

- `e48d5dc` — `perf(frontend): split store contexts, lazy-load routes + modals, switch to LazyMotion` on `main`.

---
