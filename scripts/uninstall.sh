#!/usr/bin/env bash
set -Eeuo pipefail

APP_USER="${APP_USER:-shilka}"
APP_HOME="${APP_HOME:-/opt/shilka}"
CONFIG_DIR="${CONFIG_DIR:-/etc/shilka}"
DATA_DIR="${DATA_DIR:-/var/lib/shilka}"
LOG_DIR="${LOG_DIR:-/var/log/shilka}"
UPDATE_SCRIPT_PATH="${UPDATE_SCRIPT_PATH:-/usr/local/sbin/shilka-update}"
UPDATE_SUDOERS_PATH="${UPDATE_SUDOERS_PATH:-/etc/sudoers.d/shilka-update}"

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "uninstall.sh must run as root"
    exit 1
  fi
}

stop_service() {
  if systemctl is-active --quiet shilka.service 2>/dev/null; then
    echo "Stopping shilka.service..."
    systemctl stop shilka.service
  fi
  if systemctl is-enabled --quiet shilka.service 2>/dev/null; then
    echo "Disabling shilka.service..."
    systemctl disable shilka.service
  fi
}

remove_unit() {
  if [[ -f /etc/systemd/system/shilka.service ]]; then
    echo "Removing systemd unit..."
    rm -f /etc/systemd/system/shilka.service
    systemctl daemon-reload
  fi
}

remove_cli() {
  if [[ -f /usr/local/bin/shilka ]]; then
    echo "Removing CLI script..."
    rm -f /usr/local/bin/shilka
  fi
  if [[ -f "${UPDATE_SCRIPT_PATH}" ]]; then
    echo "Removing update helper..."
    rm -f "${UPDATE_SCRIPT_PATH}"
  fi
  if [[ -f "${UPDATE_SUDOERS_PATH}" ]]; then
    echo "Removing update sudoers rule..."
    rm -f "${UPDATE_SUDOERS_PATH}"
  fi
}

remove_user() {
  if id "${APP_USER}" >/dev/null 2>&1; then
    echo "Removing user ${APP_USER}..."
    userdel -r "${APP_USER}" 2>/dev/null || userdel "${APP_USER}"
  fi
}

remove_dirs() {
  for dir in "${APP_HOME}" "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"; do
    if [[ -d "${dir}" ]]; then
      echo "Removing ${dir}..."
      rm -rf "${dir}"
    fi
  done
}

main() {
  require_root

  echo "============================================"
  echo "      Shilka Panel Uninstaller"
  echo "============================================"
  echo ""
  echo "This will remove the shilka panel, sing-box binary,"
  echo "all configuration, database, logs and the systemd unit."
  echo ""
  echo "Paths to be removed:"
  echo "  ${APP_HOME}"
  echo "  ${CONFIG_DIR}"
  echo "  ${DATA_DIR}"
  echo "  ${LOG_DIR}"
  echo "  /usr/local/bin/shilka"
  echo "  ${UPDATE_SCRIPT_PATH}"
  echo "  ${UPDATE_SUDOERS_PATH}"
  echo "  /etc/systemd/system/shilka.service"
  echo "  user: ${APP_USER}"
  echo ""
  read -rp "Proceed with uninstall? [y/N] " confirm
  if [[ "${confirm,,}" != "y" && "${confirm,,}" != "yes" ]]; then
    echo "Uninstall cancelled."
    exit 0
  fi

  stop_service
  remove_unit
  remove_cli
  remove_user
  remove_dirs

  echo ""
  echo "===== Shilka uninstalled ====="
}

main "$@"
