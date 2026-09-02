package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"nimbo/internal/config"
	"nimbo/internal/models"
	"nimbo/internal/services"
)

// FileHandler maneja las rutas de operaciones con archivos.
type FileHandler struct {
	store  *services.Storage
	logger *services.Logger
	cfg    *config.Manager
}

// NewFileHandler crea un FileHandler con sus dependencias.
func NewFileHandler(store *services.Storage, logger *services.Logger, cfg *config.Manager) *FileHandler {
	return &FileHandler{store: store, logger: logger, cfg: cfg}
}

// List maneja GET /api/files.
func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.SimpleResponse{Success: false, Message: "método no permitido"})
		return
	}
	files, err := h.store.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.SimpleResponse{Success: false, Message: "no se pudo leer el directorio de almacenamiento"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"files": files})
}

// Upload maneja POST /api/upload (multipart/form-data).
// Usa MultipartReader para procesar los archivos de forma secuencial y en
// streaming, sin cargar el cuerpo completo en memoria (soporta archivos
// grandes de hasta el límite configurado).
func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.SimpleResponse{Success: false, Message: "método no permitido"})
		return
	}
	reader, err := r.MultipartReader()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.UploadResponse{Success: false})
		return
	}

	maxSize := h.cfg.Get().MaxFileSize
	resp := models.UploadResponse{Success: true}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		// Se ignoran los campos de formulario que no sean archivos.
		if part.FileName() == "" {
			continue
		}

		name := services.SanitizeName(part.FileName())

		// El límite de tamaño se aplica durante el streaming en Save
		// (io.LimitReader + detección de exceso).
		written, err := h.store.Save(name, part, maxSize)
		if err != nil {
			h.uploadError(&resp, name, friendlyErr(err))
			continue
		}

		resp.Files = append(resp.Files, name)
		h.logger.Log(services.LevelInfo, services.EventFileUploaded, fmt.Sprintf("%s (%d bytes)", name, written))
	}

	status := http.StatusOK
	if len(resp.Errors) > 0 && len(resp.Files) == 0 {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, resp)
}

// uploadError agrega un error a la respuesta y lo registra en los logs.
func (h *FileHandler) uploadError(resp *models.UploadResponse, name, msg string) {
	resp.Success = false
	resp.Errors = append(resp.Errors, models.UploadError{Name: name, Msg: msg})
	h.logger.Log(services.LevelError, services.EventErrorUpload, fmt.Sprintf("%s: %s", name, msg))
}

// Download maneja GET /api/download/{filename}.
func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.SimpleResponse{Success: false, Message: "método no permitido"})
		return
	}
	name := services.SanitizeName(strings.TrimPrefix(r.URL.Path, "/api/download/"))
	if name == "" {
		writeJSON(w, http.StatusBadRequest, models.SimpleResponse{Success: false, Message: "nombre de archivo inválido"})
		return
	}

	file, err := h.store.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, models.SimpleResponse{Success: false, Message: "archivo no encontrado"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, models.SimpleResponse{Success: false, Message: "no se pudo abrir el archivo"})
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	// ServeContent maneja rangos, modificación y HEAD de forma correcta.
	http.ServeContent(w, r, name, time.Time{}, file)

	h.logger.Log(services.LevelInfo, services.EventFileDownloaded, name)
}

// DownloadZip maneja POST /api/download-zip.
func (h *FileHandler) DownloadZip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.SimpleResponse{Success: false, Message: "método no permitido"})
		return
	}
	var req struct {
		Files []string `json:"files"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.SimpleResponse{Success: false, Message: "solicitud inválida"})
		return
	}
	if len(req.Files) == 0 {
		writeJSON(w, http.StatusBadRequest, models.SimpleResponse{Success: false, Message: "no se seleccionaron archivos"})
		return
	}

	// Verificar existencia antes de escribir el ZIP para no enviar
	// una respuesta corrupta a medias.
	for _, f := range req.Files {
		name := services.SanitizeName(f)
		if _, err := h.store.Stat(name); err != nil {
			writeJSON(w, http.StatusNotFound, models.SimpleResponse{Success: false, Message: "alguno de los archivos seleccionados no existe"})
			return
		}
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", "nimbo-"+time.Now().Format("20060102-150405")+".zip"))

	added, err := h.store.CreateZip(req.Files, w)
	if err != nil {
		h.logger.Log(services.LevelError, services.EventErrorUpload, fmt.Sprintf("error al crear ZIP: %v", err))
		return
	}
	h.logger.Log(services.LevelInfo, services.EventFileDownloadZip, fmt.Sprintf("ZIP con %d archivo(s)", added))
}

// Delete maneja DELETE /api/delete/{filename}.
func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, models.SimpleResponse{Success: false, Message: "método no permitido"})
		return
	}
	name := services.SanitizeName(strings.TrimPrefix(r.URL.Path, "/api/delete/"))
	if name == "" {
		writeJSON(w, http.StatusBadRequest, models.SimpleResponse{Success: false, Message: "nombre de archivo inválido"})
		return
	}
	if err := h.store.Delete(name); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.SimpleResponse{Success: false, Message: "no se pudo eliminar el archivo"})
		return
	}
	h.logger.Log(services.LevelInfo, services.EventFileDeleted, name)
	writeJSON(w, http.StatusOK, models.SimpleResponse{Success: true, Message: "archivo eliminado"})
}

// friendlyErr traduce errores internos a mensajes seguros, sin exponer
// rutas del sistema en la respuesta al cliente.
func friendlyErr(err error) string {
	switch {
	case errors.Is(err, services.ErrFileTooLarge):
		return "el archivo supera el tamaño máximo permitido"
	case strings.Contains(err.Error(), "ya existe"):
		return "ya existe un archivo con ese nombre"
	default:
		return "error al guardar el archivo"
	}
}
