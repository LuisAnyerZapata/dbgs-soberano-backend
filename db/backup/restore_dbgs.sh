#!/usr/bin/env bash
# ==============================================================================
# Script: restore_dbgs.sh
# Descripción: Automatización del proceso de restauración desde un archivo .dump
# USO: ./restore_dbgs.sh /ruta/al/archivo_backup.dump
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CONFIG_FILE="${DB_CONFIG_FILE:-${REPO_ROOT}/config/config.json}"

read_config_value() {
    local key="$1"
    local default="$2"
    local py_cmd="${PY_BIN:-}"
    if [ -z "${py_cmd}" ]; then
        if command -v python3 >/dev/null 2>&1; then
            py_cmd="python3"
        elif command -v python >/dev/null 2>&1; then
            py_cmd="python"
        fi
    fi
    if [ -f "${CONFIG_FILE}" ] && [ -n "${py_cmd}" ]; then
        "${py_cmd}" - "${CONFIG_FILE}" "${key}" "${default}" <<'PY'
import json, sys
config_path, key, default = sys.argv[1], sys.argv[2], sys.argv[3]
try:
    with open(config_path, 'r', encoding='utf-8') as fh:
        data = json.load(fh)
except Exception:
    print(default)
    sys.exit(0)
value = data.get('database', {}).get(key, default)
print(value)
PY
    else
        echo "${default}"
    fi
}

# Configuración por defecto (override con variables de entorno)
DB_HOST="${DB_HOST:-$(read_config_value host localhost)}"
DB_PORT="${DB_PORT:-$(read_config_value port 5432)}"
DB_USER="${DB_USER:-$(read_config_value user postgres)}"
DB_PASSWORD="${DB_PASSWORD:-$(read_config_value password postgres)}"
DB_NAME="${DB_NAME:-$(read_config_value name dbgs_soberano)}"

BACKUP_FILE="${1:-}"
AUTO_APPROVE="${AUTO_APPROVE:-0}"
echo "=================================================="

# 1. Validaciones iniciales
if [ -z "${BACKUP_FILE}" ]; then
    echo "ERROR: You must specify a backup file." >&2
    echo "Usage: $0 /path/to/backup_file.dump" >&2
    exit 1
fi

if [ ! -f "${BACKUP_FILE}" ]; then
    echo "ERROR: Backup file not found: ${BACKUP_FILE}" >&2
    exit 1
fi

if ! command -v pg_restore &> /dev/null; then
    echo "ERROR: 'pg_restore' is not installed or not in PATH." >&2
    exit 1
fi

# 2. Confirmación de seguridad en entornos interactivos
if [ "${AUTO_APPROVE}" != "1" ] && [ -t 0 ]; then
    read -rp "WARNING: This process will restore data into '${DB_NAME}'. Continue? (y/N): " CONFIRM
    if [[ ! "${CONFIRM}" =~ ^[Yy]$ ]]; then
        echo "Restoration aborted by user."
        exit 0
    fi
fi

# 3. Ejecución de la restauración
echo "--> Restoring database from: ${BACKUP_FILE}..."

# --clean: Elimina objetos antes de recrearlos
# --if-exists: Previene errores si no existen objetos al limpiar
# --no-owner: Evita errores de asignación de propietarios de objetos
export PGPASSWORD="${DB_PASSWORD}"

if pg_restore \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --clean \
    --if-exists \
    --no-owner \
    -v \
    "${BACKUP_FILE}"; then
    
    echo "SUCCESS: Database restoration finished successfully!"
else
    # Propagamos el fallo al llamador (la capa Go distingue FALLIDO de COMPLETADO)
    echo "ERROR: Database restoration failed!" >&2
    exit 1
fi