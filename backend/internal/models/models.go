// Package models define las estructuras de datos compartidas del servidor.
package models

// FileInfo representa un archivo dentro del almacenamiento.
type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

// UploadError describe un error al subir un archivo concreto.
type UploadError struct {
	Name string `json:"name"`
	Msg  string `json:"msg"`
}

// UploadResponse es la respuesta del endpoint de subida.
type UploadResponse struct {
	Success bool          `json:"success"`
	Files   []string      `json:"files,omitempty"`
	Errors  []UploadError `json:"errors,omitempty"`
}

// SimpleResponse es una respuesta genérica {success, message}.
type SimpleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
