package sender

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSendUploadsFileWithChecksum(t *testing.T) {
	content := []byte("arquivo enviado pelo GoDrop")
	path := filepath.Join(t.TempDir(), "teste.txt")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	sum := sha256.Sum256(content)
	wantChecksum := hex.EncodeToString(sum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/upload" {
			t.Errorf("path = %q, want /upload", r.URL.Path)
		}
		if got := r.Header.Get("X-GoDrop-Filename"); got != "teste.txt" {
			t.Errorf("filename = %q, want teste.txt", got)
		}
		if got := r.Header.Get("X-GoDrop-SHA256"); got != wantChecksum {
			t.Errorf("checksum = %q, want %q", got, wantChecksum)
		}
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ReadAll() error = %v", err)
		}
		if !bytes.Equal(got, content) {
			t.Errorf("body = %q, want %q", got, content)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	var progress bytes.Buffer

	result, err := Send(context.Background(), server.Client(), server.URL, path, &progress)

	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.Bytes != int64(len(content)) || result.Checksum != wantChecksum {
		t.Fatalf("Send() result = %+v", result)
	}
	if !strings.Contains(progress.String(), "100%") {
		t.Fatalf("progress = %q, want 100%%", progress.String())
	}
	if count := strings.Count(progress.String(), "100%"); count != 1 {
		t.Fatalf("progress contains 100%% %d times, want 1; output = %q", count, progress.String())
	}
}

func TestSendReturnsReceiverError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "teste.txt")
	if err := os.WriteFile(path, []byte("dados"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "arquivo já existe", http.StatusConflict)
	}))
	defer server.Close()

	_, err := Send(context.Background(), server.Client(), server.URL, path, io.Discard)

	if err == nil || !strings.Contains(err.Error(), "409") {
		t.Fatalf("Send() error = %v, want status 409", err)
	}
}

func TestNormalizeTarget(t *testing.T) {
	got, err := normalizeTarget("192.168.1.20:8080")
	if err != nil {
		t.Fatalf("normalizeTarget() error = %v", err)
	}
	if want := "http://192.168.1.20:8080/upload"; got != want {
		t.Fatalf("normalizeTarget() = %q, want %q", got, want)
	}
}
