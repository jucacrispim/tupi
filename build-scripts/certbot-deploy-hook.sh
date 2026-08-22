#!/usr/bin/env bash
set -euo pipefail

LINEAGE="${RENEWED_LINEAGE:?}"
TUPI_USER="${TUPI_USER:-tupi}"
TUPI_GROUP="${TUPI_GROUP:-tupi}"

chown "${TUPI_USER}:${TUPI_GROUP}" "${LINEAGE}/privkey.pem"
chmod 640 "${LINEAGE}/privkey.pem"
chmod 644 "${LINEAGE}/fullchain.pem"

if [[ -n "${TUPI_PID_FILE:-}" ]]; then
    kill -HUP "$(cat "${TUPI_PID_FILE}")"
elif command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet tupi; then
    systemctl reload tupi
else
    pkill -HUP -x tupi || true
fi
