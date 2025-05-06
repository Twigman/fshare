package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

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
