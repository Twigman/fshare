package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/twigman/fshare/src/testutils/fake"
)

func initServices(testPath string, dbName string) (*FileService, error) {
	db, err := NewDB(filepath.Join(testPath, dbName))
	if err != nil {
		return nil, err
	}

	fileService := NewFileService(testPath, db)
	apiKeyService := NewAPIKeyService(db)

	_, err = apiKeyService.AddAPIKey("123", "123")
	if err != nil {
		return nil, err
	}

	return fileService, nil
}

func TestFileService_SaveUploadedFile(t *testing.T) {
	uploadDir := t.TempDir()
	const testFilename = "test.txt"

	// Testfile
	content := []byte("Hello World")
	file := &fake.FakeMultipartFile{Reader: bytes.NewReader(content)}

	res := &Resource{
		Name:              testFilename,
		IsPrivate:         true,
		OwnerHashedKey:    HashAPIKey("123"),
		AutoDeleteInHours: 0,
	}

	fs, err := initServices(uploadDir, "temp_db.sqlite")
	if err != nil {
		t.Fatalf("Error initializing test services: %v", err)
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
