# Frontend Rules — `sing-grok-frontend`

Engineering rules for the Vite 8 + React 19 + TypeScript (strict) SPA in `frontend/`. Every new feature, bug fix, or refactor must comply. These rules exist because the codebase has already paid down the corresponding performance debt (see `.context/frontend-perf-optimization-2026-05-29.md` for the rationale). Violations are regressions.

When in doubt, choose the option that ships less JS, re-renders fewer components, and consumes a narrower slice of context.

---

## 1. State and Context

### 1.1 — Never reintroduce a broad `useStore()`

The mock store in `frontend/src/lib/mock/store.tsx` is split into six narrow contexts. Always consume the smallest slice you need.

| Need              | Hook                  |
|-------------------|-----------------------|
| Metrics / history | `useMetrics()`        |
| Inbounds list     | `useInbounds()`       |
| Clients list      | `useClients()`        |
| Logs              | `useLogs()`           |
| `{ paused }`      | `useRuntime()`        |
| Any mutator       | `useStoreActions()`   |

**Anti-pattern — do not write code like this:**

```tsx
// ❌ Reintroduces the cascade: any tick of metrics re-renders this component.
function MyPanel() {
  const everything = useEverything(); // or useStore()
  return <div>{everything.inbounds.length}</div>;
}
```

**Pattern:**

```tsx
// ✅ Subscribes only to inbounds; immune to metrics ticks.
function MyPanel() {
  const inbounds = useInbounds();
  return <div>{inbounds.length}</div>;
}
```

If you add a new piece of state that needs to live in the store, give it its own Context and hook. Do **not** widen an existing slice.

### 1.2 — Actions go through `useStoreActions()`

Mutators (`addInbound`, `toggleInbound`, `setPaused`, …) live in `ActionsContext`, which never re-renders consumers. Pull them via `useStoreActions()`, not through the data hooks.

### 1.3 — Don't `fetch` from components

Real `/api/*` traffic goes through the Vite proxy in `vite.config.ts`. Components must call into hooks (which call into the store / API client layer). Direct `fetch` from a component is forbidden.

---

## 2. Animation (Framer Motion)

### 2.1 — `m.*`, not `motion.*`

`App.tsx` wraps the tree in `<LazyMotion features={domMax} strict>`. In `strict` mode, using `motion.*` throws at runtime.

**Pattern:**

```tsx
import { m, AnimatePresence } from "framer-motion";

<m.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} />
```

`AnimatePresence` is unchanged — it's not a motion component.

### 2.2 — Variants and transitions live in `lib/motion.ts`

Reuse `pageVariants`, `modalVariants`, `backdropVariants`, `dropdownVariants`, `accordionVariants` from `frontend/src/lib/motion.ts`. Don't inline `Variants` blobs in components.

### 2.3 — `features={domMax}` is mandatory until `Segmented` drops `layoutId`

`components/ui/tabs.tsx` uses `<m.span layoutId="segmented-active">` (shared layout animation). `domAnimation` does not support layout animations and will silently break the sliding indicator. If `Segmented` is rewritten without `layoutId`, downgrade to `domAnimation` (saves ~10 KB gzip) — but until then, leave `domMax` alone.

---

## 3. Code splitting and lazy loading

### 3.1 — Pages are `React.lazy`

All route components in `App.tsx` are `lazy()`. New pages follow the same pattern:

```tsx
const NewPage = lazy(() =>
  import("@/pages/NewPage").then((m) => ({ default: m.NewPage }))
);
```

The single `<Suspense fallback={<RouteFallback />}>` in `PanelLayout.tsx` covers them.

### 3.2 — Heavy modals are lazy + gated + prefetched

Any modal larger than ~3 KB raw, or any modal that opens on user action, must:

1. Be `React.lazy` imported.
2. Render gated by `{ open ? <Suspense fallback={null}><Modal /></Suspense> : null }` so the chunk isn't even fetched until first open.
3. Prefetch on the trigger's `onMouseEnter` and `onFocus`.

**Pattern:**

```tsx
const HeavyModal = lazy(() =>
  import("@/components/foo/heavy-modal").then((m) => ({ default: m.HeavyModal }))
);

const prefetchHeavyModal = () => {
  void import("@/components/foo/heavy-modal");
};

function Page() {
  const modal = useDisclosure();
  return (
    <>
      <Button onClick={modal.open} onMouseEnter={prefetchHeavyModal} onFocus={prefetchHeavyModal}>
        Open
      </Button>
      {modal.isOpen ? (
        <Suspense fallback={null}>
          <HeavyModal open={modal.isOpen} onClose={modal.close} />
        </Suspense>
      ) : null}
    </>
  );
}
```

Skipping the prefetch makes the first open stall for 50–200 ms and clips the `AnimatePresence` enter animation.

### 3.3 — Recharts stays eager

Do not wrap Recharts components in `React.lazy` or render them directly under `<Suspense>`. `ResponsiveContainer` mis-sizes to 0×0 during the mount transition. The dashboard's charts must stay in the page chunk, not in a deeper lazy boundary.

If you need a fallback for a chart-heavy page, make sure the skeleton has the **same dimensions** as the real chart container (e.g. `h-72 w-full`). A zero-height skeleton triggers the same Recharts warning.

---

## 4. Hooks

### 4.1 — Reach for `frontend/src/hooks/` before inventing local state

| Hook                  | Use case                                              |
|-----------------------|-------------------------------------------------------|
| `useDisclosure()`     | Any modal / dropdown / drawer boolean state.         |
| `useCopyToClipboard()`| Copy buttons with a temporary "copied" indicator.    |
| `useLogFilter()`      | Filter `LogEntry[]` by query + level.                |
| `useClientFilter()`   | Filter `Client[]` by query + inbound + status.       |
| `useInboundForm()`    | The inbound create/edit/clone form state machine.    |

If your component grows three or more `useState`s **and** any `useEffect`/handler glue, extract a hook into `src/hooks/`. The rule of thumb: components contain JSX; hooks contain state machines.

### 4.2 — Random / crypto helpers live in `lib/random.ts`

Do not inline `crypto.randomUUID` / `crypto.getRandomValues` in components. Use `randomPort`, `randomHex`, `makeUuid` from `frontend/src/lib/random.ts`. If you need a new helper, add it there.

---

## 5. Memoization

Memoization in this codebase is precise, not reflexive. Add it only when it pays off.

### 5.1 — Hoist stable functions to module scope

If a function doesn't close over component state, define it at module level (not inside the component, not even via `useCallback`):

```tsx
// ✅ Stable across all renders, no allocation.
const formatPercentValue = (n: number) => `${n.toFixed(1)}%`;

export function CpuCard() {
  const { metrics } = useMetrics();
  return <AnimatedNumber value={metrics.cpu * 100} format={formatPercentValue} />;
}
```

### 5.2 — `useMemo` for derived data passed to children or charts

Wrap derivations of arrays/objects passed as props or `data` to Recharts:

```tsx
const spark = useMemo(
  () => history.slice(-20).map((p, i) => ({ i, v: p.down + p.up })),
  [history]
);
```

### 5.3 — `useCallback` for handlers passed to memoized children

Any handler passed to a `React.memo` child or a leaf in a `.map()` must be wrapped:

```tsx
const openEdit = useCallback((inbound: Inbound) => setModal({ open: true, mode: "edit", inbound }), []);
```

Don't blanket-wrap setters returned by `useState` — they're already stable.

### 5.4 — `React.memo` only on data-driven leaves with stable props

Apply `memo` when **all three** hold:
- The component renders inside a `.map()` or large list.
- Its props are primitives, stable references (`useCallback`), or stable objects (`useMemo`).
- It does not consume a frequently-ticking Context like `MetricsContext`.

Current examples: `InboundRow` (`inbound-row.tsx`), `Row` inside `clients-table.tsx`.

**Do not** apply `memo` to:
- Layout components (`Card`, `Modal`, `Accordion`) — their `children` prop always changes.
- Dashboard cards (`CpuCard`, `RamCard`, …) — they subscribe to ticking metrics anyway; the win was decoupling them from unrelated pages via context split, not memo.

---

## 6. Imports

### 6.1 — Destructured only

Whole-module imports are banned. They block tree-shaking.

```tsx
// ❌ Forbidden — pulls everything.
import Recharts from "recharts";
import * as Icons from "lucide-react";

// ✅ Required.
import { ResponsiveContainer, LineChart, Line, XAxis } from "recharts";
import { Copy, RefreshCw } from "lucide-react";
```

### 6.2 — Prefer per-file imports for big libraries

When a library exposes per-file entry points (e.g. `lodash/debounce`), use them:

```tsx
// ❌
import { debounce } from "lodash";

// ✅
import debounce from "lodash/debounce";
```

(`lodash` isn't currently a dependency; this rule applies if it ever is.)

### 6.3 — Import order

Follow the existing convention:
1. Node / framework / library imports.
2. `@/components/…`.
3. `@/lib/…`, `@/hooks/…`.
4. Local relative imports (`./foo`).

A blank line separates groups.

---

## 7. Bundle hygiene

### 7.1 — Use `pnpm build:analyze` to verify size changes

Before merging a change that adds a non-trivial dependency or a new heavy component, run:

```bash
cd frontend
pnpm build:analyze   # ANALYZE=true, opens dist/stats.html
```

Compare against `frontend/perf-snapshots/final-stats.html`. If your change adds more than ~5 KB gzip to the initial chunk, justify it in the PR description.

### 7.2 — Sourcemaps stay off in regular prod builds

`vite.config.ts` ties `sourcemap` to `ANALYZE`. Do not flip it on unconditionally — production sourcemaps inflate the dist by ~4 MB and leak internal structure.

### 7.3 — Don't add full polyfills or `core-js`

React 19 + modern browsers cover everything we need. If a feature requires a polyfill, gate it behind a `dynamic import` and load it only on browsers that need it (`@supports` / feature detect).

---

## 8. Testing

### 8.1 — Always render through `renderWithProviders`

`frontend/src/test/test-utils.tsx` wraps subjects in `LazyMotion`, `Suspense`, `MemoryRouter`, `I18nProvider`, `Toaster`, and `MockStoreProvider`. Direct `render()` from `@testing-library/react` will break for any component that uses `m.*`, `lazy()`, hooks like `useMetrics`, or i18n.

If you write a new helper, mirror the same wrapping. Test-only providers must mirror production providers.

### 8.2 — Lazy modals require `findBy*`

Because heavy modals are lazy in production, `getByText`/`getByRole` won't find them immediately. Use the async variants:

```tsx
await user.click(screen.getByText("berlin-edge-01"));
expect(await screen.findByText("Edit inbound connection")).toBeInTheDocument();
```

### 8.3 — Don't disable LazyMotion `strict` in tests

If a component breaks under `strict`, the production app also breaks. Fix the component (replace any remaining `motion.*` with `m.*`), don't loosen the test.

---

## 9. UI conventions (kept from `AGENTS.md`)

These are not performance rules but they affect rendering correctness. They are repeated here so a single rule file is sufficient for frontend work.

- **Sidebar.** Permanently a 64 px icons-only rail on `lg:`, slides as a 260 px drawer on smaller widths. No hover-expand, no pin.
- **Long tables (Clients, Inbounds).** Live inside a Card with `max-h-[calc(100dvh-NNNpx)]`, an inner `flex-1 overflow-auto`, and a `sticky top-0 z-10 bg-surface` header. The page never gets a second scrollbar.
- **Logs.** Wrapper is `flex h-[calc(100dvh-…)] flex-col`; the viewer is `flex-1 min-h-0`.
- **Modals.** Render through `components/ui/modal.tsx`. Flat (no header/footer dividers). Brand-green primary CTA.
- **Toggles.** Default to off unless the spec says otherwise.
- **Copy — titles only, no filler.** Do not pad the UI with explanatory subtitles, captions, or helper paragraphs. Page / section / modal headers are a bare title; do **not** add a descriptive line beneath them. Forms identify fields with a `Label` *or* an input `placeholder` (+ `aria-label`) — never a prose sentence telling the user what to type. `ModalHeader`'s `subtitle` prop stays unused unless the user explicitly asks for it. Default to the leanest text that is still unambiguous; every removed string is one less thing to translate in both `en` and `ru`.
- **Style tone.** True black, minimal, mono, kinetic. No Grok / 3x-ui copies.

---

## 10. Safety (sing-box specific, kept from `AGENTS.md`)

- sing-box Clash / V2Ray / management APIs must bind to `127.0.0.1`, never `0.0.0.0`.
- Generated sing-box configs must pass `sing-box check` before being applied.
- Never log JWT secrets, admin passwords, sing-box API secrets, UUID lists, or generated private keys — also applies if you add a frontend logger.
- Subscription URLs and QR payloads must never appear in console logs, even for debugging.

---

## 11. PR checklist

Before opening or merging a frontend PR, confirm:

- [ ] `pnpm typecheck` clean.
- [ ] `pnpm test` clean.
- [ ] No new `useStore()` (or equivalent broad hook).
- [ ] No new `motion.*` (must be `m.*`).
- [ ] No new whole-module imports (`import X from "lib"` for non-default-exporting libs, `import *`).
- [ ] Any new heavy modal is lazy + gated + has prefetch on its trigger.
- [ ] Any new route is in `App.tsx` under `React.lazy`.
- [ ] New stateful logic with ≥3 `useState`s lives in a hook under `frontend/src/hooks/`.
- [ ] Callbacks passed to `.map()` rows or `React.memo` children are `useCallback`-wrapped (or stable from a hook).
- [ ] If a new dependency is added, `pnpm build:analyze` was run and the delta is acceptable.
- [ ] No `fetch` in components; everything goes through the store / API client layer.
- [ ] No secrets, UUIDs, or subscription tokens logged to the console.

If any item fails, fix the code — do not lower the bar in this file.
