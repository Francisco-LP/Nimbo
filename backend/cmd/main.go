// Nimbo - Nube Casera
// Punto de entrada del servidor. Configura la API REST, sirve el frontend
// estático y gestiona el apagado ordenado (graceful shutdown).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"nimbo/internal/config"
	"nimbo/internal/handlers"
	"nimbo/internal/services"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Nimbo: %v", err)
	}
}

func run() error {
	// 1. Configuración (archivo + variables de entorno).
	cfg, err := config.New()
	if err != nil {
		return err
	}

	// 2. Sistema de logs.
	logPath := env("LOG_PATH", "/logs/activity.log")
	logger, err := services.NewLogger(logPath)
	if err != nil {
		return err
	}
	logger.Log(services.LevelInfo, services.EventServerStarted, fmt.Sprintf("servidor iniciado (puerto %s)", cfg.Get().Port))

	// 3. Almacenamiento.
	store, err := services.NewStorage(cfg.Get().StoragePath)
	if err != nil {
		return err
	}

	// 4. Handlers de la API.
	fileHandler := handlers.NewFileHandler(store, logger, cfg)
	configHandler := handlers.NewConfigHandler(cfg, store, logger)
	logsHandler := handlers.NewLogsHandler(logger)
	healthHandler := handlers.NewHealthHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/files", fileHandler.List)
	mux.HandleFunc("/api/upload", fileHandler.Upload)
	mux.HandleFunc("/api/download/", fileHandler.Download)
	mux.HandleFunc("/api/download-zip", fileHandler.DownloadZip)
	mux.HandleFunc("/api/delete/", fileHandler.Delete)
	mux.HandleFunc("/api/config", configHandler.Handle)
	mux.HandleFunc("/api/logs", logsHandler.List)
	mux.HandleFunc("/api/health", healthHandler.Check)

	// 5. Frontend estático (rutas / no prefijadas con /api).
	frontendDir := resolveFrontendPath()
	if _, err := os.Stat(filepath.Join(frontendDir, "index.html")); err != nil {
		logger.Log(services.LevelWarn, services.EventConfigChanged, fmt.Sprintf("no se encontró index.html en %q; el frontend responderá 404", frontendDir))
	}
	serveFrontend(mux, frontendDir)

	server := &http.Server{
		Addr:              ":" + cfg.Get().Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 6. Esperar señales de apagado (Ctrl+C / SIGTERM) y apagar con calma.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Nimbo escuchando en http://0.0.0.0:%s", cfg.Get().Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	logger.Log(services.LevelInfo, services.EventServerStopped, "servidor detenido")
	return nil
}

// serveFrontend entrega los archivos estáticos del frontend, usando
// index.html como página de inicio. Las rutas /api/* desconocidas
// devuelven un 404 JSON en lugar de caer en el catch-all del frontend.
func serveFrontend(mux *http.ServeMux, dir string) {
	fs := http.FileServer(http.Dir(dir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"no encontrado"}`))
			return
		}
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}

// resolveFrontendPath determina dónde están los archivos estáticos.
// En Docker se usa /app/frontend; si no existe, busca la carpeta frontend
// relativa al directorio de trabajo (ejecución local directa).
func resolveFrontendPath() string {
	if v := os.Getenv("FRONTEND_PATH"); v != "" {
		return v
	}
	if _, err := os.Stat("/app/frontend"); err == nil {
		return "/app/frontend"
	}
	for _, cand := range []string{"frontend", "./frontend", "../frontend"} {
		if _, err := os.Stat(cand); err == nil {
			if abs, err := filepath.Abs(cand); err == nil {
				return abs
			}
			return cand
		}
	}
	return "frontend"
}

// env devuelve el valor de una variable de entorno o un valor por defecto.
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
