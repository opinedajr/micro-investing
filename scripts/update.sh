#!/usr/bin/env bash
set -euo pipefail

APP_NAME="micro-investing"
APP_DIR="/opt/${APP_NAME}"
APP_USER="www-data"
SERVICE_NAME="${APP_NAME}.service"
BIN_PATH="${APP_DIR}/${APP_NAME}"

log() { printf '\033[1;34m[%s]\033[0m %s\n' "$(date +%H:%M:%S)" "$*"; }
err() { printf '\033[1;31m[ERROR]\033[0m %s\n' "$*" >&2; exit 1; }

[[ "$(uname)" == "Linux" ]] || err "Este script roda apenas em Linux (VPS)."
[[ $EUID -eq 0 ]] || err "Rode com sudo: sudo $0"
command -v systemctl >/dev/null || err "systemd não encontrado."

[[ -f "${BIN_PATH}" ]] || err "Binário não encontrado em ${BIN_PATH}. Faça o build nativo em Linux (o driver sqlite usa CGO) e copie a nova versão antes:
  make build
  scp bin/${APP_NAME} root@<vps>:/tmp/
  sudo install -m 0755 -o ${APP_USER} -g ${APP_USER} /tmp/${APP_NAME} ${BIN_PATH}"

[[ -f "${APP_DIR}/.env" ]] || err ".env ausente em ${APP_DIR}/.env — rode install.sh primeiro."

log "Ajustando permissões do binário"
chmod 0755 "${BIN_PATH}"
chown "${APP_USER}:${APP_USER}" "${BIN_PATH}"

log "Reiniciando ${SERVICE_NAME}"
systemctl restart "${SERVICE_NAME}"

log "Aguardando serviço subir..."
sleep 2

if systemctl is-active --quiet "${SERVICE_NAME}"; then
  log "OK — ${SERVICE_NAME} ativo"
  systemctl --no-pager --full status "${SERVICE_NAME}" || true
else
  err "Serviço não subiu. Diagnóstico:
  systemctl status ${SERVICE_NAME}
  journalctl -u ${SERVICE_NAME} -n 50 --no-pager"
fi
