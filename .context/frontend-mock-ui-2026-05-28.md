# Frontend Mock UI — ChatGPT-style Sing-Box Panel

**Date:** 2026-05-28
**Scope:** Pure visual mock build of the `frontend/` app, now structured as a **SPA + micro-backend** architecture. The frontend is a Vite-built React 18 single-page application with no real network calls — all data is driven by an in-memory mock store. The Python FastAPI backend in `backend/` is the planned micro-backend that the SPA will eventually call through the `/api` proxy.

---

## 1. Original Plan (Phase 1)

The user asked for a maximally detailed plan to build a polished, animated frontend mock for the `sing-box` web panel modelled on `3x-ui`, in a ChatGPT-inspired visual language (deep dark gray, soft borders, Inter typography, controlled motion). The intent is to ship a presentation-quality UI for design sign-off before the real backend is wired.

### Decisions confirmed up front

- Routing: separate routes for `/`, `/inbounds`, `/clients`, `/settings`, `/logs`.
- Login: **skipped entirely** — the app boots straight into the dashboard.
- Existing code: **full replacement** of the legacy "true black + neon" scaffold.
- UI language: **English** for all labels, copy, and mock data.

### Design system (unchanged across the migration)

Color tokens (in `tailwind.config.ts`):

| Token | Value | Use |
|---|---|---|
| `canvas` | `#171717` | Sidebar / deepest background |
| `surface` | `#212121` | Main content background, active pill |
| `elevated` | `#2f2f2f` | Cards, inputs, modal body |
| `hover` | `#2a2a2a` | Card / menu hover |
| `border.subtle` | `rgba(255,255,255,0.08)` | Default 1px border |
| `brand` | `#10a37f` | Primary CTA, toggle on |
| `success` | `#19c37d` | Online / Active dot |
| `danger` | `#ef4444` | Stopped / error |
| `cyan` | `#22d3ee` | Chart incoming line |
| `violet` | `#a78bfa` | Chart outgoing line / accordion divider |
| `amber` | `#facc15` | Warn log lines |

Typography:
- **Inter** loaded from Google Fonts CDN (via `<link>` in `index.html`).
- **JetBrains Mono** from the same Google Fonts request, applied only to UUIDs, ports, tokens, IPs, log lines, and key blobs.
- Sentence case everywhere — no `ALL CAPS` running copy.

Motion tokens (`lib/motion.ts`):
- Page transition: `opacity 0→1`, `y +4 → 0`, 150ms ease-out, keyed on route.
- Modal: scale `0.95→1`, opacity `0→1`, 200ms ease-out, backdrop `bg-black/70 backdrop-blur-[8px]`.
- Dropdowns / accordions: short fade + slide (150–220ms).
- Buttons: `whileTap={{ scale: 0.98 }}` (no ripple).
- Numbers: custom `<AnimatedNumber>` running an rAF cubic-out interpolation, no jitter.
- Log lines: 50ms fade-in only.

---

## 2. Phase 1 Work — Initial Build (Next.js)

The first implementation was an App-Router Next.js project. It shipped:

- Full design-system primitives (`Button`, `Card`, `Input`, `Select`, `Toggle`, segmented `Tabs`, `Accordion`, `Modal`, `Toaster`, `Progress`, `StatusDot`, `AnimatedNumber`).
- Shell: `Sidebar` (hover-reveal + pin in `localStorage`, mobile drawer), `TopBar` (clickable status pill), `PageTransition`.
- Mock store driven by a 1-second `setInterval` with Brownian-motion metric drift, traffic ring buffer (60 points), and log generator.
- Five pages: Dashboard, Inbounds (with the full creation modal — three accordion sections, dice randomizer, Reality keypair), Clients (filter bar, table, detail modal, QR modal with deterministic SVG), Settings (four sectioned cards), Logs (filter bar, mono log viewer with 50 ms append fade).
- `next/font/google` loading Inter + JetBrains Mono.
- App Router route group `(panel)` with a shared client layout.

Verification at the end of Phase 1:
- `npm run typecheck` → passed.
- `npm run build` → passed, 8/8 static pages generated.

---

## 3. Phase 2 Work — Migration to Vite + React SPA

### User directive

> Make the whole project SPA (Single Page Application) + micro-backend architecture. And rewrite from Next.js to plain React: Vite + React + TypeScript + Tailwind CSS + Framer Motion + Recharts.

The Python FastAPI backend (`backend/`) is left untouched — it serves as the future micro-backend that the SPA will call through `/api/*`. The frontend lost its SSR/SSG layer and became a pure browser bundle.

### What was removed

- `frontend/app/` (App Router tree — `layout.tsx`, `globals.css`, `(panel)/*`).
- `frontend/next.config.mjs`, `frontend/next-env.d.ts`, `frontend/tsconfig.tsbuildinfo`, `frontend/.next/`, `frontend/package-lock.json`, `frontend/node_modules/`.
- `next`, `eslint-config-next`, `eslint` dependencies.

### What was added

- `frontend/vite.config.ts` — `@vitejs/plugin-react`, `@/` alias to `./src`, `server.port = 3000`, and a `/api` proxy to `process.env.SING_GROK_API_BASE_URL || "http://127.0.0.1:8081"` (matches the original Next.js rewrite).
- `frontend/index.html` — Vite entry, Google Fonts `<link>` for Inter + JetBrains Mono, theme color `#171717`.
- `frontend/tsconfig.json` — Vite-style, `"jsx": "react-jsx"`, `"moduleResolution": "bundler"`, `paths` alias `@/*`.
- `frontend/tsconfig.node.json` — separate project for `vite.config.ts`.
- `frontend/postcss.config.cjs` — renamed from `.js` because the package is now `"type": "module"`.
- `frontend/src/main.tsx` — `ReactDOM.createRoot(...).render(<App />)`.
- `frontend/src/App.tsx` — `BrowserRouter` + `Routes` + layout route with `<PanelLayout />` and five child routes, wrapped in `<Toaster>`.
- `frontend/src/index.css` — same tokens / scrollbars / `.glass` utility as before, but font-family now points at the Google-Fonts-loaded `Inter` and `JetBrains Mono` via CSS variables.
- `frontend/src/vite-env.d.ts` — `/// <reference types="vite/client" />`.
- `react-router-dom@^6.26` and `@vitejs/plugin-react@^4.3`, `vite@^5.4` in `package.json`.

### What moved

- `frontend/components/**` → `frontend/src/components/**`.
- `frontend/lib/**` → `frontend/src/lib/**`.
- `frontend/app/(panel)/page.tsx` → `frontend/src/pages/DashboardPage.tsx`.
- `frontend/app/(panel)/inbounds/page.tsx` → `frontend/src/pages/InboundsPage.tsx`.
- `frontend/app/(panel)/clients/page.tsx` → `frontend/src/pages/ClientsPage.tsx`.
- `frontend/app/(panel)/settings/page.tsx` → `frontend/src/pages/SettingsPage.tsx`.
- `frontend/app/(panel)/logs/page.tsx` → `frontend/src/pages/LogsPage.tsx`.
- `frontend/app/(panel)/layout.tsx` → `frontend/src/pages/PanelLayout.tsx` (now renders `<Outlet />`).
- `frontend/app/globals.css` → `frontend/src/index.css`.

### What was rewritten

- All `"use client";` directives stripped from `src/**/*.tsx` — React-DOM client mode is implicit in Vite SPAs.
- `import Link from "next/link"` + `<Link href=...>` → `import { Link } from "react-router-dom"` + `<Link to=...>` (in `sidebar.tsx` and `quick-links.tsx`).
- `import { usePathname } from "next/navigation"` + `const pathname = usePathname()` → `import { useLocation } from "react-router-dom"` + `const pathname = useLocation().pathname` (in `sidebar.tsx`, `topbar.tsx`, `page-transition.tsx`).
- `next/font/google` loaders removed; font CSS variables now come from `:root` in `index.css` and are populated by the Google Fonts stylesheet linked from `index.html`.
- Each page changed from `export default function XxxPage()` → `export function XxxPage()` so `App.tsx` named imports work cleanly.
- `PanelLayout.tsx` now uses `<Outlet />` from `react-router-dom` instead of receiving children from the Next.js App Router.
- `tailwind.config.ts` content globs updated from `app|components|lib` to `./index.html` + `./src/**/*.{ts,tsx}`.

### Final file layout

```
frontend/
├── index.html                       ← Vite entry + Google Fonts
├── package.json                     ← Vite/React/RR deps, "type": "module"
├── postcss.config.cjs               ← renamed for ESM compatibility
├── tailwind.config.ts               ← content: [index.html, src/**/*]
├── tsconfig.json                    ← Vite-style, "jsx": "react-jsx"
├── tsconfig.node.json               ← project for vite.config.ts
├── vite.config.ts                   ← @/ alias + /api proxy to FastAPI
└── src/
    ├── main.tsx                     ← createRoot bootstrap
    ├── App.tsx                      ← BrowserRouter + Routes
    ├── index.css                    ← Tailwind + global tokens
    ├── vite-env.d.ts
    ├── components/
    │   ├── shell/                   ← Sidebar (RR Link), TopBar, PageTransition (useLocation)
    │   ├── ui/                      ← 13 primitives (unchanged)
    │   ├── dashboard/
    │   ├── inbounds/
    │   ├── clients/
    │   └── logs/
    ├── lib/
    │   ├── mock/                    ← in-memory store + seed data
    │   ├── motion.ts                ← shared Framer Motion variants
    │   ├── format.ts                ← bytes, speed, percent, uptime, date, time
    │   └── utils.ts                 ← cn()
    └── pages/
        ├── PanelLayout.tsx          ← MockStoreProvider + Sidebar + TopBar + <Outlet />
        ├── DashboardPage.tsx
        ├── InboundsPage.tsx
        ├── ClientsPage.tsx
        ├── SettingsPage.tsx
        └── LogsPage.tsx
```

### Verification (Phase 2)

```
$ cd frontend && npm install
added 230 packages, audited 231 packages

$ cd frontend && npm run typecheck
> tsc --noEmit
(no output)

$ cd frontend && npm run build
> tsc --noEmit && vite build
vite v5.4.21 building for production...
✓ 2783 modules transformed.
dist/index.html                   0.83 kB │ gzip:   0.43 kB
dist/assets/index-CFnkak4M.css   24.24 kB │ gzip:   5.42 kB
dist/assets/index-CxFH4y1n.js   760.40 kB │ gzip: 224.07 kB
✓ built in 2.17s

$ cd frontend && npm run dev (smoke test)
VITE v5.4.21  ready in 104 ms
➜  Local:   http://127.0.0.1:3000/
GET / → HTTP/1.1 200 OK
```

### Deviations / notes

1. **Bundle size warning** — the production bundle is 760 kB (224 kB gzipped). This is dominated by Framer Motion + Recharts. Splitting via `manualChunks` is a follow-up; it does not affect the mock demo.
2. **`baseUrl` removed from tsconfig.json** — TypeScript 5.5 flags it as deprecated. The `paths` alias still resolves correctly because Vite's `resolve.alias` handles the bundling side and TS resolves `paths` relative to the tsconfig directory.
3. **PostCSS config extension** — switched from `.js` to `.cjs` so it can use `module.exports` even though `package.json` is now `"type": "module"`.
4. **No backend changes** — `backend/` (FastAPI) was not touched. The SPA's `/api` proxy is wired to `127.0.0.1:8081` for the day we connect the real micro-backend.
5. **Visual parity** — every component, animation, mock fixture, and behavior from Phase 1 is preserved. The migration was structural only.

---

## 4. Architecture summary (current)

- **Frontend (`frontend/`)** — Vite + React 18 + TypeScript SPA. Single bundle served from any static host (or proxied behind the FastAPI service in production). Routing is fully client-side via `react-router-dom`. State for the mock build lives in a single React Context (`MockStoreProvider`) driven by a 1-second ticker.
- **Backend (`backend/`)** — Python 3.11 / FastAPI micro-backend (unchanged in this work). The SPA's dev server proxies `/api/*` to `127.0.0.1:8081`, matching the planned production topology.

---

## 5. Where the plan lives

- `/Users/vadim_denisovich/.claude/plans/streamed-percolating-gray.md` — original Phase 1 plan, still the canonical design reference (color tokens, motion variants, page-by-page spec).
- This document — post-implementation record covering both Phase 1 (Next.js mock build) and Phase 2 (Vite + React SPA migration).
