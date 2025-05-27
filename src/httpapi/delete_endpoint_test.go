package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/twigman/fshare/src/httpapi"
)

func TestDeleteHandler_WithoutAuth(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, _, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, rs, _, key, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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

	// database check
	r, err := rs.GetResourceByUUID(fileUUID)
	if err != nil {
		t.Errorf("Resource does not exist: %v", err)
	}

	if r.UUID != fileUUID {
		t.Errorf("Wrong resource: %v", err)
	}

	if r.DeletedAt == nil {
		t.Errorf("resource not marked as deleted: %v", err)
	}
	// check file
	_, err = os.Stat(filepath.Join(uploadDir, key.HashedKey, filename))
	if os.IsExist(err) {
		t.Errorf("Testfile still exists: %v", err)
	}
}

func TestDeleteHandler_WrongKey(t *testing.T) {
	uploadDir := t.TempDir()
	const apiKey = "123"
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, _, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, _, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, _, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
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
