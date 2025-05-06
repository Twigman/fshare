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
	"github.com/twigman/fshare/src/store"
)

func initServices(cfg *config.Config) (*store.APIKeyService, *store.FileService, error) {
	db, err := store.NewDB(cfg.SQLitePath)
	if err != nil {
		return nil, nil, err
	}

	fileService := store.NewFileService(cfg, db)
	apiKeyService := store.NewAPIKeyService(db)

	_, err = apiKeyService.AddAPIKey("123", "123")
	if err != nil {
		return nil, nil, err
	}

	return apiKeyService, fileService, nil
}

func TestUploadHandler_Success(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:               uploadDir,
		MaxFileSizeInMB:          5,
		Port:                     8080,
		SQLitePath:               filepath.Join(uploadDir, "test_db.sqlite"),
		ContinuousFileValidation: false,
		MaxFolderDepth:           3,
	}

	apiKeyService, fileService, err := initServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	restService := NewRESTService(cfg, apiKeyService, fileService)

	ts := httptest.NewServer(http.HandlerFunc(restService.UploadHandler))
	defer ts.Close()

	const testFilename = "test.txt"
	const testContent = "Hello World\nTest!123%"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", testFilename)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}

	if _, err := io.Copy(part, strings.NewReader(testContent)); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	writer.WriteField("auto_del_in_h", "0")
	writer.WriteField("is_private", "true")

	if err := writer.Close(); err != nil {
		t.Fatalf("Writer close failed: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer 123")

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

	apiKeyService, fileService, err := initServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	restService := NewRESTService(cfg, apiKeyService, fileService)

	req := httptest.NewRequest("GET", "/upload", nil)
	w := httptest.NewRecorder()

	restService.UploadHandler(w, req)

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

	apiKeyService, fileService, err := initServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	restService := NewRESTService(cfg, apiKeyService, fileService)
	ts := httptest.NewServer(http.HandlerFunc(restService.UploadHandler))
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
	req.Header.Set("Authorization", "Bearer 123")

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
	apiKeyService, fileService, err := initServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	restService := NewRESTService(cfg, apiKeyService, fileService)

	ts := httptest.NewServer(http.HandlerFunc(restService.UploadHandler))
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
	req.Header.Set("Authorization", "Bearer 123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestUploadHandler_MissingAuthHeader(t *testing.T) {
	cfg := &config.Config{UploadPath: t.TempDir()}

	apiKeyService, fileService, err := initServices(cfg)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	restService := NewRESTService(cfg, apiKeyService, fileService)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()

	restService.UploadHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}
