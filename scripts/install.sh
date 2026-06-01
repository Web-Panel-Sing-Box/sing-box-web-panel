#!/usr/bin/env bash
set -Eeuo pipefail

APP_USER="${APP_USER:-shilka}"
APP_HOME="${APP_HOME:-/opt/shilka}"
CONFIG_DIR="${CONFIG_DIR:-/etc/shilka}"
DATA_DIR="${DATA_DIR:-/var/lib/shilka}"
LOG_DIR="${LOG_DIR:-/var/log/shilka}"
SING_BOX_VERSION="${SING_BOX_VERSION:-latest}"
PANEL_VERSION="${PANEL_VERSION:-latest}"
GITHUB_REPO="${GITHUB_REPO:-Web-Panel-Sing-Box/sing-box-web-panel}"

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

resolve_sing_box_version() {
  if [[ "${SING_BOX_VERSION}" != "latest" ]]; then
    echo "${SING_BOX_VERSION#v}"
    return
  fi
  curl -fsSL https://api.github.com/repos/SagerNet/sing-box/releases/latest \
    | sed -n 's/.*"tag_name": "v\{0,1\}\([^"]*\)".*/\1/p' \
    | head -n 1
}

create_user_and_dirs() {
  if ! id "${APP_USER}" >/dev/null 2>&1; then
    useradd --system --home "${APP_HOME}" --shell /usr/sbin/nologin "${APP_USER}"
  fi
  install -d -m 0755 "${APP_HOME}" "${APP_HOME}/bin" "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"
  install -d -m 0700 "${CONFIG_DIR}/tls"
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
  curl -fL "${url}" -o "${APP_HOME}/bin/shilka"
  chmod 0755 "${APP_HOME}/bin/shilka"
}

write_prod_config() {
  local jwt_secret clash_secret admin_password
  jwt_secret="$(openssl rand -hex 32)"
  clash_secret="$(openssl rand -hex 24)"
  admin_password="$(openssl rand -base64 18)"

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
  address: "127.0.0.1:8080"
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
  admin_user: "admin"
  admin_password: "${admin_password}"
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
  mode: "file"
  cert_file: "${CONFIG_DIR}/tls/cert.pem"
  key_file: "${CONFIG_DIR}/tls/key.pem"
  acme_email: ""
  acme_domains: []
  acme_cache_dir: "${DATA_DIR}/acme"
  self_signed_hosts: []
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

subscription:
  public_url: ""
  token_ttl: "720h"
YAML

  chmod 0640 "${CONFIG_DIR}/prod.yaml"
  chown "${APP_USER}:${APP_USER}" "${CONFIG_DIR}/prod.yaml"

  echo "${admin_password}" >"${CONFIG_DIR}/initial-admin-password"
  chmod 0600 "${CONFIG_DIR}/initial-admin-password"
  chown "${APP_USER}:${APP_USER}" "${CONFIG_DIR}/initial-admin-password"
}

install_systemd() {
  cat >/etc/systemd/system/shilka.service <<'UNIT'
[Unit]
Description=Shilka web panel
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=shilka
Group=shilka
WorkingDirectory=/opt/shilka
Environment=SHILKA_CONFIG_PATH=/etc/shilka/prod.yaml
ExecStart=/opt/shilka/bin/shilka
Restart=on-failure
RestartSec=3
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ReadWritePaths=/etc/shilka /var/lib/shilka /var/log/shilka
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
UNIT

  cat >/usr/local/bin/shilka <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail

CONFIG_DIR="${CONFIG_DIR:-/etc/shilka}"

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
  systemctl stop shilka.service
  local tmp="$(mktemp)"
  sed "s/^  admin_password: .*/  admin_password: \"${password}\"/" "${CONFIG_DIR}/prod.yaml" >"${tmp}"
  mv "${tmp}" "${CONFIG_DIR}/prod.yaml"
  echo "${password}" >"${CONFIG_DIR}/initial-admin-password"
  chmod 0600 "${CONFIG_DIR}/initial-admin-password"
  rm -f /var/lib/shilka/panel.db /var/lib/shilka/panel.db-wal /var/lib/shilka/panel.db-shm
  systemctl start shilka.service
  echo "Password reset. Log in with username 'admin' and the new password."
}

change_panel_port() {
  local port
  read -rp "New panel port: " port
  if ! [[ "${port}" =~ ^[0-9]+$ ]] || (( port < 1 || port > 65535 )); then
    echo "Invalid port"
    return 1
  fi
  sed -i.bak "s/^  address: \"127.0.0.1:.*$/  address: \"127.0.0.1:${port}\"/" "${CONFIG_DIR}/prod.yaml"
  systemctl restart shilka.service
}

show_menu() {
  cat <<'MENU'
1. Start   2. Stop   3. Restart   4. Reset admin password
5. Change port   6. Status   7. Logs   0. Exit
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
  create_user_and_dirs
  install_sing_box
  install_panel_binary
  write_prod_config
  install_systemd

  echo ""
  echo "===== Shilka installed ====="
  echo "Panel: http://<server-ip>:8080 (behind nginx/caddy for HTTPS)"
  echo "Admin: admin"
  echo "Password: $(cat "${CONFIG_DIR}/initial-admin-password")"
  echo ""
  echo "Manage: shilka"
  echo "Logs:   journalctl -u shilka.service -f"
  echo "Config: ${CONFIG_DIR}/prod.yaml"
  echo "==============================="
}

main "$@"
