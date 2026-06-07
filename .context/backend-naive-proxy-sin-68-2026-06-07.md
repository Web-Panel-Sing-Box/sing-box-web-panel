# Backend Naive Proxy SIN-68 - 2026-06-07

## Task

Implemented Linear issue `SIN-68` on branch
`vadigofficial/sin-68-ne-rabotaet-protokol-naive-proxy`, based on release
`v1.11.1` from `main` (`9b8027e`).

The bug was that Naive Proxy subscriptions and Linux client support were not
valid for real sing-box usage:

- sing-box rejects Naive outbound JSON configs containing `tls.insecure`;
- the installer copied only the `sing-box` binary and skipped the bundled
  `libcronet.so`, which Naive outbound needs on Linux.

Graphify context was generated with `graphify update . --no-cluster`. Relevant
nodes were `sbNaiveInbound`, `buildNaive()`, `naiveOutbound`, and the
subscription rendering path.

## What Changed

- Added `sublink.ErrNaiveJSONRequiresTrustedTLS`.
- Changed Naive JSON subscription generation so it never emits unsupported
  `tls.insecure`.
- Naive JSON subscriptions for self-signed or otherwise insecure TLS now fail
  with a clear client error explaining that trusted TLS is required and plain or
  base64 links should be used instead.
- Trusted Naive JSON subscriptions still render valid sing-box outbound TLS with
  `enabled` and `server_name`.
- The public subscription handler maps that Naive JSON limitation to HTTP `400`
  instead of a generic `500`.
- The installer now stages `libcronet.so` from the official sing-box tarball and
  installs it next to `/opt/shilka/bin/sing-box`. If a future sing-box release
  does not ship `libcronet.so`, the installer removes a stale copy with rollback
  support.

## Files Touched

- `internal/services/sublink/subscription.go`
- `internal/transport/handler/subscription_handler.go`
- `scripts/install.sh`
- `tests/services/sublink/builder_test.go`
- `tests/transport/handler/subscription_handler_test.go`

## Verification

Local checks passed:

- `go test ./tests/services/sublink`
- `go test ./tests/transport/handler`
- `go test ./tests/...`
- `go build ./...`
- `go vet ./...`
- `bash -n scripts/install.sh`

Remote node `31.76.12.23`:

- Existing service was upgraded to a branch binary reporting
  `shilka v1.11.1-sin68`.
- Previous panel binary was backed up under `/opt/shilka/bin/`.
- Installed `/opt/shilka/bin/libcronet.so` from the active sing-box release.
- `systemctl is-active shilka.service` returned `active`.
- Managed sing-box subprocess was running as
  `/opt/shilka/bin/sing-box run -c /etc/shilka/config.json`.
- `sing-box check` for a Naive outbound without `tls.insecure` passed after
  `libcronet.so` was installed.
- Public Naive `format=json` subscription returned HTTP `400` with the trusted
  TLS message for the existing self-signed Naive inbound.
- Public Naive `format=plain` subscription still returned HTTP `200` and a
  `naive+https://` share link.

Remote superpanel `31.200.229.59`:

- Found installed release `shilka v1.11.1` and active `shilka.service`.
- No active Naive client existed in the superpanel database, so no live
  subscription mutation was performed there.

## Notes

- No JWT secrets, admin passwords, API tokens, subscription tokens, generated
  private keys, or full share links were recorded in this summary.
- Linear `SIN-68` was moved to `In Progress`.
