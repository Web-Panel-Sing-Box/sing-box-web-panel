# Frontend — Per-protocol inbound forms (vless / naive / hysteria2)

**Date:** 2026-06-01
**Branch:** `feat/per-protocol-inbounds`
**Scope:** Frontend only (`frontend/src/...`). No backend, no real `/api` wiring.

## Problem

The inbound create/edit modal rendered **one generic field set for every protocol**. As a
result the form showed fields that do not exist in sing-box for the chosen protocol, and
omitted required ones. Verified against the official sing-box docs:

- **Naive** (`/configuration/inbound/naive/`): no v2ray transport, no UUID, TLS is **always
  enabled**. Auth is `users[].username` + `users[].password`. Has `network` (tcp/udp) and
  `quic_congestion_control`. The old form wrongly showed a Transport dropdown, a UUID, and a
  TLS `None`/`TLS`/`Reality` selector.
- **Hysteria2** (`/configuration/inbound/hysteria2/`): QUIC-only — no transport, no Reality,
  TLS **required**. Auth is `users[].name` + `password`. Has `obfs` (salamander) and
  `up_mbps`/`down_mbps`. The old form wrongly showed a Transport dropdown, a UUID, and the
  full TLS selector.
- **VLESS** (`/configuration/inbound/vless/`): v2ray transports are `tcp`/`ws`/`grpc`/`http`/
  `httpupgrade`; `mKCP` and `XHTTP` are **Xray-only** and rejected by sing-box. The old form
  listed mKCP/XHTTP and was missing the `flow` field (`xtls-rprx-vision`).

## What changed

### `frontend/src/lib/mock/inbounds.ts` (model + options + seeds)
- Replaced the loose `Transmission` type with `VlessTransport = "tcp" | "ws" | "grpc" | "http"
  | "httpupgrade"`. Dropped `mkcp`/`xhttp` entirely.
- Added new types: `Network`, `Flow`, `ObfsType`, `QuicCc`.
- Extended `Inbound` with optional, protocol-scoped fields: `transport`, `flow` (vless);
  `username`, `password` (naive + hysteria2); `network`, `quicCc` (naive); `obfsType`,
  `obfsPassword`, `upMbps`, `downMbps` (hysteria2). Removed the always-present `transmission`.
- Replaced the single `TRANSMISSION_OPTIONS` with per-protocol option arrays:
  `VLESS_TRANSPORT_OPTIONS`, `FLOW_OPTIONS`, `NETWORK_OPTIONS`, `QUIC_CC_OPTIONS`,
  `OBFS_OPTIONS`. Added `DEFAULT_VLESS_TRANSPORT`, `DEFAULT_NETWORK`, `DEFAULT_QUIC_CC`.
- Rewrote `SEED_INBOUNDS` so every seed is spec-valid (hysteria2/naive seeds get TLS on +
  username/password; hysteria2 seeds get obfs + up/down mbps; naive seed gets network +
  quic_cc; vless seeds get transport + flow). Remarks preserved so existing tests still pass.

### `frontend/src/hooks/useInboundForm.ts` (state + save)
- Added state for all new fields. Added `regeneratePassword` / `regenerateObfsPassword`
  callbacks (reuse `randomHex` from `lib/random.ts`).
- `setProtocol` is now a wrapper that **normalizes invalid combinations**: switching to naive
  or hysteria2 forces `tls = "tls"` (helper `tlsForProtocol`) and generates a password if
  empty. The open-effect initializes every field from the edited inbound.
- `handleSave` builds a **protocol-specific payload** via `buildPayload()` — only the fields
  valid for the selected protocol are persisted (e.g. vless gets transport/flow/sni/dest;
  naive gets username/password/network/quicCc; hysteria2 gets username/password/obfs/up-down).

### `frontend/src/components/inbounds/inbound-form-modal.tsx` (conditional rendering)
- Transport dropdown + transport-specific sub-fields are **vless-only**. Naive shows a
  `Network` select; hysteria2 shows a read-only `Transport = QUIC` field.
- TLS: vless keeps the `None`/`TLS`/`Reality` segmented control; naive/hysteria2 show a locked
  "TLS is always enabled for this protocol" badge. SNI shows whenever `tls === "tls"`.
- User auth: vless → `User ID` + `UUID` (regenerate); naive/hysteria2 → `Username` + `Password`
  (regenerate).
- Added a Flow select (vless), a QUIC congestion-control select (naive), and an obfs +
  up/down-mbps block (hysteria2, obfs password shown only for `salamander`).

### `frontend/src/components/inbounds/inbound-row.tsx` (list display)
- `TransportChip` now takes the whole `inbound` and derives the label by protocol: vless →
  transport (TCP/WS/gRPC/HTTP-2/HTTPUpgrade), naive → network (TCP/UDP/TCP+UDP), hysteria2 →
  `QUIC`. Removed mKCP/XHTTP labels.

### `frontend/src/lib/i18n.tsx` (labels)
- Added keys `inbounds.transport|network|flow|quicCc|upMbps|downMbps|obfs|obfsPassword|
  tlsRequired|username|password` to **both** the `en` and `ru` dictionaries (`ru` is typed
  `Record<keyof typeof en, string>`, so a missing key fails `tsc`).

## Verification

- `pnpm --ignore-workspace run typecheck` → clean.
- `pnpm --ignore-workspace run test` → 8 files / 10 tests passing.
- `pnpm --ignore-workspace run build` → success.
- Preview (Vite dev server) + browser checks:
  - List chips render per protocol: `TCP · REALITY`, `WS · TLS`, `GRPC · REALITY` (vless),
    `QUIC · TLS` (hysteria2), `TCP+UDP · TLS` (naive).
  - New-inbound modal, naive: Network select (no transport dropdown), TLS-required badge,
    QUIC congestion control, username + password (no UUID).
  - hysteria2: read-only QUIC transport, TLS-required badge, Up/Down (Mbps), obfs,
    username + password; no Flow / Network / quic_cc.
  - vless: transport dropdown (tcp/ws/grpc/http/httpupgrade — no mKCP/XHTTP), Flow,
    None/TLS/Reality segmented, User ID + UUID.
  - Saved a vless inbound end-to-end → new `TCP · TLS` row added, no console errors.

## Out of scope

Backend changes, real API wiring, new protocols (vmess/trojan/tuic/…), and advanced
hysteria2 options (masquerade, realm, bbr_profile, ignore_client_bandwidth).
