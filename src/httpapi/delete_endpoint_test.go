package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
	"github.com/twigman/fshare/src/store"
)

func TestDeleteHandler_WithoutAuth(t *testing.T) {
	dataDir := t.TempDir()
	const apiKey = "123"
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, _, _, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(dataDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", config.EndpointDelete+fileUUID, nil)
	rr := httptest.NewRecorder()

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestDeleteHandler_Success(t *testing.T) {
	dataDir := t.TempDir()
	const apiKey = "123"
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, rs, _, key, cfg, fileUUID, err := httpapi.SetupExistingTestUpload(dataDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", config.EndpointDelete+fileUUID, nil)
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
	_, err = os.Stat(filepath.Join(cfg.UploadPath, key.HashedKey, filename))
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
	restService, _, _, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(uploadDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", config.EndpointDelete+fileUUID, nil)
	rr := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer 321")

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestDeleteHandler_NoExistingUUID(t *testing.T) {
	dataDir := t.TempDir()
	const apiKey = "123"
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, _, _, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(dataDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", config.EndpointDelete+fileUUID+"a", nil)
	rr := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+apiKey)

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestDeleteHandler_WrongMethod(t *testing.T) {
	dataDir := t.TempDir()
	const apiKey = "123"
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, _, _, _, _, fileUUID, err := httpapi.SetupExistingTestUpload(dataDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("GET", config.EndpointDelete+fileUUID, nil)
	rr := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+apiKey)

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// should not be possible
func TestDeleteHandler_DeleteHomeDir(t *testing.T) {
	dataDir := t.TempDir()
	const apiKey = "123"
	const keyHighlyTrusted = false

	cfg := &config.Config{
		DataPath:        dataDir,
		UploadPath:      filepath.Join(dataDir, "upload"),
		MaxFileSizeInMB: 5,
		Port:            8080,
	}

	as, rs, restService, err := httpapi.InitTestServices(cfg)
	if err != nil {
		t.Fatalf("Test setup error: %v", err)
	}

	err = store.CreateDirsFromConfig(cfg)
	if err != nil {
		t.Fatalf("Error creating app dirs: %v", err)
	}

	// add key
	key, err := as.AddAPIKey(apiKey, "test key", keyHighlyTrusted, nil)
	if err != nil {
		t.Fatalf("Error adding API key: %v", err)
	}

	// create home
	r, err := rs.GetOrCreateHomeDir(key.HashedKey)
	if err != nil {
		t.Fatalf("Error creating home dir: %v", err)
	}

	// check home
	_, err = os.Stat(filepath.Join(cfg.UploadPath, key.UUID))
	if os.IsNotExist(err) {
		t.Fatalf("home directory was not created")
	}

	req := httptest.NewRequest("DELETE", config.EndpointDelete+r.UUID, nil)
	rr := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+apiKey)

	restService.DeleteHandler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// check home
	_, err = os.Stat(filepath.Join(cfg.UploadPath, key.UUID))
	if os.IsNotExist(err) {
		t.Fatalf("home directory was deleted")
	}
}

func TestDeleteHandler_TwoTimes(t *testing.T) {
	dataDir := t.TempDir()
	const apiKey = "123"
	const filename = "test.txt"
	const isPrivate = true
	const keyHighlyTrusted = false
	restService, rs, _, key, cfg, fileUUID, err := httpapi.SetupExistingTestUpload(dataDir, apiKey, filename, isPrivate, keyHighlyTrusted)
	if err != nil {
		t.Errorf("Test setup error: %v", err)
	}

	req := httptest.NewRequest("DELETE", config.EndpointDelete+fileUUID, nil)
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
	_, err = os.Stat(filepath.Join(cfg.UploadPath, key.HashedKey, filename))
	if os.IsExist(err) {
		t.Errorf("Testfile still exists: %v", err)
	}

	// delete second time
	rr2 := httptest.NewRecorder()
	restService.DeleteHandler(rr2, req)

	if rr2.Code != http.StatusNotFound {
		t.Errorf("expected %d, got %d", http.StatusNotFound, rr.Code)
	}
}
