package receiver

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerReceivesFile(t *testing.T) {
	dir := t.TempDir()
	content := "conteúdo de teste"
	sum := sha256.Sum256([]byte(content))
	checksum := hex.EncodeToString(sum[:])
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(content))
	req.Header.Set("X-GoDrop-Filename", "teste.txt")
	req.Header.Set("X-GoDrop-SHA256", checksum)
	response := httptest.NewRecorder()

	NewHandler(dir, 1024).ServeHTTP(response, req)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %q", response.Code, http.StatusCreated, response.Body.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "teste.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != content {
		t.Fatalf("saved content = %q, want %q", string(got), content)
	}
	if response.Header().Get("X-GoDrop-SHA256") != checksum {
		t.Fatalf("checksum response = %q, want %q", response.Header().Get("X-GoDrop-SHA256"), checksum)
	}
}

func TestHandlerRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("dados"))
	req.Header.Set("X-GoDrop-Filename", "../segredo.txt")
	response := httptest.NewRecorder()

	NewHandler(dir, 1024).ServeHTTP(response, req)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestHandlerRejectsInvalidChecksum(t *testing.T) {
	dir := t.TempDir()
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("dados"))
	req.Header.Set("X-GoDrop-Filename", "teste.txt")
	req.Header.Set("X-GoDrop-SHA256", strings.Repeat("0", 64))
	response := httptest.NewRecorder()

	NewHandler(dir, 1024).ServeHTTP(response, req)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
	if _, err := os.Stat(filepath.Join(dir, "teste.txt")); !os.IsNotExist(err) {
		t.Fatalf("destination should not exist, Stat() error = %v", err)
	}
}

func TestHandlerRejectsFilesOverLimit(t *testing.T) {
	dir := t.TempDir()
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("arquivo grande"))
	req.Header.Set("X-GoDrop-Filename", "grande.txt")
	response := httptest.NewRecorder()

	NewHandler(dir, 4).ServeHTTP(response, req)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestHandlerDoesNotOverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "existente.txt")
	if err := os.WriteFile(destination, []byte("original"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("novo"))
	req.Header.Set("X-GoDrop-Filename", "existente.txt")
	response := httptest.NewRecorder()

	NewHandler(dir, 1024).ServeHTTP(response, req)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "original" {
		t.Fatalf("existing content = %q, want original", string(got))
	}
}
