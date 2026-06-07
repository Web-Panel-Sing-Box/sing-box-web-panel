# Backend Hysteria2 Sublink Fix (SIN-67) - 2026-06-07

## Summary

SIN-67 reported that the Hysteria2 protocol did not work. Linear issue context and the linked GitHub issue did not include reproduction details, so the investigation used Graphify output, local code paths, and the live RF servers.

The live node already had an active Hysteria2 inbound and sing-box was able to start it. A temporary loopback sing-box client on the node successfully proxied an HTTPS request through that Hysteria2 inbound. This narrowed the issue away from core server config generation and toward exported client material.

## What Changed

- Hysteria2 share links now include official URI obfuscation parameters when the inbound uses obfs:
  - `obfs=salamander`
  - `obfs-password=<password>`
- Hysteria2 JSON subscriptions now mirror protocol options from the inbound into the client outbound:
  - `up_mbps`
  - `down_mbps`
  - `network`
  - `obfs`
  - `bbr_profile`
  - `brutal_debug`
- Added focused tests for Hysteria2 link and JSON subscription output.

## Why

The server-side Hysteria2 config supported obfs, network, bandwidth, BBR profile, and debug options, but `internal/services/sublink` only emitted password, server, port, and TLS fields for clients. If a Hysteria2 inbound was configured with obfs or transport-specific options, the exported client config no longer matched the server and could fail even though the server listener itself was healthy.

Official Hysteria2 URI docs support `obfs` and `obfs-password` in share links. Official sing-box Hysteria2 outbound docs support the mirrored JSON fields added here.

## Files Touched

- `internal/services/sublink/builder.go`
  - Adds Hysteria2 obfs query parameters to share links.
- `internal/services/sublink/subscription.go`
  - Adds Hysteria2 outbound fields and obfs struct for JSON subscriptions.
- `tests/services/sublink/builder_test.go`
  - Covers Hysteria2 obfs links and JSON outbound option mirroring.

## Verification

- Linear `SIN-67` moved to `In Progress`.
- Branch created from current `origin/develop`: `vadigofficial/sin-67-ne-rabotaet-protokol-hysteria2`.
- Graphify context used from `graphify-out`:
  - `Go Domain Models`
  - `Inbound & Log DTOs`
  - `Subscription Handler`
  - Direct graph hits for `sbHysteria2Inbound`, `buildHysteria2()`, and Hysteria2 sublink tests.
- Live RF server checks:
  - Superpanel: `shilka.service` active; no configured inbounds at the time of inspection.
  - Node: `shilka.service` active with VLESS, Hysteria2, and Naive inbounds; Hysteria2 UDP listener active on port 8444.
  - Node sing-box version: `1.13.12`.
  - Temporary Hysteria2 client config passed `sing-box check`.
  - Temporary loopback client proxied `curl https://api.ipify.org` through Hysteria2 and returned the node public IP.
  - New Hysteria2 JSON fields accepted by real node `sing-box check`.
- Local checks:
  - `go test ./tests/services/sublink`
  - `go test ./tests/...`
  - `go build ./...`
  - `go vet ./...`

No production service was restarted or modified during verification. Temporary test files were removed.
