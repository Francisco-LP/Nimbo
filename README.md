# ☁️ Nimbo — Nube Casera

Nimbo es una aplicación de **nube personal** para red local. Permite subir, descargar, eliminar y gestionar archivos desde el navegador web.

- **Backend:** Go 
- **Frontend:** HTML + CSS + JavaScript vanilla 
- **Despliegue:** Docker 

---

## ✅ Requisitos previos

Para que `./scripts/start.sh` funcione tal cual (recién bajado desde GitHub, sin configurar nada) se debe tener:

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

## 🧱 Desarrollo local

Requisitos: **Go 1.21+**

```bash
# Compilar
cd backend && go build ./... && go vet ./...

# Ejecutar sin Docker (desde la raíz del proyecto)
./scripts/start.sh start --local
```

---

## 🌐 Conectarse desde otro dispositivo de la red

Nimbo escucha en todas las interfaces (`0.0.0.0`), así que desde otro equipo de la **misma red** solo abre en el navegador:

```
http://IP_DE_LA_MAQUINA_QUE_CORRE_NIMBO:8080
```

Para saber la IP de esta máquina: `hostname -I` (suele ser algo como `192.168.x.x`).

> Si no carga desde el otro dispositivo, revisa el firewall de esta máquina (abre el puerto 8080) y que el router no tenga aislamiento de clientes (AP isolation).

---

## 📜 Licencia

Ver archivo [LICENSE](LICENSE).
