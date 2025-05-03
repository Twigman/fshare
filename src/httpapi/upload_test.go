package httpapi

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twigman/fshare/src/config"
)

func TestUploadHandler_Success(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
	}
	handler := NewHTTPHandler(cfg)

	ts := httptest.NewServer(http.HandlerFunc(handler.UploadHandler))
	defer ts.Close()

	testFilename := "test.txt"
	testContent := "Hello World\nTest!123%"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", testFilename)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}

	if _, err := io.Copy(part, strings.NewReader(testContent)); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Writer close failed: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	savedPath := filepath.Join(uploadDir, testFilename)
	content, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("File not saved: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("File content mismatch.\nGot:  %q\nWant: %q", content, testContent)
	}
}

func TestUploadHandler_WrongMethod(t *testing.T) {
	cfg := &config.Config{UploadPath: "./doesnotmatter/"}
	handler := NewHTTPHandler(cfg)

	req := httptest.NewRequest("GET", "/upload", nil)
	w := httptest.NewRecorder()

	handler.UploadHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Recieved status 405, instead %d", resp.StatusCode)
	}
}

func TestUploadHandler_TooLarge(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 1,
		Port:            8080,
	}
	handler := NewHTTPHandler(cfg)
	ts := httptest.NewServer(http.HandlerFunc(handler.UploadHandler))
	defer ts.Close()

	// create 2 MiB content
	var bigContent bytes.Buffer
	for i := 0; i < 2<<20; i++ {
		bigContent.WriteByte('A')
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "big.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(part, &bigContent); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status %d, got %d", http.StatusRequestEntityTooLarge, resp.StatusCode)
	}
}

func TestUploadHandler_MissingFileField(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
	}
	handler := NewHTTPHandler(cfg)
	ts := httptest.NewServer(http.HandlerFunc(handler.UploadHandler))
	defer ts.Close()

	// create multipart without "file"-Field
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("something_else", "value")
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
