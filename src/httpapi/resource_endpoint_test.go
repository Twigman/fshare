package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
)

func TestResourceHandler_MethodNotAllowed(t *testing.T) {
	s := &httpapi.RESTService{}
	req := httptest.NewRequest(http.MethodPost, "/r/someuuid", nil)
	w := httptest.NewRecorder()

	s.ResourceHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestResourceHandler_NotFound(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
	}
	_, _, restService, err := initTestServices(cfg)
	if err != nil {
		t.Errorf("Error initialising test services: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/invaliduuid", nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestResourceHandler_PrivateUnauthorized(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	const filename = "secret.txt"
	const isPrivate = true
	restService, _, _, fileUUID, err := setupExistingTestUpload(uploadDir, apiKey, filename, isPrivate)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestResourceHandler_PrivateWrongKey(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	const filename = "secret.txt"
	const isPrivate = true
	restService, _, _, fileUUID, err := setupExistingTestUpload(uploadDir, apiKey, filename, isPrivate)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer 321")

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestResourceHandler_PrivateAuthorized(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "validkey"
	const filename = "private.txt"
	const isPrivate = true

	restService, _, _, fileUUID, err := setupExistingTestUpload(uploadDir, apiKey, filename, isPrivate)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestResourceHandler_PublicTextFile(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "apipub"
	const filename = "example.go"
	const isPrivate = false

	restService, _, _, fileUUID, err := setupExistingTestUpload(uploadDir, apiKey, filename, isPrivate)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	if !strings.Contains(w.Body.String(), "<html") {
		t.Errorf("Expected HTML response for text file")
	}
}

func TestResourceHandler_PublicImage(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "imgkey"
	const filename = "image.png"
	const isPrivate = false

	restService, _, _, fileUUID, err := setupExistingTestUpload(uploadDir, apiKey, filename, isPrivate)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/") {
		t.Errorf("Expected image content type, got %s", ct)
	}
}

func TestResourceHandler_PublicBinaryDownload(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "binkey"
	const filename = "data.zip"
	const isPrivate = false

	restService, _, _, fileUUID, err := setupExistingTestUpload(uploadDir, apiKey, filename, isPrivate)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	if disp := w.Header().Get("Content-Disposition"); !strings.Contains(disp, "attachment") {
		t.Errorf("Expected attachment disposition, got %s", disp)
	}
}

func TestResourceHandler_FileMissing(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "misskey"
	const filename = "ghost.txt"
	const isPrivate = false

	restService, rs, key, fileUUID, err := setupExistingTestUpload(uploadDir, apiKey, filename, isPrivate)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// delete file
	filePath := filepath.Join(uploadDir, key.UUID, filename)
	if err := os.Remove(filePath); err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected %d, got %d", http.StatusInternalServerError, w.Code)
	}

	r, err := rs.GetResourceByUUID(fileUUID)
	if err != nil {
		t.Fatalf("File got lost in database?: %v", err)
	}

	if r.Name != filename {
		t.Fatalf("Wrong filename: Got %q not %q", r.Name, filename)
	}

	if !r.IsBroken {
		t.Fatalf("Missing file was not updated in db")
	}
}
