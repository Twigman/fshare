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
		EnvPath:         filepath.Join(uploadDir, "test.env"),
	}
	_, _, restService, err := httpapi.InitTestServices(cfg)
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
	const keyHighlyTrusted = false
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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
	const keyHighlyTrusted = false
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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

func TestResourceHandler_PrivateAuthorizedNotTrusted(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "validkey"
	const filename = "private.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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

func TestResourceHandler_PublicTextFileNotTrusted(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "apipub"
	const filename = "example.go"
	const isPrivate = false
	const keyHighlyTrusted = false
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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

func TestResourceHandler_PublicPNGNotTruested(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "imgkey"
	const filename = "image.png"
	const isPrivate = false
	const keyHighlyTrusted = false
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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

func TestResourceHandler_PublicBinaryDownloadNotTrusted(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "binkey"
	const filename = "data.zip"
	const isPrivate = false
	const keyHighlyTrusted = false
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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

func TestResourceHandler_PublicFileMissing(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "misskey"
	const filename = "ghost.txt"
	const isPrivate = false
	const keyHighlyTrusted = false
	restService, rs, key, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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

func TestResourceHandler_PublicPDFNotTrusted(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "key"
	const filename = "test.pdf"
	const isPrivate = false
	const keyHighlyTrusted = false
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/octet-stream") {
		t.Errorf("Expected octet-stream content type, got %s", ct)
	}
	if disp := w.Header().Get("Content-Disposition"); !strings.Contains(disp, "attachment") {
		t.Errorf("Expected attachment disposition, got %s", disp)
	}
}

func TestResourceHandler_PublicPDFTrusted(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "key"
	const filename = "test.pdf"
	const isPrivate = false
	const keyHighlyTrusted = true
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	// not application/pdf because the pdf is displayed in an iframe
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Expected html content type, got %s", ct)
	}
}

func TestResourceHandler_PublicSVGTrusted(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "key"
	const filename = "test.svg"
	const isPrivate = false
	const keyHighlyTrusted = true
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/svg") {
		t.Errorf("Expected svg content type, got %s", ct)
	}
}

func TestResourceHandler_PublicSVGNotTrusted(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "key"
	const filename = "test.svg"
	const isPrivate = false
	const keyHighlyTrusted = false
	restService, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.ResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/octet-stream") {
		t.Errorf("Expected octet-stream content type, got %s", ct)
	}
	if disp := w.Header().Get("Content-Disposition"); !strings.Contains(disp, "attachment") {
		t.Errorf("Expected attachment disposition, got %s", disp)
	}
}
