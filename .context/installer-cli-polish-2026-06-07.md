# Installer CLI Polish - 2026-06-07

## Status

- Linear issue: `SIN-73` (`Polish shell CLI and installer UX`), set to In Progress.
- Branch: `Vadim-Denisovich/sin-73-cli-installer-polish`.
- Scope: `scripts/install.sh` installer UX and generated `/usr/local/bin/shilka` shell menu.

## What Changed

- Added a non-root bootstrap to `scripts/install.sh`.
  - The installer now re-runs itself through `sudo` when launched by a regular user from a readable script file.
  - It preserves original arguments and install-related environment values such as `PANEL_BINARY`, `PANEL_VERSION`, `SING_BOX_VERSION`, `SHILKA_*`-mapped values, and path overrides.
  - If `sudo` is unavailable, it fails before network or filesystem mutation with a clear error.

- Polished installer prompts.
  - Added color-aware log helpers, section headers, yes/no confirmation, validated text prompts, port validation, path validation, admin username validation, email validation, and password confirmation.
  - Kept `--yes` deterministic and compatible with existing env knobs.
  - Fixed `SHILKA_TLS_MODE=off` so it writes `tls.mode: off` instead of falling through to self-signed TLS.
  - Preserved existing staging, checksum verification, retry/resume downloads, rollback, update helper sudoers setup, and `libcronet.so` handling.

- Rewrote the generated `shilka` shell menu.
  - The menu now has a 3x-ui-inspired status header, numbered actions, colored status/log messages, confirmations, validated inputs, and a return-to-menu pause.
  - Root-only menu actions use `sudo` when launched by a regular user.
  - Application behavior still delegates to the Shilka binary subcommands instead of duplicating Go logic.
  - Direct `shilka <subcommand>` calls also re-run through `sudo` for privileged commands, while `shilka version` remains non-privileged.

## Files Touched

- `scripts/install.sh`
- `.context/installer-cli-polish-2026-06-07.md`

## Verification

- `bash -n scripts/install.sh scripts/update.sh scripts/uninstall.sh` - passed.
- Extracted the generated `/usr/local/bin/shilka` heredoc to `/private/tmp/shilka-menu.sh`; `bash -n /private/tmp/shilka-menu.sh` - passed.
- `shellcheck scripts/install.sh scripts/update.sh scripts/uninstall.sh` - passed.
- `go test ./tests/cmd ./tests/services/updater` - passed.
- `go build ./...` - passed.
- `PATH=/private/tmp/no-sudo /bin/bash scripts/install.sh --yes` - failed early with the expected no-sudo root privilege message and did not prompt or mutate.
- `rg -q "Start panel|Reset admin password|Core config check|Press Enter to return" /private/tmp/shilka-menu.sh` - passed.

## Notes

- `Context7` was not needed because no external bash/TUI library was added.
- Project structure was checked through the local `graphify-out/` graph report and related graph nodes before implementation.
