# Installer Piped Bootstrap and Prompt Fix

## Linear

- Issue: SIN-75, "Fix piped installer sudo bootstrap and prompts"
- Branch: `Vadim-Denisovich/sin-75-fix-piped-installer-prompts`

## What Changed

- Fixed the README install shape:
  `curl -fsSL https://raw.githubusercontent.com/Web-Panel-Sing-Box/shilka-web-panel/main/scripts/install.sh | bash`
- When the installer starts as a non-root piped script, it now re-fetches the installer through `sudo` instead of failing because `$0` is `bash` and not a readable script path.
- Added `SHILKA_INSTALL_URL` so the bootstrap source can be overridden for dry checks, forks, mirrors, or local test URLs.
- Preserved the existing root-owned install model: `/opt/shilka`, `/etc/shilka`, `/var/lib/shilka`, `/usr/local/bin/shilka`, systemd units, download verification, staged binary commits, rollback, retry/resume download handling, and `libcronet.so` behavior.

## Prompt Input Fix

- Added installer prompt input/output routing.
- If stdin is not a terminal, the installer opens `/dev/tty` for interactive prompts and password confirmation.
- This prevents interactive prompts from reading the remaining curl pipe and aborting with `Input aborted`.
- If no terminal is available, interactive mode now fails with a clear headless-install message instead of a misleading prompt failure.
- `--yes` remains deterministic and does not require a terminal.

## Files Touched

- `scripts/install.sh`
- `.context/installer-piped-bootstrap-2026-06-07.md`

## Verification

- `bash -n scripts/install.sh scripts/update.sh scripts/uninstall.sh`
- `shellcheck scripts/install.sh scripts/update.sh scripts/uninstall.sh`
- README-shaped dry bootstrap with fake sudo:
  `cat scripts/install.sh | env PATH="/private/tmp/shilka-fakebin:$PATH" SHILKA_INSTALL_URL="file://${PWD}/scripts/install.sh" bash`
- Generated `/usr/local/bin/shilka` menu heredoc extracted and checked with `bash -n`.
- `git diff --check`
- `go test ./tests/cmd ./tests/services/updater`
- `go build ./...`

## Notes

- No external Bash or TUI library was added, so Context7 was not needed for this fix.
- `graphify-out/` was checked as the local project graph source; the fix stayed scoped to installer shell behavior.
