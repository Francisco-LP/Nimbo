package handlers

import (
	"net/http"
	"strconv"

	"nimbo/internal/models"
	"nimbo/internal/services"
)

// LogsHandler maneja GET /api/logs.
type LogsHandler struct {
	logger *services.Logger
}

// NewLogsHandler crea un LogsHandler con sus dependencias.
func NewLogsHandler(logger *services.Logger) *LogsHandler {
	return &LogsHandler{logger: logger}
}

// List maneja GET /api/logs?limit=100. Devuelve las últimas líneas de log.
func (h *LogsHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.SimpleResponse{Success: false, Message: "método no permitido"})
		return
	}
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}
	entries, err := h.logger.ReadLastLines(limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.SimpleResponse{Success: false, Message: "no se pudieron leer los logs"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"logs": entries})
}
