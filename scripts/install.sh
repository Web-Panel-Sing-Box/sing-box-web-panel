#!/usr/bin/env bash
set -Eeuo pipefail

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

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "install.sh must run as root"
    exit 1
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
  curl -fsSL https://api.github.com/repos/SagerNet/sing-box/releases/latest \
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

is_domain() {
  local host="$1"
  [[ -z "${host}" ]] && return 1
  is_ipv4 "${host}" && return 1
  [[ "${host}" == *.* ]] && return 0
  return 1
}

port_in_use() {
  local port="$1"
  ss -tuln | awk '{print $5}' | grep -Eq ":${port}$"
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

  curl -s https://get.acme.sh | sh
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

install_sing_box() {
  local arch version asset url tmp
  arch="$(detect_arch)"
  if [[ "${arch}" == "unsupported" ]]; then
    echo "Unsupported CPU architecture: $(uname -m)"
    exit 1
  fi
  version="$(resolve_sing_box_version)"
  echo "Installing sing-box ${version}..."
  asset="sing-box-${version}-linux-${arch}.tar.gz"
  url="https://github.com/SagerNet/sing-box/releases/download/v${version}/${asset}"
  tmp="$(mktemp -d)"
  curl -fL "${url}" -o "${tmp}/${asset}"
  tar -xzf "${tmp}/${asset}" -C "${tmp}"
  install -m 0755 "${tmp}/sing-box-${version}-linux-${arch}/sing-box" "${APP_HOME}/bin/sing-box"
  rm -rf "${tmp}"
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
  curl -fL "${checksums_url}" -o "${tmp}/checksums.txt"
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

install_panel_binary() {
  local arch version asset url tmp
  arch="$(detect_arch)"
  if [[ "${arch}" == "unsupported" ]]; then
    echo "Unsupported CPU architecture: $(uname -m)"
    exit 1
  fi

  if [[ "${PANEL_VERSION}" == "latest" ]]; then
    echo "Fetching latest Shilka release..."
    version="$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" \
      | sed -n 's/.*"tag_name": "v\{0,1\}\([^"]*\)".*/\1/p' \
      | head -n 1)"
    if [[ -z "${version}" ]]; then
      echo "ERROR: could not find latest Shilka release. Set PANEL_VERSION manually."
      exit 1
    fi
  else
    version="${PANEL_VERSION#v}"
  fi

  asset="shilka-linux-${arch}"
  url="https://github.com/${GITHUB_REPO}/releases/download/v${version}/${asset}"
  echo "Downloading Shilka ${version} (linux-${arch})..."
  tmp="$(mktemp -d)"
  curl -fL "${url}" -o "${tmp}/${asset}"
  verify_panel_checksum "${tmp}" "${asset}" "${version}"
  install -m 0755 "${tmp}/${asset}" "${APP_HOME}/bin/shilka"
  rm -rf "${tmp}"
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

install_update_helper() {
  local update_url helper_dir
  helper_dir="$(dirname "${UPDATE_SCRIPT_PATH}")"
  update_url="https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/update.sh"
  echo "Installing Shilka update helper..."
  install -d -m 0755 "${helper_dir}"
  curl -fsSL "${update_url}" -o "${UPDATE_SCRIPT_PATH}"
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

  echo ""
  echo "============================================"
  echo "        Shilka Panel Installer"
  echo "============================================"
  echo ""

  # Domain / IP
  local default_host="${public_ip:-127.0.0.1}"
  echo "Enter the domain or IP address users will use to reach the panel."
  read -rp "Domain/IP [${default_host}]: " input_host
  PANEL_HOST="${input_host:-${default_host}}"

  PANEL_EXPOSURE="${PANEL_EXPOSURE:-direct}"
  echo ""
  echo "Panel exposure mode:"
  echo "  1) Direct high port - Shilka listens publicly on the selected port."
  echo "  2) Reverse proxy - Shilka listens on 127.0.0.1 and nginx/Caddy handles public TLS."
  read -rp "Mode [1]: " input_exposure
  if [[ "${input_exposure}" == "2" || "${input_exposure,,}" == "reverse" || "${input_exposure,,}" == "proxy" ]]; then
    PANEL_EXPOSURE="reverse_proxy"
  fi

  # Detect whether this is a domain (eligible for Let's Encrypt) or a bare IP.
  USE_ACME=false
  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    echo ""
    echo "Reverse proxy mode selected. Panel TLS will be off and the panel will bind to 127.0.0.1."
  elif is_domain "${PANEL_HOST}"; then
    echo ""
    echo "This looks like a domain. Let's Encrypt can provide a trusted certificate."
    if port_in_use 80; then
      echo "WARNING: port 80 is already in use. Standalone Let's Encrypt may disrupt the existing service or fail."
      read -rp "Use Let's Encrypt anyway? [y/N] " use_le
      if [[ "${use_le,,}" == "y" || "${use_le,,}" == "yes" ]]; then
        USE_ACME=true
        CERT_TYPE="domain"
      fi
    else
      read -rp "Use Let's Encrypt? [Y/n] " use_le
      if [[ "${use_le,,}" != "n" && "${use_le,,}" != "no" ]]; then
        USE_ACME=true
        CERT_TYPE="domain"
      fi
    fi
  elif is_ipv4 "${PANEL_HOST}"; then
    echo ""
    echo "Let's Encrypt supports IP addresses via the shortlived profile (6-day validity, auto-renews)."
    if port_in_use 80; then
      echo "WARNING: port 80 is already in use. Standalone Let's Encrypt may disrupt the existing service or fail."
      read -rp "Use Let's Encrypt for this IP anyway? [y/N] " use_le
      if [[ "${use_le,,}" == "y" || "${use_le,,}" == "yes" ]]; then
        USE_ACME=true
        CERT_TYPE="ip"
      fi
    else
      read -rp "Use Let's Encrypt for this IP? [Y/n] " use_le
      if [[ "${use_le,,}" != "n" && "${use_le,,}" != "no" ]]; then
        USE_ACME=true
        CERT_TYPE="ip"
      fi
    fi
  fi

  if [[ "${USE_ACME}" == "true" ]]; then
    echo ""
    echo "Enter your email address for Let's Encrypt notifications."
    read -rp "Email: " input_email
    ACME_EMAIL="${input_email}"
    if [[ -z "${ACME_EMAIL}" ]]; then
      echo "Email is required for Let's Encrypt."
      exit 1
    fi
    echo ""
    echo "Make sure port 80 is reachable from the internet and not in use."
    echo "(acme.sh needs it briefly for HTTP-01 validation)"
  fi

  # Port
  PANEL_PORT="${PANEL_PORT:-$(random_port)}"
  echo ""
  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    echo "Panel local port for the reverse proxy target (must be free on 127.0.0.1)."
  else
    echo "Panel public port (must be free)."
  fi
  read -rp "Port [${PANEL_PORT}]: " input_port
  PANEL_PORT="${input_port:-${PANEL_PORT}}"
  if port_in_use "${PANEL_PORT}"; then
    echo "ERROR: port ${PANEL_PORT} is already in use"
    exit 1
  fi

  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    PANEL_LISTEN_ADDRESS="127.0.0.1:${PANEL_PORT}"
  else
    PANEL_LISTEN_ADDRESS=":${PANEL_PORT}"
  fi

  # Base path
  local default_path
  default_path="$(random_base_path)"
  echo ""
  echo "Web panel path prefix for obscurity (starts with /)."
  read -rp "Path [${default_path}]: " input_path
  PANEL_PATH="${input_path:-${default_path}}"
  # Ensure leading /
  [[ "${PANEL_PATH}" != /* ]] && PANEL_PATH="/${PANEL_PATH}"

  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
    PANEL_PUBLIC_URL="https://${PANEL_HOST}${PANEL_PATH}"
  else
    PANEL_PUBLIC_URL="https://${PANEL_HOST}:${PANEL_PORT}${PANEL_PATH}"
  fi

  # Admin username
  local default_user
  default_user="$(random_username)"
  echo ""
  echo "Admin username."
  read -rp "Username [${default_user}]: " input_user
  ADMIN_USER="${input_user:-${default_user}}"

  # Admin password
  local default_pass
  default_pass="$(openssl rand -base64 18)"
  echo ""
  echo "Admin password (leave empty for auto-generated)."
  read -rp "Password [auto]: " input_pass
  ADMIN_PASSWORD="${input_pass:-${default_pass}}"

  echo ""
  echo "--------------------------------------------"
  echo "Summary:"
  echo "  Domain/IP:  ${PANEL_HOST}"
  echo "  Listen:     ${PANEL_LISTEN_ADDRESS}"
  echo "  Public URL: ${PANEL_PUBLIC_URL}"
  echo "  Path:       ${PANEL_PATH}"
  echo "  TLS:        $(if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then echo "off (reverse proxy)"; elif [[ "${USE_ACME}" == "true" ]]; then echo "Let's Encrypt (acme.sh)"; else echo "self-signed"; fi)"
  echo "  Username:   ${ADMIN_USER}"
  echo "  Password:   ${ADMIN_PASSWORD}"
  echo "--------------------------------------------"
  echo ""
  read -rp "Proceed with installation? [Y/n] " confirm
  if [[ "${confirm,,}" == "n" || "${confirm,,}" == "no" ]]; then
    echo "Installation cancelled."
    exit 0
  fi
}

write_prod_config() {
  local jwt_secret clash_secret tls_mode tls_cert_file tls_key_file self_signed_hosts
  jwt_secret="$(openssl rand -hex 32)"
  clash_secret="$(openssl rand -hex 24)"

  if [[ "${PANEL_EXPOSURE}" == "reverse_proxy" ]]; then
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

CONFIG_DIR="${CONFIG_DIR:-/etc/shilka}"
DATA_DIR="${DATA_DIR:-/var/lib/shilka}"
BIN="${BIN:-/opt/shilka/bin/shilka}"
export SHILKA_CONFIG_PATH="${SHILKA_CONFIG_PATH:-${CONFIG_DIR}/prod.yaml}"

if [[ "$#" -gt 0 ]]; then
  exec "${BIN}" "$@"
fi

start_services()   { systemctl start shilka.service; }
stop_services()    { systemctl stop shilka.service; }
restart_services() { systemctl restart shilka.service; }

reset_admin_password() {
  local password
  read -rsp "New admin password: " password
  echo
  if [[ -z "${password}" ]]; then
    echo "Password cannot be empty"
    exit 1
  fi
  "${BIN}" admin reset-password -password "${password}"
  echo "${password}" >"${CONFIG_DIR}/initial-admin-password"
  chmod 0600 "${CONFIG_DIR}/initial-admin-password"
  systemctl restart shilka.service
  echo "Password reset. Log in with your admin username and the new password."
}

change_panel_port() {
  local port
  read -rp "New panel port: " port
  if ! [[ "${port}" =~ ^[0-9]+$ ]] || (( port < 1 || port > 65535 )); then
    echo "Invalid port"
    return 1
  fi
  "${BIN}" setting -port "${port}"
  systemctl restart shilka.service
}

set_domain() {
  local domain
  read -rp "Panel domain: " domain
  if [[ -z "${domain}" ]]; then
    echo "Domain cannot be empty"
    return 1
  fi
  "${BIN}" setting -domain "${domain}" -public-url "https://${domain}"
  systemctl restart shilka.service
}

create_api_token() {
  local name
  read -rp "Token name [node]: " name
  name="${name:-node}"
  "${BIN}" api-token create -name "${name}" -scopes node
}

set_custom_cert() {
  local cert key
  read -rp "Certificate path: " cert
  read -rp "Key path: " key
  "${BIN}" cert set-files -cert "${cert}" -key "${key}"
  systemctl restart shilka.service
}

disable_panel_tls() {
  "${BIN}" cert reset
  systemctl restart shilka.service
}

backup_db() {
  local out="${DATA_DIR}/panel-$(date +%Y%m%d-%H%M%S).db"
  cp -a "${DATA_DIR}/panel.db" "${out}"
  echo "Backup: ${out}"
}

core_reload() {
  "${BIN}" core reload
}

show_menu() {
  cat <<'MENU'
1. Start   2. Stop   3. Restart   4. Reset admin password
5. Change port   6. Status   7. Logs   8. Create API token
9. Set domain/public URL   10. Set custom TLS cert   11. Disable TLS
12. Core reload   13. Backup DB   0. Exit
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
      4) reset_admin_password ;;
      5) change_panel_port ;;
      6) systemctl status shilka.service --no-pager ;;
      7) journalctl -u shilka.service -n 120 --no-pager ;;
      8) create_api_token ;;
      9) set_domain ;;
      10) set_custom_cert ;;
      11) disable_panel_tls ;;
      12) core_reload ;;
      13) backup_db ;;
      0) exit 0 ;;
      *) echo "Unknown option" ;;
    esac
  done
}
main "$@"
SCRIPT

  chmod 0755 /usr/local/bin/shilka
  systemctl daemon-reload
  systemctl enable --now shilka.service
}

main() {
  require_root
  gather_input
  create_user_and_dirs
  if [[ "${USE_ACME}" == "true" ]]; then
    install_acme
    if [[ "${CERT_TYPE}" == "domain" ]]; then
      issue_domain_cert "${PANEL_HOST}"
    else
      issue_ip_cert "${PANEL_HOST}"
    fi
  fi
  install_sing_box
  install_panel_binary
  install_update_helper
  write_prod_config
  install_systemd

  echo ""
  echo "===== Shilka installed ====="
  echo "URL:      ${PANEL_PUBLIC_URL}"
  echo "Username: ${ADMIN_USER}"
  echo "Password: ${ADMIN_PASSWORD}"
  echo ""
  echo "Manage: shilka"
  echo "Logs:   journalctl -u shilka.service -f"
  echo "Config: ${CONFIG_DIR}/prod.yaml"
  echo "==============================="
}

main "$@"
