# Installer network resilience — SIN-54 (2026-06-04)

**Linear:** [SIN-54](https://linear.app/sing-box-pannel/issue/SIN-54) (Urgent, parent epic SIN-50) · GitHub issue #35
**Branch:** `vadigofficial/sin-54-installsh-ne-ustoichiv-k-obrivam-seti-flap-rvet-ustanovku` → merged into `develop`
**Files:** `scripts/install.sh`, `scripts/update.sh`, `.github/workflows/ci.yml`

## What was broken

On a cheap VPS with an unstable link (node `2.26.22.151`), `install.sh` died while downloading binaries:

```
Installing sing-box 1.13.12...
curl: (35) getpeername() failed with errno 107: Transport endpoint is not connected
BDEPLOY_DONE exit=35
```

Root causes:

1. **No retry/resume on downloads.** Every binary/checksum/helper fetch used a bare
   `curl -fL` / `curl -fsSL` with no `--retry`, `--retry-all-errors`, or `-C -`. Any
   momentary connection drop returned a hard error, and `set -Eeuo pipefail` killed the
   whole script.
2. **Half-state on update.** The manual update workflow is `uninstall.sh` → `install.sh`.
   `uninstall.sh:remove_dirs` wipes `APP_HOME/CONFIG/DATA/LOG` first; if the subsequent
   `install.sh` then failed mid-download, the node was left with the old panel gone and the
   new one never installed (`prod.yaml` missing, `/opt/shilka/bin` missing, service inactive).
3. **`update.sh` had the same fragile curls** (the web "update" button path), even though it
   already did an atomic binary swap — so the web updater was equally exposed to curl 35.

## What changed

### 1. Robust download helper (retry + resume) — `install.sh` and `update.sh`

The two scripts don't share a common lib (they each duplicate `detect_arch`, `sha256_file`),
so the helper was added to **both**. Flags follow the ticket's recommendation plus the live
workaround that actually survived the flaps:

```bash
DOWNLOAD_RETRIES="${DOWNLOAD_RETRIES:-5}"
DOWNLOAD_RETRY_DELAY="${DOWNLOAD_RETRY_DELAY:-2}"
DOWNLOAD_CONNECT_TIMEOUT="${DOWNLOAD_CONNECT_TIMEOUT:-10}"

_curl_retry_flags() {              # --retry-all-errors needs curl >= 7.71; degrade if absent
  if curl --help all 2>/dev/null | grep -q -- '--retry-all-errors'; then
    printf '%s\n' --retry-all-errors
  fi
}

download_file() {                  # <url> <out> — resumable (-C -), retried
  local url="$1" out="$2" extra=()
  mapfile -t extra < <(_curl_retry_flags)
  curl --fail --location --show-error --connect-timeout "${DOWNLOAD_CONNECT_TIMEOUT}" \
    --retry "${DOWNLOAD_RETRIES}" --retry-delay "${DOWNLOAD_RETRY_DELAY}" \
    --retry-connrefused ${extra[@]+"${extra[@]}"} -C - -o "${out}" "${url}"
}

fetch_url() {                      # <url> — retried fetch to stdout (GitHub API metadata)
  ... same flags, no -C -/-o ...
}
```

Notes:
- `-C -` + `--retry` makes curl resume the partial file across retries (the exact trick that
  worked in the manual workaround).
- `${extra[@]+"${extra[@]}"}` keeps empty-array expansion safe under `set -u`.
- On ancient curl (< 7.71) `--retry-all-errors` is silently dropped; plain `--retry` remains.

Call sites converted:
- `install.sh`: `resolve_sing_box_version`, `install_acme` (the `get.acme.sh` pipe), sing-box
  tarball, `verify_panel_checksum` checksums.txt, latest-release lookup, panel binary, update
  helper download.
- `update.sh`: `resolve_panel_version`, `verify_panel_checksum`, `install_panel_binary`.
- **Left as-is:** `detect_public_ip` (ipify/ifconfig/icanhazip) — non-fatal, already has
  fallbacks and `|| true`.

### 2. Stage-then-commit + rollback — `install.sh`

`main()` was restructured so all network-fragile work happens and is verified **before** any
live file is touched:

- **`stage_binaries()`** (new): one `mktemp -d` staging dir (EXIT-trap cleanup). Downloads +
  extracts sing-box, downloads + checksum-verifies the shilka panel binary, and downloads the
  `update.sh` helper — all into staging. Any failure here exits with the working install
  untouched (no half-state).
- **`commit_binaries()`** (new): only runs after staging fully succeeds. For each target it
  backs up the live file to `*.previous`, writes `*.new`, then `mv -f` (atomic rename). An
  `ERR` trap runs `rollback_commit` to restore `*.previous` if a later swap fails.
- **`configure_update_helper()`** (new): the sudoers/`visudo` wiring that used to live in
  `install_update_helper` — the helper script itself is now placed by `commit_binaries`.
- Removed the old `install_sing_box`, `install_panel_binary`, `install_update_helper`
  (download logic folded into staging/commit).

New `main()` order:
```
require_root → gather_input → create_user_and_dirs
→ stage_binaries          # fragile downloads + checksum verify, all-or-nothing
→ [acme issuance]         # only after binaries are staged & verified
→ commit_binaries         # atomic swap + rollback
→ configure_update_helper → write_prod_config → install_systemd
```
`.previous` backups are kept (same convention as `update.sh`) for manual rollback.

### 3. CI

`.github/workflows/ci.yml` lint step extended from `shellcheck scripts/install.sh` to also
lint `scripts/update.sh scripts/uninstall.sh` (update.sh is now in scope).

## Verification

- `bash -n` clean on all three scripts.
- `shellcheck scripts/install.sh scripts/update.sh scripts/uninstall.sh` — **clean**.
- No bare `curl -fL`/`curl -fsSL` remain on binary/checksum/helper downloads (only the helper
  defs, the feature-detect, and the intentional non-fatal `detect_public_ip`).
- Empty-array flag expansion verified safe under `set -Eeuo pipefail`.
- Real GitHub release CDN confirmed to support byte ranges (`HTTP 206`, `accept-ranges: bytes`,
  `content-range: …`), so `-C -` resume is valid in production.
- No new bash-version floor: `install.sh` already required bash ≥ 4.0 via `${var,,}`; `mapfile`
  is also 4.0+. Target VPS (Debian/Ubuntu) ship bash ≥ 4.4.

**Not run here (macOS):** the live flap simulation (`tc qdisc … netem loss`) and the
idempotency/rollback end-to-end are Linux/root/VPS-only — to be exercised on a test node or the
unstable node from the ticket. macOS ships bash 3.2 (no `mapfile`), so `download_file` can only
run on the Linux target, not in this sandbox.
