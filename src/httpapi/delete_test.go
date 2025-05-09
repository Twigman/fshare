package httpapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
	"github.com/twigman/fshare/src/store"
	"github.com/twigman/fshare/src/testutil/fake"
)

func setupTest(uploadDir string, apiKey string) (*httpapi.RESTService, string, error) {
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
	}

	as, rs, restService, err := initTestServices(cfg)
	if err != nil {
		return nil, "", err
	}

	key, err := as.AddAPIKey(apiKey, "test key")
	if err != nil {
		return nil, "", err
	}

	_, err = rs.GetOrCreateHomeDir(key.HashedKey)
	if err != nil {
		return nil, "", err
	}

	content := []byte("Hello World")
	file := &fake.FakeMultipartFile{Reader: bytes.NewReader(content)}
	const filename = "test.txt"

	r := &store.Resource{
		Name:              filename,
		IsPrivate:         false,
		APIKeyUUID:        key.UUID,
		AutoDeleteInHours: 0,
	}

	fileUUID, err := rs.SaveUploadedFile(file, r)
	if err != nil {
		return nil, "", err
	}

	return restService, fileUUID, nil
}

func TestDeleteHandler_WithoutAuth(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	restService, fileUUID, err := setupTest(uploadDir, apiKey)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/delete/"+fileUUID, nil)
	rr := httptest.NewRecorder()

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestDeleteHandler_Success(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	restService, fileUUID, err := setupTest(uploadDir, apiKey)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/delete/"+fileUUID, nil)
	rr := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+apiKey)

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestDeleteHandler_WrongKey(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	restService, fileUUID, err := setupTest(uploadDir, apiKey)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/delete/"+fileUUID, nil)
	rr := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer 321")

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestDeleteHandler_NoExistingUUID(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	restService, fileUUID, err := setupTest(uploadDir, apiKey)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/delete/"+fileUUID+"a", nil)
	rr := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+apiKey)

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeleteHandler_WrongMethod(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	restService, fileUUID, err := setupTest(uploadDir, apiKey)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("GET", "/delete/"+fileUUID, nil)
	rr := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+apiKey)

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
