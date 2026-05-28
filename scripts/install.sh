#!/usr/bin/env bash
set -Eeuo pipefail

APP_USER="${APP_USER:-sing-grok}"
APP_HOME="${APP_HOME:-/opt/sing-grok}"
CONFIG_DIR="${CONFIG_DIR:-/etc/sing-grok}"
DATA_DIR="${DATA_DIR:-/var/lib/sing-grok}"
LOG_DIR="${LOG_DIR:-/var/log/sing-grok}"
PANEL_PORT="${PANEL_PORT:-3000}"
API_PORT="${API_PORT:-8081}"
SING_BOX_VERSION="${SING_BOX_VERSION:-latest}"
REPO_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "install.sh must run as root"
    exit 1
  fi
}

install_packages() {
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update
    apt-get install -y curl tar ca-certificates python3 python3-venv nodejs npm rsync openssl
  elif command -v dnf >/dev/null 2>&1; then
    dnf install -y curl tar ca-certificates python3 nodejs npm rsync openssl
  elif command -v pacman >/dev/null 2>&1; then
    pacman -Sy --noconfirm curl tar ca-certificates python python-virtualenv nodejs npm rsync openssl
  else
    echo "Unsupported package manager. Install curl, tar, python3, node, npm, rsync, openssl manually."
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

install_sing_box() {
  local arch version asset url tmp
  arch="$(detect_arch)"
  if [[ "${arch}" == "unsupported" ]]; then
    echo "Unsupported CPU architecture: $(uname -m)"
    exit 1
  fi
  version="$(resolve_sing_box_version)"
  asset="sing-box-${version}-linux-${arch}.tar.gz"
  url="https://github.com/SagerNet/sing-box/releases/download/v${version}/${asset}"
  tmp="$(mktemp -d)"
  curl -fL "${url}" -o "${tmp}/${asset}"
  tar -xzf "${tmp}/${asset}" -C "${tmp}"
  install -m 0755 "${tmp}/sing-box-${version}-linux-${arch}/sing-box" "${APP_HOME}/bin/sing-box"
}

create_user_and_dirs() {
  if ! id "${APP_USER}" >/dev/null 2>&1; then
    useradd --system --home "${APP_HOME}" --shell /usr/sbin/nologin "${APP_USER}"
  fi
  install -d -m 0755 "${APP_HOME}" "${APP_HOME}/bin" "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"
  chown -R "${APP_USER}:${APP_USER}" "${APP_HOME}" "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"
}

write_env() {
  local jwt_secret clash_secret admin_password
  jwt_secret="$(openssl rand -hex 32)"
  clash_secret="$(openssl rand -hex 24)"
  admin_password="$(openssl rand -base64 18)"
  cat >"${CONFIG_DIR}/backend.env" <<EOF
SING_GROK_API_HOST=127.0.0.1
SING_GROK_API_PORT=${API_PORT}
SING_GROK_DATABASE_URL=sqlite+aiosqlite:////var/lib/sing-grok/panel.db
SING_GROK_CONFIG_DIR=${CONFIG_DIR}
SING_GROK_DATA_DIR=${DATA_DIR}
SING_GROK_LOG_DIR=${LOG_DIR}
SING_GROK_SING_BOX_BINARY=${APP_HOME}/bin/sing-box
SING_GROK_SING_BOX_CONFIG_PATH=${CONFIG_DIR}/config.json
SING_GROK_SING_BOX_LOG_PATH=${LOG_DIR}/sing-box.log
SING_GROK_JWT_SECRET=${jwt_secret}
SING_GROK_CLASH_API_SECRET=${clash_secret}
SING_GROK_BOOTSTRAP_ADMIN_USERNAME=admin
SING_GROK_BOOTSTRAP_ADMIN_PASSWORD=${admin_password}
EOF
  cat >"${CONFIG_DIR}/web.env" <<EOF
HOSTNAME=0.0.0.0
PORT=${PANEL_PORT}
SING_GROK_API_BASE_URL=http://127.0.0.1:${API_PORT}
EOF
  chmod 0640 "${CONFIG_DIR}/backend.env" "${CONFIG_DIR}/web.env"
  chown "${APP_USER}:${APP_USER}" "${CONFIG_DIR}/backend.env" "${CONFIG_DIR}/web.env"
  echo "${admin_password}" >"${CONFIG_DIR}/initial-admin-password"
  chmod 0600 "${CONFIG_DIR}/initial-admin-password"
  chown "${APP_USER}:${APP_USER}" "${CONFIG_DIR}/initial-admin-password"
}

write_minimal_sing_box_config() {
  local clash_secret
  clash_secret="$(sed -n 's/^SING_GROK_CLASH_API_SECRET=//p' "${CONFIG_DIR}/backend.env")"
  cat >"${CONFIG_DIR}/config.json" <<EOF
{
  "log": { "level": "info", "timestamp": true },
  "inbounds": [],
  "outbounds": [{ "type": "direct", "tag": "direct" }],
  "route": { "final": "direct" },
  "experimental": {
    "cache_file": { "enabled": true, "path": "${DATA_DIR}/sing-box-cache.db" },
    "clash_api": {
      "external_controller": "127.0.0.1:9090",
      "secret": "${clash_secret}",
      "access_control_allow_origin": ["http://127.0.0.1", "http://localhost"],
      "access_control_allow_private_network": false
    }
  }
}
EOF
  chmod 0640 "${CONFIG_DIR}/config.json"
  chown "${APP_USER}:${APP_USER}" "${CONFIG_DIR}/config.json"
}

install_app() {
  rsync -a --delete --exclude .git --exclude node_modules --exclude .next "${REPO_ROOT}/" "${APP_HOME}/"
  chown -R "${APP_USER}:${APP_USER}" "${APP_HOME}"
  sudo -u "${APP_USER}" python3 -m venv "${APP_HOME}/backend/.venv"
  sudo -u "${APP_USER}" "${APP_HOME}/backend/.venv/bin/python" -m pip install --upgrade pip
  sudo -u "${APP_USER}" "${APP_HOME}/backend/.venv/bin/python" -m pip install -e "${APP_HOME}/backend"
  cd "${APP_HOME}/frontend"
  sudo -u "${APP_USER}" npm install
  sudo -u "${APP_USER}" npm run build
  local env_args=()
  while IFS= read -r line; do
    [[ -z "${line}" || "${line}" == \#* ]] && continue
    env_args+=("${line}")
  done < "${CONFIG_DIR}/backend.env"
  sudo -u "${APP_USER}" env "${env_args[@]}" \
    "${APP_HOME}/backend/.venv/bin/python" -m app.cli reset-admin \
    --username admin \
    --password "$(cat "${CONFIG_DIR}/initial-admin-password")"
}

install_systemd() {
  install -m 0644 "${APP_HOME}/systemd/sing-grok-api.service" /etc/systemd/system/sing-grok-api.service
  install -m 0644 "${APP_HOME}/systemd/sing-grok-web.service" /etc/systemd/system/sing-grok-web.service
  install -m 0644 "${APP_HOME}/systemd/sing-grok-singbox.service" /etc/systemd/system/sing-grok-singbox.service
  install -m 0755 "${APP_HOME}/scripts/sing-grok" /usr/local/bin/sing-grok
  systemctl daemon-reload
  systemctl enable --now sing-grok-api.service sing-grok-web.service sing-grok-singbox.service
}

main() {
  require_root
  install_packages
  create_user_and_dirs
  install_sing_box
  write_env
  write_minimal_sing_box_config
  install_app
  install_systemd
  echo "Sing Grok installed on port ${PANEL_PORT}"
  echo "Initial admin password: $(cat "${CONFIG_DIR}/initial-admin-password")"
}

main "$@"
