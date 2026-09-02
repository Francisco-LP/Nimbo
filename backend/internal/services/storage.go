package services

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"nimbo/internal/models"
)

// ErrFileTooLarge se retorna cuando un archivo supera el tamaño máximo.
var ErrFileTooLarge = errors.New("el archivo supera el tamaño máximo permitido")

// Storage encapsula las operaciones sobre el directorio de almacenamiento.
// El mutex protege la raíz, que puede cambiar en caliente desde la configuración.
type Storage struct {
	mu   sync.RWMutex
	root string
}

// NewStorage crea el almacenamiento y asegura que el directorio raíz exista.
func NewStorage(root string) (*Storage, error) {
	s := &Storage{root: root}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("no se pudo crear el directorio de almacenamiento: %w", err)
	}
	return s, nil
}

// Root devuelve la ruta raíz actual.
func (s *Storage) Root() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.root
}

// SetRoot cambia la ruta raíz en caliente (al actualizar la configuración).
func (s *Storage) SetRoot(root string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	s.root = root
	return nil
}

// SanitizeName limpia un nombre de archivo para evitar path traversal y
// caracteres peligrosos. filepath.Base descarta cualquier ruta previa.
func SanitizeName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, `\`, "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "..", "_")
	name = strings.Trim(name, " .")
	if name == "" {
		name = "archivo"
	}
	return name
}

// Resolve asegura que un nombre de archivo resuelva dentro de la raíz,
// devolviendo la ruta absoluta. Es la barrera principal anti path traversal.
func (s *Storage) Resolve(name string) (string, error) {
	root := filepath.Clean(s.Root())
	clean := filepath.Clean(filepath.Join(root, filepath.Base(name)))
	if clean != root && !strings.HasPrefix(clean, root+string(os.PathSeparator)) {
		return "", errors.New("ruta de archivo inválida")
	}
	return clean, nil
}

// List devuelve los archivos del directorio raíz ordenados por nombre.
func (s *Storage) List() ([]models.FileInfo, error) {
	entries, err := os.ReadDir(s.Root())
	if err != nil {
		return nil, err
	}
	files := make([]models.FileInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		// Ocultar archivos ocultos (p. ej. .gitkeep) del listado.
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, models.FileInfo{
			Name:    e.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})
	return files, nil
}

// Save escribe el contenido de reader en el almacenamiento de forma eficiente
// (streaming, sin cargar el archivo completo en memoria). No permite
// sobreescribir archivos existentes.
func (s *Storage) Save(name string, reader io.Reader, maxSize int64) (int64, error) {
	dest, err := s.Resolve(name)
	if err != nil {
		return 0, err
	}
	if _, err := os.Stat(dest); err == nil {
		return 0, errors.New("ya existe un archivo con ese nombre")
	}

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return 0, err
	}

	// io.LimitReader frena la escritura al superar el tamaño máximo,
	// evitando llenar el disco con archivos demasiado grandes.
	written, err := io.Copy(f, io.LimitReader(reader, maxSize+1))
	if err != nil {
		f.Close()
		_ = os.Remove(dest)
		return written, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(dest)
		return written, err
	}
	if written > maxSize {
		_ = os.Remove(dest)
		return written, ErrFileTooLarge
	}
	return written, nil
}

// Open abre un archivo para lectura (descarga). La ruta se valida contra
// path traversal.
func (s *Storage) Open(name string) (*os.File, error) {
	path, err := s.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

// Stat devuelve la información de un archivo.
func (s *Storage) Stat(name string) (os.FileInfo, error) {
	path, err := s.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(path)
}

// Delete elimina un archivo del almacenamiento.
func (s *Storage) Delete(name string) error {
	path, err := s.Resolve(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// UsedSize devuelve el espacio total ocupado por los archivos.
func (s *Storage) UsedSize() (int64, error) {
	entries, err := os.ReadDir(s.Root())
	if err != nil {
		return 0, err
	}
	var total int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if info, err := e.Info(); err == nil {
			total += info.Size()
		}
	}
	return total, nil
}

// CreateZip escribe un ZIP con los archivos solicitados en el writer indicado.
// Cada archivo se agrega de forma secuencial sin cargar todo en memoria.
// Devuelve la cantidad de archivos agregados al ZIP.
func (s *Storage) CreateZip(names []string, w io.Writer) (int, error) {
	zw := zip.NewWriter(w)
	added := 0

	for _, name := range names {
		path, err := s.Resolve(name)
		if err != nil {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		info, err := f.Stat()
		if err != nil {
			f.Close()
			continue
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			f.Close()
			continue
		}
		header.Name = filepath.Base(name)
		header.Method = zip.Deflate

		entry, err := zw.CreateHeader(header)
		if err != nil {
			f.Close()
			continue
		}
		if _, err := io.Copy(entry, f); err != nil {
			f.Close()
			continue
		}
		f.Close()
		added++
	}

	if err := zw.Close(); err != nil {
		return added, err
	}
	if added == 0 {
		return 0, errors.New("ninguno de los archivos seleccionados existe")
	}
	return added, nil
}
