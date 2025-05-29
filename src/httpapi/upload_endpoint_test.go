package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
)

func TestUploadHandler_Success(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
		EnvPath:         filepath.Join(uploadDir, "test.env"),
	}

	as, rs, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	const apiKey = "123"
	key, err := as.AddAPIKey(apiKey, "test key", false)
	if err != nil {
		t.Fatalf("Can not add API key: %v", err)
	}

	// create home dir
	if err := os.MkdirAll(filepath.Join(uploadDir, key.UUID), 0o700); err != nil {
		t.Fatalf("Error creating home dir: %v", err)
	}

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
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	savedPath := filepath.Join(uploadDir, key.UUID, testFilename)
	content, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("File not saved: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("File content mismatch.\nGot:  %q\nWant: %q", content, testContent)
	}

	// parse response JSON in map
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	// extract UUID
	uuid, ok := result["uuid"].(string)
	if !ok {
		t.Fatalf("UUID not found or not a string")
	}

	// database check
	r, err := rs.GetResourceByUUID(uuid)
	if err != nil {
		t.Errorf("Resource does not exist: %v", err)
	}

	if r.UUID != uuid {
		t.Errorf("Wrong resource: %v", err)
	}

	if r.DeletedAt != nil {
		t.Errorf("resource is marked as deleted")
	}
}

func TestUploadHandler_WrongMethod(t *testing.T) {
	cfg := &config.Config{UploadPath: "./doesnotmatter/"}

	_, _, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

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

	as, _, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	const apiKey = "123"
	_, err = as.AddAPIKey(apiKey, "test key", false)
	if err != nil {
		t.Fatalf("Can not add API key: %v", err)
	}

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
	req.Header.Set("Authorization", "Bearer "+apiKey)

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
	as, _, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	const apiKey = "123"
	_, err = as.AddAPIKey(apiKey, "test key", false)
	if err != nil {
		t.Fatalf("Can not add API key: %v", err)
	}

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
	req.Header.Set("Authorization", "Bearer "+apiKey)

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
	cfg := &config.Config{
		UploadPath: t.TempDir(),
	}

	_, _, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()

	restService.UploadHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestUploadHandler_UnregisteredAPIKeyHeader(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
		EnvPath:         filepath.Join(uploadDir, "test.env"),
	}

	as, _, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	const apiKey = "123"
	_, err = as.AddAPIKey(apiKey, "test key", false)
	if err != nil {
		t.Fatalf("Can not add API key: %v", err)
	}

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
	req.Header.Set("Authorization", "Bearer 321")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestUploadHandler_HiddenFile(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
		EnvPath:         filepath.Join(uploadDir, "test.env"),
	}

	as, rs, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	const apiKey = "123"
	key, err := as.AddAPIKey(apiKey, "test key", false)
	if err != nil {
		t.Fatalf("Can not add API key: %v", err)
	}

	rs.GetOrCreateHomeDir(key.HashedKey)

	ts := httptest.NewServer(http.HandlerFunc(restService.UploadHandler))
	defer ts.Close()

	const testFilename = ".hidden"
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
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestUploadHandler_InvalidFilename_1(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
		EnvPath:         filepath.Join(uploadDir, "test.env"),
	}

	as, rs, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Can not initialize test services: %v", err)
	}

	const apiKey = "123"
	key, err := as.AddAPIKey(apiKey, "test key", false)
	if err != nil {
		t.Fatalf("Can not add API key: %v", err)
	}

	rs.GetOrCreateHomeDir(key.HashedKey)

	ts := httptest.NewServer(http.HandlerFunc(restService.UploadHandler))
	defer ts.Close()

	const testFilename = "../hall\\o.txt"
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
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
