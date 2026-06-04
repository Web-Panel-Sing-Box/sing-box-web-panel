#!/usr/bin/env bash
set -Eeuo pipefail

APP_HOME="${APP_HOME:-/opt/shilka}"
LOG_DIR="${LOG_DIR:-/var/log/shilka}"
GITHUB_REPO="${GITHUB_REPO:-Web-Panel-Sing-Box/shilka-web-panel}"
PANEL_VERSION="${PANEL_VERSION:-latest}"
SERVICE_NAME="${SERVICE_NAME:-shilka.service}"
LOCK_DIR="${LOCK_DIR:-/run/shilka-update.lock}"
BIN_PATH="${BIN_PATH:-${APP_HOME}/bin/shilka}"
LOG_FILE="${LOG_FILE:-${LOG_DIR}/update.log}"

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "update.sh must run as root"
    exit 1
  fi
}

setup_logging() {
  install -d -m 0755 "${LOG_DIR}"
  touch "${LOG_FILE}"
  chmod 0640 "${LOG_FILE}" 2>/dev/null || true
  exec > >(tee -a "${LOG_FILE}") 2>&1
}

log() {
  printf '[%s] %s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$*"
}

acquire_lock() {
  if ! mkdir "${LOCK_DIR}" 2>/dev/null; then
    log "Another Shilka update is already running."
    exit 1
  fi
  trap 'rm -rf "${LOCK_DIR}"' EXIT
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    aarch64 | arm64) echo "arm64" ;;
    *) echo "unsupported" ;;
  esac
}

sha256_file() {
  local file="$1"
  if command -v sha256sum &>/dev/null; then
    sha256sum "${file}" | awk '{print $1}'
    return
  fi
  if command -v shasum &>/dev/null; then
    shasum -a 256 "${file}" | awk '{print $1}'
    return
  fi
  log "ERROR: sha256sum or shasum is required"
  exit 1
}

resolve_panel_version() {
  if [[ "${PANEL_VERSION}" != "latest" ]]; then
    echo "${PANEL_VERSION#v}"
    return
  fi
  curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" \
    | sed -n 's/.*"tag_name": "v\{0,1\}\([^"]*\)".*/\1/p' \
    | head -n 1
}

verify_panel_checksum() {
  local tmp="$1" asset="$2" version="$3" expected actual checksums_url
  checksums_url="https://github.com/${GITHUB_REPO}/releases/download/v${version}/checksums.txt"
  log "Verifying checksum for ${asset}"
  curl -fL "${checksums_url}" -o "${tmp}/checksums.txt"
  expected="$(awk -v asset="${asset}" '$2 == asset || $2 == "dist/" asset { print $1; found=1; exit } END { if (!found) exit 1 }' "${tmp}/checksums.txt")" || {
    log "ERROR: checksum for ${asset} not found"
    exit 1
  }
  actual="$(sha256_file "${tmp}/${asset}")"
  if [[ "${actual}" != "${expected}" ]]; then
    log "ERROR: checksum mismatch"
    log "Expected: ${expected}"
    log "Actual:   ${actual}"
    exit 1
  fi
}

install_panel_binary() {
  local arch version asset url tmp backup owner group
  arch="$(detect_arch)"
  if [[ "${arch}" == "unsupported" ]]; then
    log "ERROR: unsupported CPU architecture: $(uname -m)"
    exit 1
  fi

  version="$(resolve_panel_version)"
  if [[ -z "${version}" ]]; then
    log "ERROR: could not resolve Shilka release version"
    exit 1
  fi

  asset="shilka-linux-${arch}"
  url="https://github.com/${GITHUB_REPO}/releases/download/v${version}/${asset}"
  tmp="$(mktemp -d)"
  trap 'rm -rf "${tmp}"; rm -rf "${LOCK_DIR}"' EXIT

  log "Downloading Shilka ${version} (${asset})"
  curl -fL "${url}" -o "${tmp}/${asset}"
  verify_panel_checksum "${tmp}" "${asset}" "${version}"

  if [[ ! -x "${BIN_PATH}" ]]; then
    log "ERROR: existing Shilka binary not found at ${BIN_PATH}"
    exit 1
  fi

  owner="$(stat -c '%U' "${BIN_PATH}" 2>/dev/null || echo root)"
  group="$(stat -c '%G' "${BIN_PATH}" 2>/dev/null || echo root)"
  backup="${BIN_PATH}.previous"

  log "Installing new binary at ${BIN_PATH}"
  cp -a "${BIN_PATH}" "${backup}"
  install -m 0755 "${tmp}/${asset}" "${BIN_PATH}.new"
  chown "${owner}:${group}" "${BIN_PATH}.new" 2>/dev/null || true
  mv -f "${BIN_PATH}.new" "${BIN_PATH}"
}

restart_service() {
  log "Restarting ${SERVICE_NAME}"
  systemctl restart "${SERVICE_NAME}"
  log "Shilka update complete"
}

main() {
  require_root
  setup_logging
  acquire_lock
  log "Starting Shilka update"
  install_panel_binary
  restart_service
}

main "$@"
