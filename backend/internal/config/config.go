// Package config maneja la carga, validación y persistencia de la configuración.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
)

// Config es la configuración persistente del servidor.
type Config struct {
	StoragePath string `json:"storagePath"`
	MaxFileSize int64  `json:"maxFileSize"`
	Port        string `json:"port"`
}

// Valores por defecto usados cuando no existen variables de entorno ni archivo.
const (
	defaultStoragePath = "/data"
	defaultMaxFileSize = int64(1073741824) // 1 GB
	defaultPort        = "8080"
)

// Manager permite acceder y persistir la configuración de forma segura
// (protegida con un mutex para lecturas/escrituras concurrentes).
type Manager struct {
	mu         sync.RWMutex
	config     Config
	configPath string
}

// New crea un Manager cargando la configuración desde el archivo JSON y las
// variables de entorno. Prioridad: variables de entorno > archivo > defaults.
func New() (*Manager, error) {
	m := &Manager{
		configPath: env("CONFIG_PATH", "/config/config.json"),
		config: Config{
			StoragePath: env("STORAGE_PATH", defaultStoragePath),
			MaxFileSize: envInt64("MAX_FILE_SIZE", defaultMaxFileSize),
			Port:        env("PORT", defaultPort),
		},
	}

	// Cargar desde el archivo JSON si existe (no es fatal si falta).
	if data, err := os.ReadFile(m.configPath); err == nil {
		var fromFile Config
		if err := json.Unmarshal(data, &fromFile); err == nil {
			if fromFile.StoragePath != "" {
				m.config.StoragePath = fromFile.StoragePath
			}
			if fromFile.MaxFileSize > 0 {
				m.config.MaxFileSize = fromFile.MaxFileSize
			}
			if fromFile.Port != "" {
				m.config.Port = fromFile.Port
			}
		}
	}

	// Las variables de entorno siempre tienen la última palabra.
	if v := os.Getenv("STORAGE_PATH"); v != "" {
		m.config.StoragePath = v
	}
	if v := os.Getenv("PORT"); v != "" {
		m.config.Port = v
	}
	if v := os.Getenv("MAX_FILE_SIZE"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			m.config.MaxFileSize = n
		}
	}

	return m, nil
}

// Get devuelve una copia de la configuración actual.
func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Update valida, persiste en disco y actualiza la configuración en memoria.
// Persistir primero evita quedar con config en memoria distinta a la del disco.
func (m *Manager) Update(c Config) error {
	if c.StoragePath == "" {
		return fmt.Errorf("la ruta de almacenamiento no puede estar vacía")
	}
	if c.MaxFileSize <= 0 {
		return fmt.Errorf("el tamaño máximo debe ser mayor que cero")
	}
	if c.Port == "" {
		return fmt.Errorf("el puerto no puede estar vacío")
	}

	next := Config{
		StoragePath: c.StoragePath,
		MaxFileSize: c.MaxFileSize,
		Port:        c.Port,
	}

	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(m.configPath, data, 0o644); err != nil {
		return fmt.Errorf("no se pudo guardar la configuración: %w", err)
	}

	m.mu.Lock()
	m.config = next
	m.mu.Unlock()
	return nil
}

// env devuelve el valor de una variable de entorno o un valor por defecto.
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// envInt64 devuelve una variable de entorno como entero o un valor por defecto.
func envInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return def
}
