// Package services contiene la lógica de negocio del servidor:
// almacenamiento de archivos y sistema de logs.
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Niveles de log soportados.
const (
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// Eventos de actividad registrados por el sistema.
const (
	EventFileUploaded    = "FILE_UPLOADED"
	EventFileDeleted     = "FILE_DELETED"
	EventFileDownloaded  = "FILE_DOWNLOADED"
	EventFileDownloadZip = "FILE_DOWNLOAD_ZIP"
	EventConfigChanged   = "CONFIG_CHANGED"
	EventErrorUpload     = "ERROR_UPLOAD"
	EventServerStarted   = "SERVER_STARTED"
	EventServerStopped   = "SERVER_STOPPED"
)

// Límites de rotación de logs.
const (
	// maxLogSizeBytes es el tamaño máximo del archivo activo antes de rotar (10 MB).
	maxLogSizeBytes = 10 * 1024 * 1024
	// maxLogDays es la antigüedad máxima de los logs rotados que se conservan.
	maxLogDays = 7
)

// LogEntry es un registro de actividad en formato JSON (JSON Lines).
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Event     string `json:"event"`
	Message   string `json:"message"`
}

// Logger escribe logs en formato JSON Lines con rotación diaria y por tamaño.
type Logger struct {
	mu   sync.Mutex
	path string // ruta del archivo de logs activo
	day  string // día (YYYY-MM-DD) al que corresponde el archivo activo
}

// NewLogger crea un logger apuntando a la ruta indicada, asegurando que el
// directorio de logs exista.
func NewLogger(path string) (*Logger, error) {
	l := &Logger{path: path}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("no se pudo crear el directorio de logs: %w", err)
	}
	l.day = time.Now().Format("2006-01-02")
	return l, nil
}

// Log escribe una entrada en el archivo de logs. Es seguro llamarlo desde
// múltiples goroutines gracias al mutex.
func (l *Logger) Log(level, event, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Event:     event,
		Message:   message,
	}

	l.rotateIfNeeded()

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(append(line, '\n'))
}

// rotateIfNeeded rota el archivo activo si cambió el día o si superó el
// tamaño máximo. Los archivos rotados se nombran con un sufijo de fecha/hora.
func (l *Logger) rotateIfNeeded() {
	today := time.Now().Format("2006-01-02")
	size := fileSize(l.path)

	if today == l.day && size < maxLogSizeBytes {
		return
	}

	// Renombrar el archivo activo solo si tiene contenido.
	if size > 0 {
		stamp := time.Now().Format("2006-01-02_150405")
		_ = os.Rename(l.path, fmt.Sprintf("%s.%s", l.path, stamp))
	}
	l.day = today
	l.cleanupOld()
}

// cleanupOld elimina archivos rotados con más de maxLogDays de antigüedad.
func (l *Logger) cleanupOld() {
	matches, err := filepath.Glob(l.path + ".*")
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -maxLogDays)
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(m)
		}
	}
}

// ReadLastLines devuelve las últimas n líneas de log, combinando el archivo
// activo con los rotados más recientes en orden cronológico.
func (l *Logger) ReadLastLines(n int) ([]LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Archivos de más reciente a más antiguo: primero el activo.
	var files []string
	if fileSize(l.path) > 0 {
		files = append(files, l.path)
	}
	rotated, _ := filepath.Glob(l.path + ".*")
	sort.Slice(rotated, func(i, j int) bool {
		return modTime(rotated[i]).After(modTime(rotated[j]))
	})
	files = append(files, rotated...)

	var entries []LogEntry
	for _, f := range files {
		lines, err := readLogFile(f)
		if err != nil {
			continue
		}
		// Anteponer el archivo más antiguo para mantener el orden cronológico.
		entries = append(lines, entries...)
		if len(entries) >= n {
			break
		}
	}
	if len(entries) > n {
		entries = entries[len(entries)-n:]
	}
	return entries, nil
}

// fileSize devuelve el tamaño en bytes de un archivo (0 si no existe).
func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// modTime devuelve la fecha de modificación de un archivo.
func modTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// readLogFile lee y decodifica todas las líneas JSON de un archivo de log.
func readLogFile(path string) ([]LogEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []LogEntry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e LogEntry
		if json.Unmarshal([]byte(line), &e) == nil {
			entries = append(entries, e)
		}
	}
	return entries, nil
}
