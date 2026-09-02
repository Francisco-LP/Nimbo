# ============================================================
# Nimbo · Imagen Docker (multi-stage build)
#
# Etapa 1: compila el binario de Go.
# Etapa 2: imagen liviana (Alpine) con el binario + frontend + config.
# ============================================================

FROM golang:1.21-alpine AS build

WORKDIR /src

# Copiar módulos y descargar dependencias (solo stdlib, así que es veloz).
COPY backend/go.mod backend/go.sum ./
RUN go mod download || true

# Copiar el código del backend y compilar.
COPY backend/ ./backend/
WORKDIR /src/backend
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/nimbo ./cmd/main.go

# ------------------------------------------------------------
# Imagen final
# ------------------------------------------------------------
FROM alpine:3.19

# tzdata: sin ella, musl ignora la variable TZ y todo queda en UTC.
# El script start.sh detecta la zona horaria del host y la inyecta.
RUN apk add --no-cache tzdata \
    && adduser -D -u 1000 nimbo \
    && mkdir -p /app/frontend /data /logs /config \
    && chown -R nimbo:nimbo /app /data /logs /config

WORKDIR /app

COPY --from=build /bin/nimbo /usr/local/bin/nimbo
COPY frontend/ /app/frontend/
COPY config/config.json /config/config.json

USER nimbo

# Variables de entorno con los valores por defecto.
ENV STORAGE_PATH=/data \
    CONFIG_PATH=/config/config.json \
    LOG_PATH=/logs/activity.log \
    PORT=8080 \
    MAX_FILE_SIZE=1073741824 \
    FRONTEND_PATH=/app/frontend

EXPOSE 8080
VOLUME ["/data", "/logs", "/config"]

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/api/health || exit 1

ENTRYPOINT ["nimbo"]
