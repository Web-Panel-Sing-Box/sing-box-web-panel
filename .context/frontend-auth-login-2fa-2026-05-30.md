# Frontend Login + 2FA (mock) — 2026-05-30

Added a client-side authentication gate to the `frontend/` SPA: a login screen
that must be passed before any panel route renders, plus an optional TOTP-style
two-factor step that the admin can enable from Settings. This is a **frontend
mock** consistent with the rest of the build (no Go/backend changes) — the
session lives in `localStorage` and is ready to swap for real `/api/auth/*`
later. Mock credentials are `admin` / `admin`; the 2FA code is `123456`.

## Round 1 — Login gate + 2FA (mock)

### Behaviour

- Visiting any panel route while logged out redirects to `/login`
  (`RequireAuth` guard + `<Navigate to="/login" state={{ from }}>`).
- `login(user, pass)`: wrong credentials → error toast, no navigation. Correct
  credentials with 2FA **off** → session starts, navigate to the originally
  requested route. Correct credentials with 2FA **on** → advance to the code
  step **without** starting a session (so `/login` can't be skipped).
- `verifyTwoFactor(code)`: `123456` → session starts → navigate; anything else
  → error toast, stays on the code step.
- Enabling 2FA from Settings → Security opens a setup modal (fake QR + secret +
  confirm code); the toggle only flips on once `123456` is confirmed. Disabling
  is immediate.
- Session persists across reloads; "Log out" in the sidebar clears the session
  and returns to `/login` (the 2FA-enabled flag persists).

### Auth model

`localStorage` keys reuse the existing `sing-grok:` namespace:
- `sing-grok:auth` — `"1"` when a session is active.
- `sing-grok:twofa` — `"on"` when 2FA is enabled.

Constants in `lib/auth.tsx`: `MOCK_USERNAME`/`MOCK_PASSWORD` = `admin`,
`MOCK_TOTP` = `123456`, `TWO_FACTOR_SECRET` = `JBSWY3DPEHPK3PXP`, plus a
`buildOtpAuthUri()` helper for the QR payload. Reads use the
`typeof window === "undefined"` defensive guard (mirrors `I18nProvider`).

### New files

- `frontend/src/lib/auth.tsx` — `AuthProvider` + `useAuth()` (context +
  `localStorage`), the mock credential/TOTP logic, and the demo constants.
- `frontend/src/components/auth/require-auth.tsx` — route guard
  (`isAuthenticated ? <Outlet/> : <Navigate to="/login" .../>`).
- `frontend/src/pages/LoginPage.tsx` — two-step (`credentials` → `twofactor`)
  state machine; lazy-loaded like the other routes; uses `m.form` (LazyMotion
  `strict`).
- `frontend/src/components/auth/two-factor-setup-modal.tsx` — the enable-2FA
  setup modal.
- `frontend/src/components/ui/fake-qr.tsx` — `FakeQrCode({ payload })`, extracted
  from `qr-modal.tsx` so the deterministic mock-QR logic lives in one place.

### Files touched

- `frontend/src/App.tsx` — wrapped the tree in `<AuthProvider>` (inside
  `<Toaster>`, around `<BrowserRouter>`); added the `/login` route and a root
  `<Suspense>` for the lazy `LoginPage`; wrapped the panel routes in
  `<RequireAuth>`.
- `frontend/src/components/clients/qr-modal.tsx` — uses the extracted
  `FakeQrCode` (deleted the local `deterministicQr`).
- `frontend/src/pages/SettingsPage.tsx` — the Security → Two-factor toggle is now
  backed by `useAuth()` and opens `TwoFactorSetupModal`.
- `frontend/src/components/shell/sidebar.tsx` — added a "Log out" control next to
  the GitHub link.
- `frontend/src/lib/i18n.tsx` — added `login.*`, `settings.twoFactor*`, and
  `nav.logout` keys to both `en` and `ru`.
- `frontend/src/test/test-utils.tsx` — `renderWithProviders` now wraps subjects
  in `<AuthProvider>`.
- `frontend/e2e/panel.smoke.spec.ts` — `test.beforeEach` seeds
  `localStorage["sing-grok:auth"] = "1"` so the guard lets the smoke routes
  render.

## Round 2 — UI trim (same day)

Removed explanatory chrome from the auth screens per the "titles only, no filler"
convention (now codified in `.context/frontend-rules.md`, section 9):

- **Login page (credentials step):** removed the "Sing box" brand text (icon
  only), the "Sign in" title + "Enter your credentials…" subtitle, and the
  `Username`/`Password` field labels. Fields now identify themselves with a
  `placeholder` + `aria-label`. Kept the icon, the two inputs, the Sign-in
  button, and the "Demo: admin / admin" hint.
- **2FA code step:** removed the title, subtitle, code label, "Back" button, and
  the "Demo code: 123456" hint. Kept the icon, the code input (mono placeholder),
  and the Verify button.
- **2FA setup modal:** dropped the header subtitle and the "Scan this QR code"
  caption above the QR. Kept the title, the QR, the manual-entry key box, the
  verification-code input, and a single primary button (removed the Cancel
  button — the header `×` still dismisses).
- Deleted the now-unused i18n keys from both dictionaries: `login.title`,
  `login.subtitle`, `login.twoFactorTitle`, `login.twoFactorSubtitle`,
  `login.back`, `login.codeHint`, `settings.twoFactorSetupSubtitle`,
  `settings.twoFactorScan`.

### Files touched (Round 2)

- `frontend/src/pages/LoginPage.tsx` — trimmed both steps; `Label` import dropped.
- `frontend/src/components/auth/two-factor-setup-modal.tsx` — title-only header,
  no QR caption, single footer button.
- `frontend/src/lib/i18n.tsx` — removed the 8 unused keys from `en` and `ru`.
- `.context/frontend-rules.md` — added the "Copy — titles only, no filler"
  convention under section 9.

## Verification

- `cd frontend && pnpm --ignore-workspace run typecheck` — clean.
- `cd frontend && pnpm --ignore-workspace run test` — 8 files / 10 tests passed.
- `cd frontend && pnpm --ignore-workspace run build` — built, `LoginPage` is its
  own ~2.86 kB chunk; `fake-qr` is shared.
- Browser (Claude Preview, port 3000):
  - Login gate redirect; `admin`/`admin` → dashboard (2FA off).
  - Settings → enable 2FA via setup modal (`123456`) → toggle on, `twofa="on"`.
  - Log out → `/login`; re-login now shows the 2FA step (no nav until the code is
    entered); wrong password and wrong code (`000000`) both rejected; correct
    code `123456` → panel; session survives a refresh.
  - Round-2 screens confirmed: login = icon + 2 inputs + button + demo hint;
    2FA step = icon + code input + Verify; setup modal = title + QR (no caption)
    + key box + code input + single Enable button.

## Known notes

- Frontend-only mock. The Go backend has no HTTP server yet; JWT cookie auth /
  Argon2id remain "Future" in `.context/sing-box-web-panel-plan.md`. The auth
  context is structured to swap `login`/`verifyTwoFactor`/`logout` for real
  `/api/auth/*` calls without touching the screens.
- The partial "credentials accepted, awaiting code" state lives only in
  `LoginPage` local state and is never persisted, so the 2FA step cannot be
  bypassed by editing `localStorage`.
