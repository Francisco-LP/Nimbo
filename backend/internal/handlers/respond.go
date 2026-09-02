// Package handlers contiene los controladores HTTP de la API REST.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// writeJSON escribe una respuesta JSON con el código de estado indicado.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error escribiendo respuesta JSON: %v", err)
	}
}

// readJSON decodifica el cuerpo de la petición en v, rechazando campos
// desconocidos para evitar payloads mal formados.
func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
