package sender

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Result describes a completed transfer.
type Result struct {
	Bytes    int64
	Checksum string
}

// Send streams a file to a GoDrop receiver.
func Send(ctx context.Context, client *http.Client, target, path string, progress io.Writer) (Result, error) {
	file, err := os.Open(path)
	if err != nil {
		return Result{}, fmt.Errorf("abrir arquivo: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return Result{}, fmt.Errorf("obter informações do arquivo: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Result{}, fmt.Errorf("o caminho não aponta para um arquivo comum")
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return Result{}, fmt.Errorf("calcular checksum: %w", err)
	}
	checksum := hex.EncodeToString(hash.Sum(nil))
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return Result{}, fmt.Errorf("preparar arquivo: %w", err)
	}

	uploadURL, err := normalizeTarget(target)
	if err != nil {
		return Result{}, err
	}
	body := io.Reader(file)
	if progress != nil {
		body = &progressReader{reader: file, total: info.Size(), writer: progress}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return Result{}, fmt.Errorf("criar requisição: %w", err)
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-GoDrop-Filename", filepath.Base(path))
	req.Header.Set("X-GoDrop-SHA256", checksum)

	response, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("conectar ao receptor: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return Result{}, fmt.Errorf("receptor respondeu %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	return Result{Bytes: info.Size(), Checksum: checksum}, nil
}

func normalizeTarget(target string) (string, error) {
	if !strings.Contains(target, "://") {
		target = "http://" + target
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("endereço do receptor inválido")
	}
	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/upload"
	}
	return parsed.String(), nil
}

type progressReader struct {
	reader io.Reader
	total  int64
	read   int64
	last   int
	writer io.Writer
}

func (p *progressReader) Read(buffer []byte) (int, error) {
	n, err := p.reader.Read(buffer)
	p.read += int64(n)
	percent := 100
	if p.total > 0 {
		percent = int(p.read * 100 / p.total)
	}
	if percent != p.last {
		fmt.Fprintf(p.writer, "\rEnviando: %3d%%", percent)
		p.last = percent
	}
	if err == io.EOF {
		fmt.Fprintln(p.writer)
	}
	return n, err
}
