package handlers

import "net/http"

// HealthHandler maneja GET /api/health.
type HealthHandler struct{}

// NewHealthHandler crea un HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check maneja GET /api/health. Útil para healthchecks de Docker.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"status": "error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
