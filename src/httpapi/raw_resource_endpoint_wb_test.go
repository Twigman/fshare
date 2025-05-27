package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/twigman/fshare/src/config"
)

func TestRawResourceHandler_MethodNotAllowed(t *testing.T) {
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath: uploadDir,
		SQLitePath: filepath.Join(uploadDir, "test_db.sqlite"),
	}
	_, _, restService, err := InitTestServices(cfg)
	if err != nil {
		t.Fatalf("InitTestServices error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/raw/someuuid", nil)
	w := httptest.NewRecorder()

	restService.RawResourceHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestRawResourceHandler_Unauthorized(t *testing.T) {
	uploadDir := t.TempDir()
	restService, _, _, _, fileUUID, err := SetupExistingTestUpload(uploadDir, "apikey", "test.txt", false, false)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// UUID without signature
	req := httptest.NewRequest(http.MethodGet, "/raw/"+fileUUID, nil)
	w := httptest.NewRecorder()

	restService.RawResourceHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRawResourceHandler_FileMissing(t *testing.T) {
	uploadDir := t.TempDir()
	filename := "missing.txt"
	restService, rs, _, key, fileUUID, err := SetupExistingTestUpload(uploadDir, "apikey", filename, false, false)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// delete file
	filePath := filepath.Join(uploadDir, key.UUID, filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Remove error: %v", err)
	}

	// valid request
	expiry := time.Now().Add(30 * time.Second)

	signURL, err := restService.generateSignedURL("/raw", fileUUID, expiry)
	if err != nil {
		t.Fatalf("Signing error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, signURL, nil)
	w := httptest.NewRecorder()

	restService.RawResourceHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected %d, got %d", http.StatusNotFound, w.Code)
	}

	// check whether resource is marked as broken
	r, err := rs.GetResourceByUUID(fileUUID)
	if err != nil {
		t.Fatalf("GetResourceByUUID error: %v", err)
	}
	if !r.IsBroken {
		t.Errorf("Expected resource to be marked as broken")
	}
}

func TestRawResourceHandler_ValidSigningWithWrongUUID(t *testing.T) {
	uploadDir := t.TempDir()
	restService, rs, _, _, fileUUID, err := SetupExistingTestUpload(uploadDir, "apikey", "test.txt", false, false)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// valid request with non-existent uuid
	expiry := time.Now().Add(5 * time.Second)

	signURL, err := restService.generateSignedURL("/raw", "uuid", expiry)
	if err != nil {
		t.Fatalf("Signing error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, signURL, nil)
	w := httptest.NewRecorder()

	restService.RawResourceHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected %d, got %d", http.StatusNotFound, w.Code)
	}

	// file shouldnt be marked as broken
	r, err := rs.GetResourceByUUID(fileUUID)
	if err != nil {
		t.Fatalf("GetResourceByUUID error: %v", err)
	}
	if r.IsBroken {
		t.Errorf("Expected resource not to be marked as broken")
	}
}

func TestRawResourceHandler_Success(t *testing.T) {
	uploadDir := t.TempDir()
	restService, _, _, _, fileUUID, err := SetupExistingTestUpload(uploadDir, "apikey", "hello.txt", false, false)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// valid request
	expiry := time.Now().Add(5 * time.Second)

	signURL, err := restService.generateSignedURL("/raw", fileUUID, expiry)
	if err != nil {
		t.Fatalf("Signing error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, signURL, nil)
	w := httptest.NewRecorder()

	restService.RawResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/plain") {
		t.Errorf("Expected text/plain Content-Type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestRawResourceHandler_DownloadHeader(t *testing.T) {
	uploadDir := t.TempDir()
	restService, _, _, _, fileUUID, err := SetupExistingTestUpload(uploadDir, "apikey", "archive.zip", false, false)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// valid request
	expiry := time.Now().Add(5 * time.Second)

	signURL, err := restService.generateSignedURL("/raw", fileUUID, expiry)
	if err != nil {
		t.Fatalf("Signing error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, signURL+"&download=true", nil)
	w := httptest.NewRecorder()

	restService.RawResourceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
	if disp := w.Header().Get("Content-Disposition"); !strings.Contains(disp, "attachment") {
		t.Errorf("Expected attachment in Content-Disposition, got %s", disp)
	}
}
