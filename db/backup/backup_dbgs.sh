#!/usr/bin/env bash
# ==============================================================================
# Script: backup_dbgs.sh
# Descripción: Automatización de respaldos periódicos comprimidos para DBGS.
# USO: ./backup_dbgs.sh [/ruta/destino/backup]
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

# Configuración por defecto (modificable mediante variables de entorno)
DB_HOST="${DB_HOST:-$(read_config_value host localhost)}"
DB_PORT="${DB_PORT:-$(read_config_value port 5432)}"
DB_USER="${DB_USER:-$(read_config_value user postgres)}"
DB_PASSWORD="${DB_PASSWORD:-$(read_config_value password postgres)}"
DB_NAME="${DB_NAME:-$(read_config_value name dbgs_soberano)}"
BACKUP_DIR="${1:-${REPO_ROOT}/db/backup/dumps}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/dbgs_backup_${TIMESTAMP}.dump"

echo "=================================================="
echo " Starting DBGS Database Backup Process"
echo " Date: $(date)"
echo " Target Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"
echo "=================================================="

# 1. Validar binarios requeridos
if ! command -v pg_dump &> /dev/null; then
    echo "ERROR: 'pg_dump' is not installed or not in PATH." >&2
    exit 1
fi

# 2. Crear directorio de respaldo si no existe
mkdir -p "${BACKUP_DIR}"

# 3. Ejecutar el volcado de la base de datos
echo "--> Generating backup file: ${BACKUP_FILE}..."

export PGPASSWORD="${DB_PASSWORD}"

if pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -F c \
    -b \
    -v \
    -f "${BACKUP_FILE}" \
    "${DB_NAME}"; then
    
    FILE_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    echo "SUCCESS: Backup completed successfully! (Size: ${FILE_SIZE})"
else
    echo "ERROR: Database backup failed!" >&2
    exit 1
fi

# 4. Limpieza de respaldos antiguos según política de retención
echo "--> Cleaning up backups older than ${RETENTION_DAYS} days..."
find "${BACKUP_DIR}" -name "dbgs_backup_*.dump" -type f -mtime +"${RETENTION_DAYS}" -exec rm -vf {} \;

echo "--> Backup process finished successfully."
exit 0