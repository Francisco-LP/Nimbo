package handlers

import (
	"fmt"
	"net/http"

	"nimbo/internal/config"
	"nimbo/internal/models"
	"nimbo/internal/services"
)

// ConfigHandler maneja GET/POST /api/config.
type ConfigHandler struct {
	cfg    *config.Manager
	store  *services.Storage
	logger *services.Logger
}

// NewConfigHandler crea un ConfigHandler con sus dependencias.
func NewConfigHandler(cfg *config.Manager, store *services.Storage, logger *services.Logger) *ConfigHandler {
	return &ConfigHandler{cfg: cfg, store: store, logger: logger}
}

// Handle despacha la petición según su método HTTP.
func (h *ConfigHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	case http.MethodPost:
		h.Update(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, models.SimpleResponse{Success: false, Message: "método no permitido"})
	}
}

// Get maneja GET /api/config.
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.cfg.Get())
}

// Update maneja POST /api/config.
func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req config.Config
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.SimpleResponse{Success: false, Message: "solicitud inválida"})
		return
	}

	old := h.cfg.Get()
	if err := h.cfg.Update(req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.SimpleResponse{Success: false, Message: err.Error()})
		return
	}

	// Aplicar en caliente los cambios que sí pueden aplicar sin reiniciar.
	if req.StoragePath != "" && req.StoragePath != old.StoragePath {
		if err := h.store.SetRoot(req.StoragePath); err != nil {
			h.logger.Log(services.LevelError, services.EventConfigChanged, fmt.Sprintf("no se pudo cambiar la ruta de almacenamiento: %v", err))
		} else {
			h.logger.Log(services.LevelInfo, services.EventConfigChanged, fmt.Sprintf("ruta de almacenamiento cambiada a %q", req.StoragePath))
		}
	}

	msg := "configuración guardada"
	if req.Port != "" && req.Port != old.Port {
		msg += ". El cambio de puerto se aplica al reiniciar el servidor"
	}
	h.logger.Log(services.LevelInfo, services.EventConfigChanged, "configuración actualizada")
	writeJSON(w, http.StatusOK, models.SimpleResponse{Success: true, Message: msg})
}
