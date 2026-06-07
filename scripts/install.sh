#!/usr/bin/env bash
set -Eeuo pipefail

ORIGINAL_ARGS=("$@")
ORIGINAL_ARG_COUNT="$#"

APP_USER="${APP_USER:-shilka}"
APP_HOME="${APP_HOME:-/opt/shilka}"
CONFIG_DIR="${CONFIG_DIR:-/etc/shilka}"
DATA_DIR="${DATA_DIR:-/var/lib/shilka}"
LOG_DIR="${LOG_DIR:-/var/log/shilka}"
TLS_CERT_DIR="${TLS_CERT_DIR:-${CONFIG_DIR}/tls}"
UPDATE_SCRIPT_PATH="${UPDATE_SCRIPT_PATH:-/usr/local/sbin/shilka-update}"
UPDATE_SUDOERS_PATH="${UPDATE_SUDOERS_PATH:-/etc/sudoers.d/shilka-update}"
SING_BOX_VERSION="${SING_BOX_VERSION:-latest}"
PANEL_VERSION="${PANEL_VERSION:-latest}"
GITHUB_REPO="${GITHUB_REPO:-Web-Panel-Sing-Box/shilka-web-panel}"
INSTALL_SOURCE_URL="${SHILKA_INSTALL_URL:-https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/install.sh}"

# Network resilience for flaky VPS links (SIN-54): retry + resume downloads.
DOWNLOAD_RETRIES="${DOWNLOAD_RETRIES:-5}"
DOWNLOAD_RETRY_DELAY="${DOWNLOAD_RETRY_DELAY:-2}"
DOWNLOAD_CONNECT_TIMEOUT="${DOWNLOAD_CONNECT_TIMEOUT:-10}"

# Non-interactive install + local binary (SIN-59).
ASSUME_YES="${SHILKA_ASSUME_YES:-false}"
PANEL_BINARY="${PANEL_BINARY:-}"

if [[ -t 1 && -z "${NO_COLOR:-}" ]]; then
  red=$'\033[0;31m'
  green=$'\033[0;32m'
  blue=$'\033[0;34m'
  yellow=$'\033[0;33m'
  bold=$'\033[1m'
  plain=$'\033[0m'
else
  red=''
  green=''
  blue=''
  yellow=''
  bold=''
  plain=''
fi

log_info() { printf '%b[INF]%b %s\n' "${green}" "${plain}" "$*"; }
log_warn() { printf '%b[WRN]%b %s\n' "${yellow}" "${plain}" "$*" >&2; }
log_error() { printf '%b[ERR]%b %s\n' "${red}" "${plain}" "$*" >&2; }
die() {
  log_error "$*"
  exit 1
}
section() {
  printf '\n%b%s%b\n' "${bold}" "$1" "${plain}"
  printf '%s\n' "--------------------------------------------"
}

PROMPT_INPUT="/dev/stdin"
PROMPT_OUTPUT="/dev/stdout"
PROMPT_TTY_UNAVAILABLE=false
if [[ ! -t 0 ]]; then
  if { exec 3</dev/tty 4>/dev/tty; } 2>/dev/null; then
    PROMPT_INPUT="/dev/fd/3"
    PROMPT_OUTPUT="/dev/fd/4"
  else
    PROMPT_TTY_UNAVAILABLE=true
  fi
fi

ensure_prompt_input() {
  if [[ "${PROMPT_TTY_UNAVAILABLE}" == "true" ]]; then
    die "Interactive install needs a terminal. Re-run with --yes and SHILKA_* env values for headless install."
  fi
}

read_line() {
  local __var="$1"
  ensure_prompt_input
  IFS= read -r "${__var?}" <"${PROMPT_INPUT}"
}

read_secret() {
  local __var="$1"
  ensure_prompt_input
  IFS= read -rs "${__var?}" <"${PROMPT_INPUT}"
}

usage() {
  cat <<'USAGE'
Usage: install.sh [--yes] [--panel-binary PATH] [--help]

  --yes, -y            Non-interactive: skip all prompts, use env/defaults.
  --panel-binary PATH  Install the panel from a local file (skip GitHub download).
  --help, -h           Show this help.

Non-interactive env knobs (used with --yes; env always wins over prompts):
  SHILKA_HOST            Domain or IP users reach the panel on.
  SHILKA_PORT            Panel port (default: random 10000-65535).
  SHILKA_PATH            Web path prefix (default: random hex).
  SHILKA_EXPOSURE        direct | reverse_proxy (default: direct).
  SHILKA_TLS_MODE        self_signed | letsencrypt | off (default: self_signed).
  SHILKA_ACME_EMAIL      Email for Let's Encrypt (required when letsencrypt).
  SHILKA_ADMIN_USER      Admin username (default: random).
  SHILKA_ADMIN_PASSWORD  Admin password (default: auto-generated).
  SHILKA_INSTALL_URL     Installer URL for non-root piped sudo bootstrap.
  PANEL_BINARY=/path     Same as --panel-binary.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -y | --yes) ASSUME_YES=true ;;
    --panel-binary) PANEL_BINARY="${2:-}"; shift ;;
    --panel-binary=*) PANEL_BINARY="${1#*=}" ;;
    -h | --help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage; exit 1 ;;
  esac
  shift
done

# Public SHILKA_* env knobs -> internal variables. Initialised here so the
# rest of the script (under `set -u`) can reference them unconditionally.
PANEL_HOST="${PANEL_HOST:-${SHILKA_HOST:-}}"
PANEL_PORT="${PANEL_PORT:-${SHILKA_PORT:-}}"
PANEL_PATH="${PANEL_PATH:-${SHILKA_PATH:-}}"
PANEL_EXPOSURE="${PANEL_EXPOSURE:-${SHILKA_EXPOSURE:-}}"
ADMIN_USER="${ADMIN_USER:-${SHILKA_ADMIN_USER:-}}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-${SHILKA_ADMIN_PASSWORD:-}}"
ACME_EMAIL="${ACME_EMAIL:-${SHILKA_ACME_EMAIL:-}}"
TLS_MODE="${TLS_MODE:-${SHILKA_TLS_MODE:-}}" # letsencrypt | self_signed | off
CERT_TYPE="${CERT_TYPE:-}"

# ask <varname> <prompt> <default> [validator]: keep an existing (env) value;
# otherwise use the default in non-interactive mode; otherwise prompt.
ask() {
  local __var="$1" __prompt="$2" __default="$3" __validator="${4:-}" __reply __value
  if [[ -n "${!__var:-}" ]]; then
    __value="${!__var}"
    if [[ -n "${__validator}" ]] && ! "${__validator}" "${__value}"; then
      die "Invalid ${__prompt}: ${__value}"
    fi
    return
  fi
  if [[ "${ASSUME_YES}" == "true" ]]; then
    printf -v "${__var}" '%s' "${__default}"
    if [[ -n "${__validator}" ]] && ! "${__validator}" "${__default}"; then
      die "Invalid default for ${__prompt}: ${__default}"
    fi
    return
  fi
  while true; do
    printf '%b?%b %s [%s]: ' "${blue}" "${plain}" "${__prompt}" "${__default}" >"${PROMPT_OUTPUT}"
    if ! read_line __reply; then
      die "Input aborted"
    fi
    __value="${__reply:-${__default}}"
    if [[ -z "${__validator}" ]] || "${__validator}" "${__value}"; then
      printf -v "${__var}" '%s' "${__value}"
      return
    fi
  done
}

confirm_prompt() {
  local prompt="$1" default="${2:-y}" reply
  if [[ "${ASSUME_YES}" == "true" ]]; then
    [[ "${default,,}" == "y" || "${default,,}" == "yes" ]]
    return
  fi
  while true; do
    if [[ "${default,,}" == "y" || "${default,,}" == "yes" ]]; then
      printf '%b?%b %s [Y/n]: ' "${blue}" "${plain}" "${prompt}" >"${PROMPT_OUTPUT}"
    else
      printf '%b?%b %s [y/N]: ' "${blue}" "${plain}" "${prompt}" >"${PROMPT_OUTPUT}"
    fi
    if ! read_line reply; then
      die "Input aborted"
    fi
    reply="${reply:-${default}}"
    case "${reply,,}" in
      y | yes) return 0 ;;
      n | no) return 1 ;;
      *) log_warn "Enter y or n." ;;
    esac
  done
}

validate_non_empty() {
  if [[ -z "$1" ]]; then
    log_warn "Value cannot be empty."
    return 1
  fi
}

validate_host() {
  if [[ -z "$1" || "$1" =~ [[:space:]/] ]]; then
    log_warn "Enter a domain, hostname, or IP without spaces or slashes."
    return 1
  fi
}

validate_port() {
  if ! [[ "$1" =~ ^[0-9]+$ ]] || (( 10#${1} < 1 || 10#${1} > 65535 )); then
    log_warn "Enter a port from 1 to 65535."
    return 1
  fi
  if port_in_use "$1"; then
    log_warn "Port $1 is already in use."
    return 1
  fi
}

validate_panel_path() {
  if [[ -z "$1" || ! "$1" =~ ^/?[A-Za-z0-9._~/-]+$ || "$1" == *"//"* ]]; then
    log_warn "Use letters, digits, dots, dashes, underscores, tildes, and slashes only."
    return 1
  fi
}

validate_admin_user() {
  if [[ -z "$1" || ! "$1" =~ ^[A-Za-z0-9._-]{3,64}$ ]]; then
    log_warn "Use 3-64 letters, digits, dots, dashes, or underscores."
    return 1
  fi
}

validate_email() {
  if [[ -z "$1" || ! "$1" =~ ^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$ ]]; then
    log_warn "Enter a valid email address."
    return 1
  fi
}

ask_password() {
  local __var="$1" __default="$2" first second
  if [[ -n "${!__var:-}" ]]; then
    return
  fi
  if [[ "${ASSUME_YES}" == "true" ]]; then
    printf -v "${__var}" '%s' "${__default}"
    return
  fi
  while true; do
    printf '%b?%b Admin password [auto-generate]: ' "${blue}" "${plain}" >"${PROMPT_OUTPUT}"
    if ! read_secret first; then
      die "Input aborted"
    fi
    printf '\n' >"${PROMPT_OUTPUT}"
    if [[ -z "${first}" ]]; then
      printf -v "${__var}" '%s' "${__default}"
      log_info "Generated admin password."
      return
    fi
    printf '%b?%b Confirm admin password: ' "${blue}" "${plain}" >"${PROMPT_OUTPUT}"
    if ! read_secret second; then
      die "Input aborted"
    fi
    printf '\n' >"${PROMPT_OUTPUT}"
    if [[ "${first}" == "${second}" ]]; then
      printf -v "${__var}" '%s' "${first}"
      return
    fi
    log_warn "Passwords do not match."
  done
}

# --retry-all-errors needs curl >= 7.71; degrade gracefully on older curl.
_curl_retry_flags() {
  if curl --help all 2>/dev/null | grep -q -- '--retry-all-errors'; then
    printf '%s\n' --retry-all-errors
  fi
}

# download_file <url> <output-path> — resumable, retried download.
download_file() {
  local url="$1" out="$2" retry_flag extra=()
  retry_flag="$(_curl_retry_flags)"
  [[ -n "${retry_flag}" ]] && extra+=("${retry_flag}")
  curl --fail --location --show-error \
    --connect-timeout "${DOWNLOAD_CONNECT_TIMEOUT}" \
    --retry "${DOWNLOAD_RETRIES}" --retry-delay "${DOWNLOAD_RETRY_DELAY}" \
    --retry-connrefused ${extra[@]+"${extra[@]}"} \
    -C - -o "${out}" "${url}"
}

# fetch_url <url> — retried fetch to stdout (for GitHub API metadata).
fetch_url() {
  local url="$1" retry_flag extra=()
  retry_flag="$(_curl_retry_flags)"
  [[ -n "${retry_flag}" ]] && extra+=("${retry_flag}")
  curl --fail --location --show-error \
    --connect-timeout "${DOWNLOAD_CONNECT_TIMEOUT}" \
    --retry "${DOWNLOAD_RETRIES}" --retry-delay "${DOWNLOAD_RETRY_DELAY}" \
    --retry-connrefused ${extra[@]+"${extra[@]}"} "${url}"
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    if ! command -v sudo &>/dev/null; then
      die "Root privileges are required and sudo is not installed. Re-run as root."
    fi
    log_info "Root privileges required. Re-running installer through sudo..."
    local env_args=(
      "APP_USER=${APP_USER}"
      "APP_HOME=${APP_HOME}"
      "CONFIG_DIR=${CONFIG_DIR}"
      "DATA_DIR=${DATA_DIR}"
      "LOG_DIR=${LOG_DIR}"
      "TLS_CERT_DIR=${TLS_CERT_DIR}"
      "UPDATE_SCRIPT_PATH=${UPDATE_SCRIPT_PATH}"
      "UPDATE_SUDOERS_PATH=${UPDATE_SUDOERS_PATH}"
      "SING_BOX_VERSION=${SING_BOX_VERSION}"
      "PANEL_VERSION=${PANEL_VERSION}"
      "GITHUB_REPO=${GITHUB_REPO}"
      "SHILKA_INSTALL_URL=${INSTALL_SOURCE_URL}"
      "DOWNLOAD_RETRIES=${DOWNLOAD_RETRIES}"
      "DOWNLOAD_RETRY_DELAY=${DOWNLOAD_RETRY_DELAY}"
      "DOWNLOAD_CONNECT_TIMEOUT=${DOWNLOAD_CONNECT_TIMEOUT}"
      "SHILKA_ASSUME_YES=${ASSUME_YES}"
      "PANEL_BINARY=${PANEL_BINARY}"
      "PANEL_HOST=${PANEL_HOST}"
      "PANEL_PORT=${PANEL_PORT}"
      "PANEL_PATH=${PANEL_PATH}"
      "PANEL_EXPOSURE=${PANEL_EXPOSURE}"
      "ADMIN_USER=${ADMIN_USER}"
      "ADMIN_PASSWORD=${ADMIN_PASSWORD}"
      "ACME_EMAIL=${ACME_EMAIL}"
      "TLS_MODE=${TLS_MODE}"
      "CERT_TYPE=${CERT_TYPE}"
    )
    if [[ ! -r "$0" || "$(basename "$0")" == "bash" ]]; then
      log_info "Installer was started from stdin. Re-fetching through sudo..."
      if (( ORIGINAL_ARG_COUNT > 0 )); then
        fetch_url "${INSTALL_SOURCE_URL}" | sudo env "${env_args[@]}" bash -s -- "${ORIGINAL_ARGS[@]}"
      else
        fetch_url "${INSTALL_SOURCE_URL}" | sudo env "${env_args[@]}" bash -s --
      fi
      exit $?
    fi
    if (( ORIGINAL_ARG_COUNT > 0 )); then
      exec sudo env "${env_args[@]}" bash "$0" "${ORIGINAL_ARGS[@]}"
    fi
    exec sudo env "${env_args[@]}" bash "$0"
  fi
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    aarch64 | arm64) echo "arm64" ;;
    armv7l) echo "armv7" ;;
    *) echo "unsupported" ;;
  esac
}

detect_public_ip() {
  local ip
  ip="$(curl -fsS --max-time 5 -4 https://api.ipify.org 2>/dev/null)" || true
  if [[ -z "${ip}" ]]; then
    ip="$(curl -fsS --max-time 5 -4 https://ifconfig.me 2>/dev/null)" || true
  fi
  if [[ -z "${ip}" ]]; then
    ip="$(curl -fsS --max-time 5 -4 https://icanhazip.com 2>/dev/null)" || true
  fi
  echo "${ip}"
}

resolve_sing_box_version() {
  if [[ "${SING_BOX_VERSION}" != "latest" ]]; then
    echo "${SING_BOX_VERSION#v}"
    return
  fi
  fetch_url https://api.github.com/repos/SagerNet/sing-box/releases/latest \
    | sed -n 's/.*"tag_name": "v\{0,1\}\([^"]*\)".*/\1/p' \
    | head -n 1
}

random_port() {
  local port
  while true; do
    port=$(( (RANDOM % 55535) + 10000 ))
    if ! port_in_use "${port}"; then
      echo "${port}"
      return
    fi
  done
}

random_hex() {
  local len="${1:-16}"
  openssl rand -hex "${len}"
}

random_username() {
  local prefixes=("admin" "root" "operator" "manager" "super")
  local prefix="${prefixes[$((RANDOM % ${#prefixes[@]}))]}"
  echo "${prefix}_$(random_hex 4)"
}

random_base_path() {
  echo "/$(random_hex 8)"
}

is_ipv4() {
  [[ "$1" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]
}

is_ipv6() {
  [[ "$1" == *:* ]]
}

url_host() {
  if is_ipv6 "$1" && [[ "$1" != \[*\] ]]; then
    printf '[%s]\n' "$1"
  else
    printf '%s\n' "$1"
  fi
}

is_domain() {
  local host="$1"
  [[ -z "${host}" ]] && return 1
  is_ipv4 "${host}" && return 1
  [[ "${host}" == *.* ]] && return 0
  return 1
}

port_in_use() {
  local port="$1"
  if command -v ss >/dev/null 2>&1; then
    ss -tuln 2>/dev/null | awk '{print $5}' | grep -Eq ":${port}$"
    return
  fi
  if command -v netstat >/dev/null 2>&1; then
    netstat -lnt 2>/dev/null | awk '{print $4}' | grep -Eq ":${port}$|:${port} "
    return
  fi
  if command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
    return
  fi
  return 1
}

install_acme() {
  if [[ -x ~/.acme.sh/acme.sh ]]; then
    echo "acme.sh is already installed."
    return
  fi

  echo "Installing acme.sh..."
  if ! command -v curl &>/dev/null; then
    if command -v apt-get &>/dev/null; then
      apt-get update -qq && apt-get install -y -qq curl
    else
      echo "ERROR: curl is required to install acme.sh"
      exit 1
    fi
  fi

  fetch_url https://get.acme.sh | sh
  if [[ ! -x ~/.acme.sh/acme.sh ]]; then
    echo "ERROR: acme.sh installation failed"
    exit 1
  fi
  ~/.acme.sh/acme.sh --set-default-ca --server letsencrypt --force
}

write_cert_deploy_hook() {
  cat >/usr/local/bin/shilka-cert-deploy <<HOOK
#!/usr/bin/env bash
set -Eeuo pipefail
CERT_DIR="${TLS_CERT_DIR}"
APP_USER="${APP_USER}"
chown "\${APP_USER}:\${APP_USER}" "\${CERT_DIR}/cert.pem" "\${CERT_DIR}/key.pem" 2>/dev/null || true
chmod 0600 "\${CERT_DIR}/cert.pem" "\${CERT_DIR}/key.pem" 2>/dev/null || true
if systemctl is-active --quiet shilka.service; then
  systemctl restart shilka.service
fi
HOOK
  chmod 0755 /usr/local/bin/shilka-cert-deploy
}

install_cert() {
  local name="$1"
  echo "Installing certificate for ${name}..."
  ~/.acme.sh/acme.sh --installcert -d "${name}" \
    --key-file "${TLS_CERT_DIR}/key.pem" \
    --fullchain-file "${TLS_CERT_DIR}/cert.pem" \
    --reloadcmd "/usr/local/bin/shilka-cert-deploy" \
    --force

  chown "${APP_USER}:${APP_USER}" "${TLS_CERT_DIR}/cert.pem" "${TLS_CERT_DIR}/key.pem"
  chmod 0600 "${TLS_CERT_DIR}/cert.pem" "${TLS_CERT_DIR}/key.pem"

  ~/.acme.sh/acme.sh --upgrade --auto-upgrade
}

issue_domain_cert() {
  local domain="$1"
  echo "Requesting Let's Encrypt certificate for ${domain}..."

  if ! ~/.acme.sh/acme.sh --issue \
    -d "${domain}" \
    --standalone \
    --httpport 80 \
    --force \
    --accountemail "${ACME_EMAIL}"; then
    echo "ERROR: failed to issue certificate for ${domain}. Is port 80 open?"
    return 1
  fi

  write_cert_deploy_hook
  install_cert "${domain}"
}

issue_ip_cert() {
  local ip="$1"
  echo "Requesting Let's Encrypt IP certificate for ${ip} (shortlived profile, 6-day validity)..."

  if ! ~/.acme.sh/acme.sh --issue \
    -d "${ip}" \
    --standalone \
    --server letsencrypt \
    --certificate-profile shortlived \
    --days 6 \
    --httpport 80 \
    --force \
    --accountemail "${ACME_EMAIL}"; then
    echo "ERROR: failed to issue IP certificate for ${ip}. Is port 80 open?"
    return 1
  fi

  write_cert_deploy_hook
  install_cert "${ip}"
}

create_user_and_dirs() {
  if ! id "${APP_USER}" >/dev/null 2>&1; then
    useradd --system --home "${APP_HOME}" --shell /usr/sbin/nologin "${APP_USER}"
  fi
  install -d -m 0755 "${APP_HOME}" "${APP_HOME}/bin" "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"
  install -d -m 0700 "${DATA_DIR}/tls" "${TLS_CERT_DIR}"
  chown -R "${APP_USER}:${APP_USER}" "${APP_HOME}" "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"
}

# Download + verify every binary into a staging dir BEFORE touching the live
# install (SIN-54). Any network flap fails here, leaving the working node
# untouched — no half-state. Sets STAGE_DIR for commit_binaries.
STAGE_DIR=""

stage_binaries() {
  local arch sb_version sb_asset sb_url panel_version panel_asset panel_url update_url
  arch="$(detect_arch)"
  if [[ "${arch}" == "unsupported" ]]; then
    echo "Unsupported CPU architecture: $(uname -m)"
    exit 1
  fi

  STAGE_DIR="$(mktemp -d)"
  trap 'rm -rf "${STAGE_DIR}"' EXIT

  # sing-box core
  sb_version="$(resolve_sing_box_version)"
  echo "Downloading sing-box ${sb_version}..."
  sb_asset="sing-box-${sb_version}-linux-${arch}.tar.gz"
  sb_url="https://github.com/SagerNet/sing-box/releases/download/v${sb_version}/${sb_asset}"
  download_file "${sb_url}" "${STAGE_DIR}/${sb_asset}"
  tar -xzf "${STAGE_DIR}/${sb_asset}" -C "${STAGE_DIR}"
  cp "${STAGE_DIR}/sing-box-${sb_version}-linux-${arch}/sing-box" "${STAGE_DIR}/sing-box"
  if [[ -f "${STAGE_DIR}/sing-box-${sb_version}-linux-${arch}/libcronet.so" ]]; then
    cp "${STAGE_DIR}/sing-box-${sb_version}-linux-${arch}/libcronet.so" "${STAGE_DIR}/libcronet.so"
  fi

  # Shilka panel — local file (SIN-59) or release download.
  if [[ -n "${PANEL_BINARY}" ]]; then
    if [[ ! -f "${PANEL_BINARY}" ]]; then
      echo "ERROR: PANEL_BINARY not found: ${PANEL_BINARY}"
      exit 1
    fi
    echo "Using local Shilka binary ${PANEL_BINARY} (skipping download + checksum)..."
    install -m 0755 "${PANEL_BINARY}" "${STAGE_DIR}/shilka"
  else
    if [[ "${PANEL_VERSION}" == "latest" ]]; then
      echo "Fetching latest Shilka release..."
      panel_version="$(fetch_url "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" \
        | sed -n 's/.*"tag_name": "v\{0,1\}\([^"]*\)".*/\1/p' \
        | head -n 1)"
      if [[ -z "${panel_version}" ]]; then
        echo "ERROR: could not find latest Shilka release. Set PANEL_VERSION manually."
        exit 1
      fi
    else
      panel_version="${PANEL_VERSION#v}"
    fi
    panel_asset="shilka-linux-${arch}"
    panel_url="https://github.com/${GITHUB_REPO}/releases/download/v${panel_version}/${panel_asset}"
    echo "Downloading Shilka ${panel_version} (linux-${arch})..."
    download_file "${panel_url}" "${STAGE_DIR}/${panel_asset}"
    verify_panel_checksum "${STAGE_DIR}" "${panel_asset}" "${panel_version}"
    cp "${STAGE_DIR}/${panel_asset}" "${STAGE_DIR}/shilka"
  fi

  # Update helper script
  update_url="https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/update.sh"
  echo "Downloading Shilka update helper..."
  download_file "${update_url}" "${STAGE_DIR}/shilka-update"
}

# Atomic swap with rollback (SIN-54): back up the live file to .previous, write
# .new, then rename. If a later swap fails, restore everything already swapped.
COMMITTED_TARGETS=()

rollback_commit() {
  local target
  for target in "${COMMITTED_TARGETS[@]}"; do
    if [[ -e "${target}.previous" ]]; then
      echo "Rolling back ${target}..."
      mv -f "${target}.previous" "${target}"
    fi
  done
}

commit_binary() {
  local src="$1" dest="$2"
  if [[ -e "${dest}" ]]; then
    cp -a "${dest}" "${dest}.previous"
  fi
  install -m 0755 "${src}" "${dest}.new"
  mv -f "${dest}.new" "${dest}"
  COMMITTED_TARGETS+=("${dest}")
}

remove_committed_binary() {
  local dest="$1"
  if [[ ! -e "${dest}" ]]; then
    return
  fi
  cp -a "${dest}" "${dest}.previous"
  rm -f "${dest}"
  COMMITTED_TARGETS+=("${dest}")
}

commit_binaries() {
  trap 'rollback_commit' ERR
  commit_binary "${STAGE_DIR}/sing-box" "${APP_HOME}/bin/sing-box"
  if [[ -f "${STAGE_DIR}/libcronet.so" ]]; then
    commit_binary "${STAGE_DIR}/libcronet.so" "${APP_HOME}/bin/libcronet.so"
  else
    remove_committed_binary "${APP_HOME}/bin/libcronet.so"
  fi
  commit_binary "${STAGE_DIR}/shilka" "${APP_HOME}/bin/shilka"
  install -d -m 0755 "$(dirname "${UPDATE_SCRIPT_PATH}")"
  commit_binary "${STAGE_DIR}/shilka-update" "${UPDATE_SCRIPT_PATH}"
  trap - ERR
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
  echo "ERROR: sha256sum or shasum is required to verify Shilka release assets" >&2
  return 1
}

verify_panel_checksum() {
  local tmp="$1" asset="$2" version="$3" expected actual checksums_url
  checksums_url="https://github.com/${GITHUB_REPO}/releases/download/v${version}/checksums.txt"
  echo "Verifying Shilka release checksum..."
  download_file "${checksums_url}" "${tmp}/checksums.txt"
  expected="$(awk -v asset="${asset}" '$2 == asset || $2 == "dist/" asset { print $1; found=1; exit } END { if (!found) exit 1 }' "${tmp}/checksums.txt")" || {
    echo "ERROR: checksum for ${asset} not found in release checksums.txt"
    exit 1
  }
  actual="$(sha256_file "${tmp}/${asset}")"
  if [[ "${actual}" != "${expected}" ]]; then
    echo "ERROR: checksum mismatch for ${asset}"
    echo "Expected: ${expected}"
    echo "Actual:   ${actual}"
    exit 1
  fi
}

ensure_sudo() {
  if command -v sudo &>/dev/null; then
    return 0
  fi
  if command -v apt-get &>/dev/null; then
    echo "Installing sudo for the panel update helper..."
    apt-get update -qq && apt-get install -y -qq sudo
    return 0
  fi
  echo "WARNING: sudo is not installed. The web update button will stay unavailable."
  return 1
}

# Sudoers wiring for the update helper; the script itself is placed by
# commit_binaries (downloaded during stage_binaries).
configure_update_helper() {
  echo "Configuring Shilka update helper..."
  chown root:root "${UPDATE_SCRIPT_PATH}"
  chmod 0755 "${UPDATE_SCRIPT_PATH}"

  if ensure_sudo; then
    cat >"${UPDATE_SUDOERS_PATH}" <<SUDOERS
${APP_USER} ALL=(root) NOPASSWD: ${UPDATE_SCRIPT_PATH}
SUDOERS
    chmod 0440 "${UPDATE_SUDOERS_PATH}"
    if command -v visudo &>/dev/null; then
      visudo -cf "${UPDATE_SUDOERS_PATH}" >/dev/null
    fi
  fi
}

gather_input() {
  local public_ip
  public_ip="$(detect_public_ip)"

  printf '\n%bShilka Panel Installer%b\n' "${bold}" "${plain}"
  printf '%s\n' "============================================"

  # Domain / IP
  local default_host="${public_ip:-127.0.0.1}"
  section "Access"
  log_info "Enter the domain, hostname, or IP users will use to reach the panel."
  ask PANEL_HOST "Domain/IP" "${default_host}" validate_host

  if [[ -z "${PANEL_EXPOSURE}" ]]; then
    if [[ "${ASSUME_YES}" == "true" ]]; then
      PANEL_EXPOSURE="direct"
    else
      echo "Panel exposure mode:"
      echo "  1) Direct high port - Shilka listens publicly on the selected port."
      echo "  2) Reverse proxy - Shilka listens on 127.0.0.1 and nginx/Caddy handles public TLS."
      ask PANEL_EXPOSURE "Mode" "1" validate_non_empty
    fi
  fi
  case "${PANEL_EXPOSURE,,}" in
    2 | reverse | proxy | reverse_proxy) PANEL_EXPOSURE="reverse_proxy" ;;
    *) PANEL_EXPOSURE="direct" ;;
  esac

  # Detect whether this is a domain (eligible for Let's Encrypt) or a bare IP.
  USE_ACME=false
  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    log_info "Reverse proxy mode selected. Panel TLS will be off and the panel will bind to 127.0.0.1."
  elif [[ -n "${TLS_MODE}" || "${ASSUME_YES}" == "true" ]]; then
    # Deterministic TLS for non-interactive / explicit SHILKA_TLS_MODE.
    case "${TLS_MODE:-self_signed}" in
      letsencrypt)
        if is_domain "${PANEL_HOST}"; then
          CERT_TYPE="domain"
        elif is_ipv4 "${PANEL_HOST}"; then
          CERT_TYPE="ip"
        else
          die "SHILKA_TLS_MODE=letsencrypt requires a domain or IPv4 host"
        fi
        if [[ -z "${ACME_EMAIL}" ]]; then
          die "SHILKA_ACME_EMAIL is required for SHILKA_TLS_MODE=letsencrypt"
        fi
        validate_email "${ACME_EMAIL}" || die "Invalid SHILKA_ACME_EMAIL: ${ACME_EMAIL}"
        USE_ACME=true
        if port_in_use 80; then
          log_warn "Port 80 is already in use. Standalone Let's Encrypt may fail."
        fi
        ;;
      self_signed | off)
        USE_ACME=false
        ;;
      *)
        die "Invalid SHILKA_TLS_MODE='${TLS_MODE}' (use letsencrypt|self_signed|off)"
        ;;
    esac
  elif is_domain "${PANEL_HOST}"; then
    section "TLS"
    log_info "This looks like a domain. Let's Encrypt can provide a trusted certificate."
    if port_in_use 80; then
      log_warn "Port 80 is already in use. Standalone Let's Encrypt may disrupt the existing service or fail."
      if confirm_prompt "Use Let's Encrypt anyway?" "n"; then
        USE_ACME=true
        CERT_TYPE="domain"
      fi
    else
      if confirm_prompt "Use Let's Encrypt?" "y"; then
        USE_ACME=true
        CERT_TYPE="domain"
      fi
    fi
  elif is_ipv4 "${PANEL_HOST}"; then
    section "TLS"
    log_info "Let's Encrypt supports IP addresses via the shortlived profile (6-day validity, auto-renews)."
    if port_in_use 80; then
      log_warn "Port 80 is already in use. Standalone Let's Encrypt may disrupt the existing service or fail."
      if confirm_prompt "Use Let's Encrypt for this IP anyway?" "n"; then
        USE_ACME=true
        CERT_TYPE="ip"
      fi
    else
      if confirm_prompt "Use Let's Encrypt for this IP?" "y"; then
        USE_ACME=true
        CERT_TYPE="ip"
      fi
    fi
  fi

  if [[ "${USE_ACME}" == "true" && -z "${ACME_EMAIL}" ]]; then
    ask ACME_EMAIL "Let's Encrypt email" "" validate_email
    log_info "Make sure port 80 is reachable from the internet and not in use."
    log_info "acme.sh needs it briefly for HTTP-01 validation."
  fi

  # Port
  section "Panel"
  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    log_info "Panel local port for the reverse proxy target. It must be free on 127.0.0.1."
  else
    log_info "Panel public port. It must be free."
  fi
  ask PANEL_PORT "Port" "$(random_port)" validate_port

  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    PANEL_LISTEN_ADDRESS="127.0.0.1:${PANEL_PORT}"
  else
    PANEL_LISTEN_ADDRESS=":${PANEL_PORT}"
  fi

  # Base path
  log_info "Web panel path prefix for obscurity."
  ask PANEL_PATH "Path" "$(random_base_path)" validate_panel_path
  # Ensure leading /
  [[ "${PANEL_PATH}" != /* ]] && PANEL_PATH="/${PANEL_PATH}"

  local panel_url_host
  panel_url_host="$(url_host "${PANEL_HOST}")"
  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    PANEL_PUBLIC_URL="https://${panel_url_host}${PANEL_PATH}"
  else
    PANEL_PUBLIC_URL="https://${panel_url_host}:${PANEL_PORT}${PANEL_PATH}"
  fi

  # Admin username
  section "Admin"
  ask ADMIN_USER "Username" "$(random_username)" validate_admin_user

  # Admin password
  ask_password ADMIN_PASSWORD "$(openssl rand -base64 18)"

  local tls_summary
  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    tls_summary="off (reverse proxy)"
  elif [[ "${TLS_MODE:-}" == "off" ]]; then
    tls_summary="off"
  elif [[ "${USE_ACME}" == "true" ]]; then
    tls_summary="Let's Encrypt (acme.sh)"
  else
    tls_summary="self-signed"
  fi

  section "Summary"
  printf '  %-11s %s\n' "Domain/IP:" "${PANEL_HOST}"
  printf '  %-11s %s\n' "Listen:" "${PANEL_LISTEN_ADDRESS}"
  printf '  %-11s %s\n' "Public URL:" "${PANEL_PUBLIC_URL}"
  printf '  %-11s %s\n' "Path:" "${PANEL_PATH}"
  printf '  %-11s %s\n' "TLS:" "${tls_summary}"
  printf '  %-11s %s\n' "Username:" "${ADMIN_USER}"
  printf '  %-11s %s\n' "Password:" "${ADMIN_PASSWORD}"
  echo ""
  if [[ "${ASSUME_YES}" != "true" ]]; then
    if ! confirm_prompt "Proceed with installation?" "y"; then
      log_warn "Installation cancelled."
      exit 0
    fi
  fi
}

write_prod_config() {
  local jwt_secret clash_secret tls_mode tls_cert_file tls_key_file self_signed_hosts
  jwt_secret="$(openssl rand -hex 32)"
  clash_secret="$(openssl rand -hex 24)"

  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" || "${TLS_MODE:-}" == "off" ]]; then
    tls_mode="off"
    tls_cert_file=""
    tls_key_file=""
    self_signed_hosts=""
  elif [[ "${USE_ACME}" == "true" ]]; then
    tls_mode="file"
    tls_cert_file="${TLS_CERT_DIR}/cert.pem"
    tls_key_file="${TLS_CERT_DIR}/key.pem"
    self_signed_hosts=""
  else
    tls_mode="self_signed"
    tls_cert_file=""
    tls_key_file=""
    self_signed_hosts="
    - \"${PANEL_HOST}\""
  fi

  cat >"${CONFIG_DIR}/prod.yaml" <<YAML
env: "production"

runtime:
  gomemlimit: "180MiB"
  gogc: 50

database:
  path: "${DATA_DIR}/panel.db"
  journal_mode: "wal"
  synchronous: "normal"
  cache_size_kb: 2000
  mmap_size_mb: 32
  busy_timeout_ms: 5000
  temp_store: "memory"
  foreign_keys: true

http:
  address: "${PANEL_LISTEN_ADDRESS}"
  base_path: "${PANEL_PATH}"
  read_timeout: "10s"
  write_timeout: "15s"
  idle_timeout: "120s"
  shutdown_timeout: "30s"
  max_header_bytes: 1048576
  max_conns: 128

frontend:
  serve_mode: "embed"
  disk_path: "./frontend/dist"
  cache_ttl: "720h"

auth:
  jwt_secret: "${jwt_secret}"
  jwt_expiry: "24h"
  admin_user: "${ADMIN_USER}"
  admin_password: "${ADMIN_PASSWORD}"
  argon2_memory_kb: 65536
  argon2_iterations: 3
  argon2_parallelism: 2
  login_rate_limit: "5/m"
  api_rate_limit: "100/s"

sing_box:
  binary_path: "${APP_HOME}/bin/sing-box"
  config_path: "${CONFIG_DIR}/config.json"
  working_dir: "${CONFIG_DIR}"
  api_address: "127.0.0.1:9090"
  api_secret: "${clash_secret}"
  check_timeout: "10s"
  restart_delay: "3s"
  max_restarts: 4
  process_mode: "auto"
  service_name: "sing-box"
  core_log_path: "${LOG_DIR}/sing-box.log"

stats:
  source: "clash"
  v2ray_api_address: "127.0.0.1:8088"

tls:
  mode: "${tls_mode}"
  cert_file: "${tls_cert_file}"
  key_file: "${tls_key_file}"
  acme_email: ""
  acme_domains: []
  acme_cache_dir: "${DATA_DIR}/acme"
  self_signed_hosts:${self_signed_hosts}
  self_signed_dir: "${DATA_DIR}/tls"

metrics:
  system_interval: "5s"
  traffic_interval: "2s"
  history_size: 120
  batch_flush_interval: "8s"

logging:
  level: "info"
  format: "json"
  file_path: ""
  max_memory_lines: 500
  max_file_size_mb: 10
  max_file_backups: 3

updates:
  repo: "${GITHUB_REPO}"
  script_path: "${UPDATE_SCRIPT_PATH}"
  check_cache_ttl: "10m"
  command_timeout: "10m"

subscription:
  public_url: "${PANEL_PUBLIC_URL}"
  token_ttl: "720h"
YAML

  chmod 0640 "${CONFIG_DIR}/prod.yaml"
  chown "${APP_USER}:${APP_USER}" "${CONFIG_DIR}/prod.yaml"

  echo "${ADMIN_PASSWORD}" >"${CONFIG_DIR}/initial-admin-password"
  chmod 0600 "${CONFIG_DIR}/initial-admin-password"
  chown "${APP_USER}:${APP_USER}" "${CONFIG_DIR}/initial-admin-password"
}

install_systemd() {
  cat >/etc/systemd/system/shilka.service <<UNIT
[Unit]
Description=Shilka web panel
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${APP_USER}
Group=${APP_USER}
WorkingDirectory=${APP_HOME}
Environment=SHILKA_CONFIG_PATH=${CONFIG_DIR}/prod.yaml
ExecStart=${APP_HOME}/bin/shilka run
Restart=on-failure
RestartSec=3
NoNewPrivileges=false
PrivateTmp=true
ProtectSystem=full
ReadWritePaths=${CONFIG_DIR} ${DATA_DIR} ${LOG_DIR}
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
UNIT

  cat >/usr/local/bin/shilka <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail

APP_USER="${APP_USER:-__APP_USER__}"
CONFIG_DIR="${CONFIG_DIR:-__CONFIG_DIR__}"
DATA_DIR="${DATA_DIR:-__DATA_DIR__}"
BIN="${BIN:-__BIN_PATH__}"
SERVICE_NAME="${SERVICE_NAME:-shilka.service}"
export SHILKA_CONFIG_PATH="${SHILKA_CONFIG_PATH:-${CONFIG_DIR}/prod.yaml}"

if [[ -t 1 && -z "${NO_COLOR:-}" ]]; then
  red=$'\033[0;31m'
  green=$'\033[0;32m'
  blue=$'\033[0;34m'
  yellow=$'\033[0;33m'
  bold=$'\033[1m'
  plain=$'\033[0m'
else
  red=''
  green=''
  blue=''
  yellow=''
  bold=''
  plain=''
fi

log_info() { printf '%b[INF]%b %s\n' "${green}" "${plain}" "$*"; }
log_warn() { printf '%b[WRN]%b %s\n' "${yellow}" "${plain}" "$*" >&2; }
log_error() { printf '%b[ERR]%b %s\n' "${red}" "${plain}" "$*" >&2; }

script_path() {
  if [[ "$0" == */* ]]; then
    printf '%s\n' "$0"
    return
  fi
  command -v "$0"
}

sudo_env() {
  printf '%s\0' \
    "APP_USER=${APP_USER}" \
    "CONFIG_DIR=${CONFIG_DIR}" \
    "DATA_DIR=${DATA_DIR}" \
    "BIN=${BIN}" \
    "SERVICE_NAME=${SERVICE_NAME}" \
    "SHILKA_CONFIG_PATH=${SHILKA_CONFIG_PATH}"
}

run_root() {
  if [[ "${EUID}" -eq 0 ]]; then
    "$@"
    return
  fi
  if ! command -v sudo >/dev/null 2>&1; then
    log_error "This action needs root privileges and sudo is not installed."
    return 1
  fi
  local env_args=()
  while IFS= read -r -d '' item; do
    env_args+=("${item}")
  done < <(sudo_env)
  sudo env "${env_args[@]}" "$@"
}

if [[ "$#" -gt 0 ]]; then
  case "$1" in
    -v | --version | version)
      exec "${BIN}" "$@"
      ;;
    *)
      if [[ "${EUID}" -eq 0 ]]; then
        exec "${BIN}" "$@"
      fi
      if ! command -v sudo >/dev/null 2>&1; then
        log_error "Command needs root privileges and sudo is not installed."
        exit 1
      fi
      env_args=()
      while IFS= read -r -d '' item; do
        env_args+=("${item}")
      done < <(sudo_env)
      exec sudo env "${env_args[@]}" "$(script_path)" "$@"
      ;;
  esac
fi

service_active() { systemctl is-active "${SERVICE_NAME}" 2>/dev/null || true; }
service_enabled() { systemctl is-enabled "${SERVICE_NAME}" 2>/dev/null || true; }

pause() {
  echo
  read -rp "Press Enter to return to the main menu: " _ || true
}

confirm() {
  local prompt="$1" default="${2:-n}" reply
  while true; do
    if [[ "${default,,}" == "y" || "${default,,}" == "yes" ]]; then
      read -rp "${prompt} [Y/n]: " reply
    else
      read -rp "${prompt} [y/N]: " reply
    fi
    reply="${reply:-${default}}"
    case "${reply,,}" in
      y | yes) return 0 ;;
      n | no) return 1 ;;
      *) log_warn "Enter y or n." ;;
    esac
  done
}

prompt_required() {
  local label="$1" default="${2:-}" value
  while true; do
    if [[ -n "${default}" ]]; then
      read -rp "${label} [${default}]: " value
      value="${value:-${default}}"
    else
      read -rp "${label}: " value
    fi
    if [[ -n "${value}" ]]; then
      printf '%s\n' "${value}"
      return
    fi
    log_warn "Value cannot be empty."
  done
}

prompt_password() {
  local first second
  while true; do
    read -rsp "New admin password: " first
    echo
    if [[ -z "${first}" ]]; then
      log_warn "Password cannot be empty."
      continue
    fi
    read -rsp "Confirm admin password: " second
    echo
    if [[ "${first}" == "${second}" ]]; then
      printf '%s\n' "${first}"
      return
    fi
    log_warn "Passwords do not match."
  done
}

port_in_use() {
  local port="$1"
  if command -v ss >/dev/null 2>&1; then
    ss -ltn 2>/dev/null | awk -v p=":${port}$" '$4 ~ p {found=1} END {exit found ? 0 : 1}'
    return
  fi
  if command -v netstat >/dev/null 2>&1; then
    netstat -lnt 2>/dev/null | awk -v p=":${port} " '$4 ~ p {found=1} END {exit found ? 0 : 1}'
    return
  fi
  if command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
    return
  fi
  return 1
}

prompt_port() {
  local port
  while true; do
    read -rp "New panel port [1-65535]: " port
    if ! [[ "${port}" =~ ^[0-9]+$ ]] || (( 10#${port} < 1 || 10#${port} > 65535 )); then
      log_warn "Enter a port from 1 to 65535."
      continue
    fi
    if port_in_use "${port}"; then
      log_warn "Port ${port} is already in use."
      continue
    fi
    printf '%s\n' "${port}"
    return
  done
}

start_services() {
  if [[ "$(service_active)" == "active" ]]; then
    log_info "Panel is already running."
    return
  fi
  run_root systemctl start "${SERVICE_NAME}"
  sleep 1
  [[ "$(service_active)" == "active" ]] && log_info "Panel started." || log_warn "Start requested. Check status or logs."
}

stop_services() {
  if [[ "$(service_active)" != "active" ]]; then
    log_info "Panel is already stopped."
    return
  fi
  run_root systemctl stop "${SERVICE_NAME}"
  sleep 1
  [[ "$(service_active)" != "active" ]] && log_info "Panel stopped." || log_warn "Stop requested. Check status or logs."
}

restart_services() {
  run_root systemctl restart "${SERVICE_NAME}"
  sleep 1
  [[ "$(service_active)" == "active" ]] && log_info "Panel restarted." || log_warn "Restart requested. Check status or logs."
}

reset_admin_password() {
  local password
  password="$(prompt_password)"
  run_root "${BIN}" admin reset-password -password "${password}"
  printf '%s\n' "${password}" | run_root tee "${CONFIG_DIR}/initial-admin-password" >/dev/null
  run_root chmod 0600 "${CONFIG_DIR}/initial-admin-password"
  run_root chown "${APP_USER}:${APP_USER}" "${CONFIG_DIR}/initial-admin-password" || true
  restart_services
  log_info "Password reset. Log in with your admin username and the new password."
}

change_panel_port() {
  local port
  port="$(prompt_port)"
  run_root "${BIN}" setting -port "${port}"
  restart_services
  log_info "Panel port set to ${port}."
}

set_domain() {
  local domain public_url
  domain="$(prompt_required "Panel domain")"
  public_url="$(prompt_required "Public URL" "https://${domain}")"
  run_root "${BIN}" setting -domain "${domain}" -public-url "${public_url}"
  restart_services
  log_info "Domain and public URL updated."
}

create_api_token() {
  local name
  name="$(prompt_required "Token name" "node")"
  run_root "${BIN}" api-token create -name "${name}" -scopes node
}

set_custom_cert() {
  local cert key
  cert="$(prompt_required "Certificate path")"
  key="$(prompt_required "Key path")"
  run_root test -f "${cert}" || { log_error "Certificate file not found: ${cert}"; return 1; }
  run_root test -f "${key}" || { log_error "Key file not found: ${key}"; return 1; }
  run_root "${BIN}" cert set-files -cert "${cert}" -key "${key}"
  restart_services
  log_info "Custom TLS certificate configured."
}

disable_panel_tls() {
  confirm "Disable panel TLS? Use only behind a trusted reverse proxy or tunnel." "n" || return
  run_root "${BIN}" cert reset
  restart_services
  log_info "Panel TLS disabled."
}

backup_db() {
  local out="${DATA_DIR}/panel-$(date +%Y%m%d-%H%M%S).db"
  run_root test -f "${DATA_DIR}/panel.db" || { log_error "Database not found: ${DATA_DIR}/panel.db"; return 1; }
  run_root cp -a "${DATA_DIR}/panel.db" "${out}"
  log_info "Backup: ${out}"
}

core_reload() {
  run_root "${BIN}" core reload
}

core_config_check() {
  run_root "${BIN}" core config-check
}

show_service_status() {
  systemctl status "${SERVICE_NAME}" --no-pager || true
}

show_logs() {
  run_root journalctl -u "${SERVICE_NAME}" -n 120 --no-pager || true
}

show_settings() {
  run_root "${BIN}" setting -show || true
}

show_header() {
  local active enabled version
  active="$(service_active)"
  enabled="$(service_enabled)"
  version="$("${BIN}" version 2>/dev/null || true)"
  [[ -z "${version}" ]] && version="shilka"

  printf '\n%bShilka Panel%b\n' "${bold}" "${plain}"
  printf '%s\n' "============================================"
  case "${active}" in
    active) printf 'Status: %bRunning%b' "${green}" "${plain}" ;;
    inactive | failed) printf 'Status: %bStopped%b' "${yellow}" "${plain}" ;;
    *) printf 'Status: %bUnknown%b' "${red}" "${plain}" ;;
  esac
  case "${enabled}" in
    enabled) printf '  Autostart: %bYes%b' "${green}" "${plain}" ;;
    disabled) printf '  Autostart: %bNo%b' "${yellow}" "${plain}" ;;
    *) printf '  Autostart: %bUnknown%b' "${red}" "${plain}" ;;
  esac
  printf '  Version: %s\n' "${version#shilka }"
  printf 'Config: %s\n' "${SHILKA_CONFIG_PATH}"
}

show_menu() {
  show_header
  cat <<'MENU'

  1) Start panel              2) Stop panel
  3) Restart panel            4) Service status
  5) Show logs                6) Show settings

  7) Reset admin password     8) Change panel port
  9) Set domain/public URL   10) Create API token
 11) Set custom TLS cert     12) Disable panel TLS

 13) Core reload             14) Core config check
 15) Backup database          0) Exit
MENU
}

main() {
  while true; do
    show_menu
    read -rp "> " choice
    case "${choice}" in
      1) start_services ;;
      2) stop_services ;;
      3) restart_services ;;
      4) show_service_status ;;
      5) show_logs ;;
      6) show_settings ;;
      7) reset_admin_password ;;
      8) change_panel_port ;;
      9) set_domain ;;
      10) create_api_token ;;
      11) set_custom_cert ;;
      12) disable_panel_tls ;;
      13) core_reload ;;
      14) core_config_check ;;
      15) backup_db ;;
      0) exit 0 ;;
      *) log_warn "Unknown option." ;;
    esac
    pause
  done
}
main "$@"
SCRIPT

  local app_user_escaped config_dir_escaped data_dir_escaped bin_path_escaped
  app_user_escaped="$(printf '%s' "${APP_USER}" | sed 's/[\/&|]/\\&/g')"
  config_dir_escaped="$(printf '%s' "${CONFIG_DIR}" | sed 's/[\/&|]/\\&/g')"
  data_dir_escaped="$(printf '%s' "${DATA_DIR}" | sed 's/[\/&|]/\\&/g')"
  bin_path_escaped="$(printf '%s' "${APP_HOME}/bin/shilka" | sed 's/[\/&|]/\\&/g')"
  sed -i \
    -e "s|__APP_USER__|${app_user_escaped}|g" \
    -e "s|__CONFIG_DIR__|${config_dir_escaped}|g" \
    -e "s|__DATA_DIR__|${data_dir_escaped}|g" \
    -e "s|__BIN_PATH__|${bin_path_escaped}|g" \
    /usr/local/bin/shilka

  chmod 0755 /usr/local/bin/shilka
  systemctl daemon-reload
  systemctl enable --now shilka.service
}

main() {
  require_root
  gather_input
  create_user_and_dirs
  # Pull + verify every binary first; a network flap fails here before any
  # working install is touched (SIN-54).
  stage_binaries
  if [[ "${USE_ACME}" == "true" ]]; then
    install_acme
    if [[ "${CERT_TYPE}" == "domain" ]]; then
      issue_domain_cert "${PANEL_HOST}"
    else
      issue_ip_cert "${PANEL_HOST}"
    fi
  fi
  commit_binaries
  configure_update_helper
  write_prod_config
  install_systemd

  section "Shilka installed"
  printf '  %-10s %s\n' "URL:" "${PANEL_PUBLIC_URL}"
  printf '  %-10s %s\n' "Username:" "${ADMIN_USER}"
  printf '  %-10s %s\n' "Password:" "${ADMIN_PASSWORD}"
  printf '  %-10s %s\n' "Manage:" "shilka"
  printf '  %-10s %s\n' "Logs:" "journalctl -u shilka.service -f"
  printf '  %-10s %s\n' "Config:" "${CONFIG_DIR}/prod.yaml"
}

main "$@"
