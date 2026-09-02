#!/usr/bin/env bash
# ============================================================
# Nimbo · Nube Casera
# Backup manual de data/, logs/ y config/.
#
# Uso: ./scripts/backup.sh
# Los backups se guardan en backups/ y se conservan los últimos 7.
# ============================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="$ROOT_DIR/backups"
STAMP="$(date +%Y%m%d_%H%M%S)"
OUT="$BACKUP_DIR/nimbo-backup-$STAMP.tar.gz"
KEEP=7

mkdir -p "$BACKUP_DIR"

# Incluir solo las carpetas que existan.
INCLUDE=()
[ -d "$ROOT_DIR/data" ]   && INCLUDE+=(data)
[ -d "$ROOT_DIR/logs" ]   && INCLUDE+=(logs)
[ -d "$ROOT_DIR/config" ] && INCLUDE+=(config)

if [ "${#INCLUDE[@]}" -eq 0 ]; then
  printf '[Nimbo] No hay data/, logs/ ni config/ para respaldar.\n' >&2
  exit 1
fi

tar -czf "$OUT" -C "$ROOT_DIR" "${INCLUDE[@]}"

# Limpiar backups viejos, conservando los últimos K.
mapfile -t OLD < <(ls -1t "$BACKUP_DIR"/nimbo-backup-*.tar.gz 2>/dev/null | tail -n +$((KEEP + 1)))
for f in "${OLD[@]}"; do
  rm -f "$f"
done

printf '[Nimbo] Backup creado: %s\n' "$OUT"
[ "${#OLD[@]}" -gt 0 ] && printf '[Nimbo] Se eliminaron %d backup(s) antiguo(s).\n' "${#OLD[@]}"
