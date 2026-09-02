#!/usr/bin/env bash
# ============================================================
# Nimbo · Nube Casera
# Script principal de inicio.
#
# Uso:
#   ./scripts/start.sh start          # levanta con Docker (por defecto)
#   ./scripts/start.sh stop           # detiene
#   ./scripts/start.sh restart        # reinicia
#   ./scripts/start.sh status         # estado del servicio
#   ./scripts/start.sh logs           # sigue los logs en vivo
#   ./scripts/start.sh backup         # crea un backup
#   ./scripts/start.sh help           # esta ayuda
#
# Opciones adicionales:
#   --local   ejecuta el backend directo con Go (sin Docker).
#   PORT=9090 ./scripts/start.sh start  # cambia el puerto en modo local
# ============================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
PIDFILE="$ROOT_DIR/backend/.nimbo.pid"
BIN="$ROOT_DIR/backend/nimbo-bin"
MODE="docker"

# Detectar la opción --local en cualquier posición.
for arg in "$@"; do
  [ "$arg" = "--local" ] && MODE="local"
done

COMMAND="${1:-help}"
shift 2>/dev/null || true

log() { printf '\033[1;34m[Nimbo]\033[0m %s\n' "$*"; }
ok()  { printf '\033[1;32m[Nimbo]\033[0m %s\n' "$*"; }
err() { printf '\033[1;31m[Nimbo]\033[0m %s\n' "$*" >&2; }

# ------------------------------------------------------------
# Helpers compartidos
# ------------------------------------------------------------

# Crea las carpetas locales que usa Docker. Hacerlo ANTES de levantar
# evita que Docker las cree como root (lo que impediría escribir dentro
# del contenedor en un clon recién bajado de GitHub).
ensure_dirs() {
  mkdir -p "$ROOT_DIR/data" "$ROOT_DIR/logs" "$ROOT_DIR/config" "$ROOT_DIR/backups"
}

# Detecta la zona horaria del host para mostrarla igual dentro del
# contenedor. Fallback: variable TZ, luego UTC.
detect_tz() {
  if [ -f /etc/timezone ]; then            # Debian / Ubuntu
    cat /etc/timezone
    return 0
  fi
  if [ -L /etc/localtime ] || [ -f /etc/localtime ]; then
    local target
    target="$(readlink -f /etc/localtime 2>/dev/null)"
    if [ -n "$target" ]; then
      echo "${target#/usr/share/zoneinfo/}"
      return 0
    fi
  fi
  if [ -n "${TZ:-}" ]; then
    echo "$TZ"
    return 0
  fi
  echo "UTC"
}

# Ejecuta un comando de Docker Compose (soporta plugin v2 'docker compose'
# y la versión v1 'docker-compose'). Exporta las variables que usa el
# compose para correr el contenedor con el usuario y la zona horaria
# correctos.
dc() {
  export NIMBO_UID="$(id -u)"
  export NIMBO_GID="$(id -g)"
  export TZ="$(detect_tz)"

  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    docker compose -f "$COMPOSE_FILE" "$@"
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose -f "$COMPOSE_FILE" "$@"
  else
    err "No se encontró Docker Compose."
    err "Instala el plugin: https://docs.docker.com/compose/install/"
    return 1
  fi
}

# ------------------------------------------------------------
# Modo Docker
# ------------------------------------------------------------
start_docker() {
  ensure_dirs
  log "Compilando y levantando Nimbo con Docker Compose..."
  dc up -d --build
  ok "Nimbo disponible en http://localhost:8080"
}

stop_docker() {
  log "Deteniendo Nimbo..."
  dc down
  ok "Nimbo detenido."
}

status_docker() {
  dc ps
}

logs_docker() {
  dc logs -f --tail=50
}

restart_docker() {
  ensure_dirs
  log "Reiniciando Nimbo..."
  dc down
  dc up -d --build
  ok "Nimbo reiniciado: http://localhost:8080"
}

# ------------------------------------------------------------
# Modo local (Go directo, sin Docker)
# ------------------------------------------------------------
build_local() {
  log "Compilando el backend (Go)..."
  (cd "$ROOT_DIR/backend" && go build -o "$BIN" ./cmd/main.go)
}

start_local() {
  ensure_dirs

  if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
    err "Nimbo ya está corriendo (PID $(cat "$PIDFILE")). Usa 'restart' primero."
    exit 1
  fi

  build_local

  export STORAGE_PATH="$ROOT_DIR/data"
  export CONFIG_PATH="$ROOT_DIR/config/config.json"
  export LOG_PATH="$ROOT_DIR/logs/activity.log"
  export FRONTEND_PATH="$ROOT_DIR/frontend"
  export PORT="${PORT:-8080}"
  export MAX_FILE_SIZE="${MAX_FILE_SIZE:-1073741824}"

  log "Iniciando Nimbo en modo local (puerto $PORT)..."
  nohup "$BIN" >> "$ROOT_DIR/logs/stdout.log" 2>&1 &
  PID=$!
  echo "$PID" > "$PIDFILE"

  # Verificar que el proceso siga vivo (pudo fallar el bind del puerto).
  sleep 1
  if ! kill -0 "$PID" 2>/dev/null; then
    err "Nimbo no pudo iniciar. Revisa si el puerto $PORT ya está en uso:"
    tail -3 "$ROOT_DIR/logs/stdout.log" >&2
    rm -f "$PIDFILE"
    exit 1
  fi
  ok "Nimbo corriendo con PID $PID — http://localhost:$PORT"
  log "Logs: $ROOT_DIR/logs/activity.log"
}

stop_local() {
  if [ ! -f "$PIDFILE" ]; then
    log "No hay proceso guardado. Nada que detener."
    return
  fi
  PID="$(cat "$PIDFILE")"
  if kill -0 "$PID" 2>/dev/null; then
    log "Deteniendo Nimbo (PID $PID)..."
    kill "$PID"
    sleep 1
    ok "Nimbo detenido."
  else
    log "El proceso ya no estaba activo."
  fi
  rm -f "$PIDFILE"
}

status_local() {
  if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
    ok "Nimbo corriendo (PID $(cat "$PIDFILE"))."
  else
    err "Nimbo no está corriendo."
    exit 1
  fi
}

logs_local() {
  tail -f "$ROOT_DIR/logs/activity.log"
}

# ------------------------------------------------------------
# Comandos
# ------------------------------------------------------------
case "$COMMAND" in
  start)
    if [ "$MODE" = "local" ]; then start_local; else start_docker; fi
    ;;
  stop)
    if [ "$MODE" = "local" ]; then stop_local; else stop_docker; fi
    ;;
  restart)
    if [ "$MODE" = "local" ]; then
      stop_local || true
      start_local
    else
      restart_docker
    fi
    ;;
  status)
    if [ "$MODE" = "local" ]; then status_local; else status_docker; fi
    ;;
  logs)
    if [ "$MODE" = "local" ]; then logs_local; else logs_docker; fi
    ;;
  backup)
    "$ROOT_DIR/scripts/backup.sh"
    ;;
  help|*)
    sed -n '2,18p' "$0" | sed 's/^# \{0,1\}//'
    ;;
esac
