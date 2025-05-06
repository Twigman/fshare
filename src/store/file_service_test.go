package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/testutil/fake"
)

func initServices(cfg *config.Config) (*FileService, *APIKey, error) {
	db, err := NewDB(cfg.SQLitePath)
	if err != nil {
		return nil, nil, err
	}

	fileService := NewFileService(cfg, db)
	apiKeyService := NewAPIKeyService(db)

	key, err := apiKeyService.AddAPIKey("123", "123")
	if err != nil {
		return nil, nil, err
	}

	return fileService, key, nil
}

func TestFileService_SaveUploadedFile(t *testing.T) {
	uploadDir := t.TempDir()
	const testFilename = "test.txt"

	cfg := &config.Config{
		UploadPath:               uploadDir,
		MaxFileSizeInMB:          2,
		Port:                     8080,
		SQLitePath:               filepath.Join(uploadDir, "test_db.sqlite"),
		ContinuousFileValidation: false,
		MaxFolderDepth:           3,
	}

	// Testfile
	content := []byte("Hello World")
	file := &fake.FakeMultipartFile{Reader: bytes.NewReader(content)}

	fs, key, err := initServices(cfg)
	if err != nil {
		t.Fatalf("Error initializing test services: %v", err)
	}

	res := &Resource{
		Name:              testFilename,
		IsPrivate:         true,
		APIKeyUUID:        key.UUID,
		AutoDeleteInHours: 0,
	}

	// tested function
	_, err = fs.SaveUploadedFile(file, res)
	if err != nil {
		t.Fatalf("Error saving file: %v", err)
	}

	savedPath := filepath.Join(uploadDir, testFilename)
	data, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("File was not saved: %v", err)
	}

	if !bytes.Equal(data, content) {
		t.Errorf("Wrong file content.\nGot:  %q\nWant: %q", data, content)
	}
}

func TestFileService_GetOrCreateHomeDir(t *testing.T) {
	uploadDir := t.TempDir()

	// Setup: Config & DB
	cfg := &config.Config{
		UploadPath: uploadDir,
	}
	db, err := NewDB(filepath.Join(uploadDir, "test.sqlite"))
	if err != nil {
		t.Fatalf("DB init error: %v", err)
	}

	fs := NewFileService(cfg, db)

	// prepare api key
	apiKey := "test-api-key"
	hashed := HashAPIKey(apiKey)

	apiKeyUUID := "00000000-0000-0000-0000-000000000001"
	err = db.insertAPIKey(&APIKey{
		UUID:      apiKeyUUID,
		HashedKey: hashed,
		Comment:   "Test Key",
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("insertAPIKey failed: %v", err)
	}

	// create home dir
	resource, err := fs.GetOrCreateHomeDir(hashed)
	if err != nil {
		t.Fatalf("GetOrCreateHomeDir error: %v", err)
	}

	// check resource
	if resource.Name != apiKeyUUID {
		t.Errorf("unexpected resource name: got %s, want %s", resource.Name, apiKeyUUID)
	}
	if !resource.IsPrivate || resource.IsFile || resource.ParentUUID != nil {
		t.Errorf("resource flags invalid: %+v", resource)
	}

	// check folder
	expectedPath := filepath.Join(uploadDir, apiKeyUUID)
	if stat, err := os.Stat(expectedPath); err != nil {
		t.Errorf("home directory not created: %v", err)
	} else if !stat.IsDir() {
		t.Errorf("expected directory, got file: %s", expectedPath)
	}

	// check idempotence
	r2, err := fs.GetOrCreateHomeDir(hashed)
	if err != nil {
		t.Fatalf("second GetOrCreateHomeDir call failed: %v", err)
	}
	if r2.UUID != resource.UUID {
		t.Errorf("expected same resource UUID, got %s vs %s", r2.UUID, resource.UUID)
	}
}
