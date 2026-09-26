package receiver

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Handler receives files through HTTP without overwriting existing files.
type Handler struct {
	dir      string
	maxBytes int64
}

// NewHandler creates an HTTP handler that stores uploads inside dir.
func NewHandler(dir string, maxBytes int64) http.Handler {
	return &Handler{dir: dir, maxBytes: maxBytes}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.URL.Path != "/upload" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	filename, err := safeFilename(r.Header.Get("X-GoDrop-Filename"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		http.Error(w, "não foi possível preparar a pasta de destino", http.StatusInternalServerError)
		return
	}

	temporary, err := os.CreateTemp(h.dir, ".godrop-*.part")
	if err != nil {
		http.Error(w, "não foi possível criar o arquivo temporário", http.StatusInternalServerError)
		return
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)

	hash := sha256.New()
	body := http.MaxBytesReader(w, r.Body, h.maxBytes)
	_, copyErr := io.Copy(io.MultiWriter(temporary, hash), body)
	closeErr := temporary.Close()
	if copyErr != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(copyErr, &maxBytesError) {
			http.Error(w, "arquivo excede o limite permitido", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "falha ao receber o arquivo", http.StatusBadRequest)
		return
	}
	if closeErr != nil {
		http.Error(w, "falha ao salvar o arquivo", http.StatusInternalServerError)
		return
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	if expected := r.Header.Get("X-GoDrop-SHA256"); expected != "" && !strings.EqualFold(expected, checksum) {
		http.Error(w, "checksum SHA-256 não confere", http.StatusUnprocessableEntity)
		return
	}

	destination := filepath.Join(h.dir, filename)
	if err := os.Link(temporaryName, destination); err != nil {
		if os.IsExist(err) {
			http.Error(w, "já existe um arquivo com esse nome", http.StatusConflict)
			return
		}
		http.Error(w, "não foi possível concluir o arquivo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-GoDrop-SHA256", checksum)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s salvo com sucesso\n", filename)
}

func safeFilename(name string) (string, error) {
	if name == "" {
		return "", errors.New("informe o nome em X-GoDrop-Filename")
	}
	if filepath.Base(name) != name || name == "." || name == ".." {
		return "", errors.New("nome de arquivo inválido")
	}
	return name, nil
}
