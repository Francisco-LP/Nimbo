# ☁️ Nimbo — Nube Casera

Nimbo (del latín *nimbus*, "nube") es una aplicación de **nube personal** para red local. Permite subir, descargar, eliminar y gestionar archivos desde el navegador web, con un diseño dark-mode suave y un enfoque 100% educativo.

- **Backend:** Go (solo biblioteca estándar, sin dependencias externas)
- **Frontend:** HTML + CSS + JavaScript vanilla (sin frameworks)
- **Despliegue:** Docker (Dockerfile multi-stage + docker-compose)
- **Idioma:** Español chileno colloquial 😄

---

## ✅ Requisitos previos

Para que `./scripts/start.sh` funcione tal cual (recién bajado desde GitHub, sin configurar nada):

- **Docker** instalado y en ejecución. En Linux: Docker Engine. En Windows/macOS: Docker Desktop (en Windows usa **WSL2**, el script es bash).
- **Docker Compose v2** (plugin `docker compose`) o la versión v1 (`docker-compose`). El script detecta cuál tienes.
- **Internet** en el **primer** `start`: se descargan las imágenes base (golang y alpine) y se compila el backend; puede tardar unos minutos. Los siguientes arranques son casi inmediatos.

> El script se encarga solo de lo demás: crea `data/`, `logs/` y `config/` si no existen, y levanta el contenedor con el usuario y la zona horaria de tu máquina (las fechas de los archivos y logs se muestran en tu hora local).

## 🚀 Inicio rápido (Docker)

```bash
./scripts/start.sh start
```

Luego abre <http://localhost:8080>.

Para detener:

```bash
./scripts/start.sh stop
```

### Comandos disponibles

| Comando | Descripción |
|---|---|
| `start` | Compila y levanta Nimbo con Docker |
| `stop` | Detiene el contenedor |
| `restart` | Reinicia (recompila) |
| `status` | Muestra el estado del servicio |
| `logs` | Sigue los logs en vivo |
| `backup` | Crea un backup de data/, logs/ y config/ |
| `help` | Muestra la ayuda |

> **Sin Docker (modo local):** si tienes Go instalado, puedes ejecutar el backend directo con `./scripts/start.sh start --local`. El puerto se cambia con `PORT=9090 ./scripts/start.sh start --local`.

---

## 📁 Estructura del proyecto

```
nimbo/
├── docker/
│   └── Dockerfile              # Multi-stage build
├── backend/
│   ├── cmd/
│   │   └── main.go             # Punto de entrada
│   ├── internal/
│   │   ├── handlers/           # Controladores HTTP
│   │   │   ├── files.go        # Upload, download, delete, list, zip
│   │   │   ├── config.go       # Configuración GET/POST
│   │   │   ├── logs.go         # Visualización de logs
│   │   │   ├── health.go       # Health check
│   │   │   └── respond.go      # Helpers JSON
│   │   ├── services/
│   │   │   ├── storage.go      # Lógica de archivos
│   │   │   └── logger.go       # Sistema de logs con rotación
│   │   ├── models/
│   │   │   └── models.go       # Estructuras de datos
│   │   └── config/
│   │       └── config.go       # Cargar/guardar configuración
│   ├── go.mod                  # module nimbo
│   └── go.sum
├── frontend/
│   ├── index.html              # Página principal
│   ├── css/
│   │   └── style.css           # Estilos dark-mode suave
│   └── js/
│       ├── app.js              # Inicialización y estado global
│       ├── api.js              # Cliente API + utilidades
│       ├── files.js            # Lista y acciones de archivos
│       ├── upload.js           # Subida con barra de progreso
│       ├── notifications.js    # Sistema de toasts
│       ├── config.js           # Panel de configuración
│       └── logs.js             # Visualización de logs
├── scripts/
│   ├── start.sh                # start / stop / restart / status / logs / backup
│   └── backup.sh               # Backup manual
├── config/
│   └── config.json             # Configuración persistente
├── data/                       # Archivos subidos (volumen)
├── logs/                       # Logs de actividad (volumen)
├── backups/                    # Backups generados
├── docker-compose.yml
└── README.md
```

---

## 🧰 Configuración

La configuración vive en `config/config.json`:

```json
{
  "storagePath": "/data",
  "maxFileSize": 1073741824,
  "port": "8080"
}
```

Se puede editar a mano (y se persiste desde el panel ⚙️ Config) o mediante variables de entorno, que tienen prioridad:

| Variable | Default | Descripción |
|---|---|---|
| `STORAGE_PATH` | `/data` | Ruta de almacenamiento |
| `CONFIG_PATH` | `/config/config.json` | Ruta del archivo de configuración |
| `LOG_PATH` | `/logs/activity.log` | Ruta del archivo de logs |
| `PORT` | `8080` | Puerto del servidor |
| `MAX_FILE_SIZE` | `1073741824` (1 GB) | Tamaño máximo por archivo (bytes) |
| `FRONTEND_PATH` | `/app/frontend` | Carpeta del frontend estático |

> El cambio de puerto se aplica al **reiniciar** el servidor. El cambio de ruta de almacenamiento y de tamaño máximo se aplica **al instante**.

---

## 🔌 API REST

| Endpoint | Método | Descripción |
|---|---|---|
| `/` | GET | Frontend (HTML) |
| `/api/files` | GET | Listar archivos → `{files: [{name, size, modTime}]}` |
| `/api/upload` | POST | Subir archivos (multipart/form-data) → `{success, files, errors}` |
| `/api/download/{filename}` | GET | Descargar archivo individual |
| `/api/download-zip` | POST | Descargar ZIP → body `{files: ["a.txt"]}` |
| `/api/delete/{filename}` | DELETE | Eliminar archivo → `{success}` |
| `/api/config` | GET | Obtener configuración |
| `/api/config` | POST | Actualizar configuración |
| `/api/logs` | GET | Últimas 100 líneas de log (parámetro `limit`) |
| `/api/health` | GET | Health check → `{status: "ok"}` |

---

## 📊 Sistema de logs

Los logs se escriben en **JSON Lines** (una línea = un JSON):

```json
{"timestamp":"2026-09-01T12:30:00-03:00","level":"INFO","event":"FILE_UPLOADED","message":"foto.jpg (5242880 bytes)"}
```

### Eventos registrados

| Evento | Nivel |
|---|---|
| `FILE_UPLOADED` | INFO |
| `FILE_DELETED` | INFO |
| `FILE_DOWNLOADED` | INFO |
| `FILE_DOWNLOAD_ZIP` | INFO |
| `CONFIG_CHANGED` | INFO |
| `ERROR_UPLOAD` | ERROR |
| `SERVER_STARTED` | INFO |
| `SERVER_STOPPED` | INFO |

### Rotación

- **Diaria** (nuevo archivo por día)
- O cuando el archivo activo **supera los 10 MB**
- Se conservan los **últimos 7 días**

---

## 🔒 Seguridad

- **Sanitización de nombres:** `filepath.Base()` + reemplazo de caracteres peligrosos (`../`, `\`, etc.) para evitar path traversal.
- **Límite de tamaño:** 1 GB por archivo, aplicado durante el streaming con `io.LimitReader`.
- **Path traversal:** cada ruta se valida contra la raíz con `filepath.Clean()` antes de abrir/escribir.
- **Manejo de errores:** los mensajes al cliente no exponen rutas del sistema.
- **Usuario sin privilegios** dentro del contenedor (`nimbo`).
- **XSS en frontend:** todo texto dinámico se escapa antes de insertarse en el DOM.

---

## 🧱 Desarrollo local

Requisitos: **Go 1.21+**

```bash
# Compilar
cd backend && go build ./... && go vet ./...

# Ejecutar sin Docker (desde la raíz del proyecto)
./scripts/start.sh start --local
```

---

## 💾 Backups

```bash
./scripts/backup.sh
```

Crea `backups/nimbo-backup-<fecha>.tar.gz` con `data/`, `logs/` y `config/`, y conserva los últimos 7 backups.

---

## 🎨 Notas de diseño

- Modo oscuro suave con la paleta `#1a1a2e · #16213e · #1e2a4a · #e8e8e8`
- Acentos azul `#4a9eff` y verde `#4ade80`
- Fuente `system-ui, sans-serif`, responsive (mobile-first)
- Iconos con emojis (sin dependencias externas)
- Toasts con 4 estados (success / error / warning / info), auto-cierre en 5 s

---

## 🌐 Conectarse desde otro dispositivo de la red

Nimbo escucha en todas las interfaces (`0.0.0.0`), así que desde otro equipo de la **misma red** solo abre en el navegador:

```
http://IP_DE_ESTA_MAQUINA:8080
```

Para saber la IP de esta máquina: `hostname -I` (suele ser algo como `192.168.x.x`).

> Si no carga desde el otro dispositivo, revisa el firewall de esta máquina (abre el puerto 8080) y que el router no tenga aislamiento de clientes (AP isolation).

---

## 🔧 Solución de problemas

- **Las fechas/horas se ven en UTC o corridas:** se usa la hora local del host automáticamente. Si igual ves UTC, revisa que la variable `TZ` se detectó bien: `echo $TZ`. Si ejecutas `docker compose up` a mano (sin el script), Nimbo arranca en UTC.
- **El puerto 8080 ya está en uso:** detén lo que lo ocupa o cambia el puerto. En Docker se edita `docker-compose.yml` (mapeo `"8080:8080"`); en modo local: `PORT=9090 ./scripts/start.sh start --local`.
- **`config/config.json` aparece modificado en git:** es normal — al guardar cambios desde el panel ⚙️ el servidor persiste ahí (es la configuración). Puedes hacer `git checkout config/config.json` para descartar cambios.
- **Nimbo no inicia tras `start`:** revisa los logs con `./scripts/start.sh logs` o `docker compose logs nimbo`.

---

## 📜 Licencia

Ver archivo [LICENSE](LICENSE).
