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

func initServices(cfg *config.Config) (*ResourceService, *APIKey, error) {
	db, err := NewDB(cfg.SQLitePath)
	if err != nil {
		return nil, nil, err
	}

	rs := NewResourceService(cfg, db)
	as := NewAPIKeyService(db)

	key, err := as.AddAPIKey("123", "123", false)
	if err != nil {
		return nil, nil, err
	}

	return rs, key, nil
}

func TestFileService_SaveUploadedFile(t *testing.T) {
	uploadDir := t.TempDir()
	const testFilename = "test.txt"

	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 2,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
	}

	// Testfile
	content := []byte("Hello World")
	file := &fake.FakeMultipartFile{Reader: bytes.NewReader(content)}

	rs, key, err := initServices(cfg)
	if err != nil {
		t.Fatalf("Error initializing test services: %v", err)
	}

	res := &Resource{
		Name:         testFilename,
		IsPrivate:    true,
		APIKeyUUID:   key.UUID,
		AutoDeleteAt: nil,
	}

	// create dir
	err = os.Mkdir(filepath.Join(uploadDir, key.UUID), 0o700)
	if err != nil {
		t.Fatalf("Error creating home dir: %v", err)
	}

	// tested function
	_, err = rs.SaveUploadedFile(file, res, false)
	if err != nil {
		t.Fatalf("Error saving file: %v", err)
	}

	savedPath := filepath.Join(uploadDir, key.UUID, testFilename)
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

	rs := NewResourceService(cfg, db)

	// prepare api key
	apiKey := "test-api-key"
	hashedKey, err := hashAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Error hashing key: %v", err)
	}

	apiKeyUUID := "00000000-0000-0000-0000-000000000001"
	err = db.insertAPIKey(&APIKey{
		UUID:      apiKeyUUID,
		HashedKey: hashedKey,
		Comment:   "Test Key",
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("insertAPIKey failed: %v", err)
	}

	// create home dir
	resource, err := rs.GetOrCreateHomeDir(hashedKey)
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
	r2, err := rs.GetOrCreateHomeDir(hashedKey)
	if err != nil {
		t.Fatalf("second GetOrCreateHomeDir call failed: %v", err)
	}
	if r2.UUID != resource.UUID {
		t.Errorf("expected same resource UUID, got %s vs %s", r2.UUID, resource.UUID)
	}
}

func TestFileService_SaveUploadedFileMultipleTimes(t *testing.T) {
	uploadDir := t.TempDir()
	const testFilename = "test.txt"

	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 2,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
	}

	rs, key, err := initServices(cfg)
	if err != nil {
		t.Fatalf("Error initializing test services: %v", err)
	}

	// create dir
	err = os.Mkdir(filepath.Join(uploadDir, key.UUID), 0o700)
	if err != nil {
		t.Fatalf("Error creating home dir: %v", err)
	}

	res := &Resource{
		Name:         testFilename,
		IsPrivate:    true,
		APIKeyUUID:   key.UUID,
		AutoDeleteAt: nil,
	}

	counter := 15

	for i := 0; i < counter; i++ {
		// testfile
		// needs to be created every iteration
		content := []byte("Hello World")
		file := &fake.FakeMultipartFile{Reader: bytes.NewReader(content)}

		// create file for the first time
		fileUUID, err := rs.SaveUploadedFile(file, res, true)
		if err != nil {
			t.Fatalf("Error saving file: %v", err)
		}

		r, err := rs.GetResourceByUUID(fileUUID)
		if err != nil {
			t.Fatalf("Error loading resource from db: %v", err)
		}

		savedPath := filepath.Join(uploadDir, key.UUID, r.Name)
		data, err := os.ReadFile(savedPath)
		if err != nil {
			t.Fatalf("File was not saved: %v", err)
		}

		if !bytes.Equal(data, content) {
			t.Errorf("Wrong file content.\nGot:  %q\nWant: %q", data, content)
		}
	}
}

func TestCleanupExpiredFiles(t *testing.T) {
	uploadDir := t.TempDir()
	dbPath := filepath.Join(uploadDir, "test_db.sqlite")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}

	cfg := &config.Config{UploadPath: uploadDir}
	rs := NewResourceService(cfg, db)
	as := NewAPIKeyService(db)

	apiKey := "test-key"
	key, err := as.AddAPIKey(apiKey, "test", false)
	if err != nil {
		t.Fatalf("failed to add api key: %v", err)
	}

	now := time.Now().UTC()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name          string
		resource      *Resource
		expectDeleted bool
	}{
		{
			name: "expired file",
			resource: &Resource{
				UUID:         "res-123",
				Name:         "old.txt",
				IsFile:       true,
				APIKeyUUID:   key.UUID,
				AutoDeleteAt: &past,
				CreatedAt:    now,
			},
			expectDeleted: true,
		},
		{
			name: "non-expired file",
			resource: &Resource{
				UUID:         "res-321",
				Name:         "future.txt",
				IsFile:       true,
				APIKeyUUID:   key.UUID,
				AutoDeleteAt: &future,
				CreatedAt:    now,
			},
			expectDeleted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Insert resource into DB
			if err := db.insertResource(tt.resource); err != nil {
				t.Fatalf("failed to insert resource: %v", err)
			}

			// Create file
			homeDir := filepath.Join(uploadDir, key.UUID)
			_ = os.MkdirAll(homeDir, 0o755)
			testFile := filepath.Join(homeDir, tt.resource.Name)
			os.WriteFile(testFile, []byte("test"), 0o644)

			// Run cleanup
			err := rs.cleanupExpiredFiles()
			if err != nil {
				t.Fatalf("cleanup error: %v", err)
			}

			// Check file existence
			_, err = os.Stat(testFile)
			if tt.expectDeleted {
				if !os.IsNotExist(err) {
					t.Errorf("expected file to be deleted, but it still exists")
				}
			} else {
				if err != nil {
					t.Errorf("expected file to exist, but got error: %v", err)
				}
			}

			// Check DB state
			updated, err := db.findResourceByUUID(tt.resource.UUID)
			if err != nil {
				t.Fatalf("db lookup error: %v", err)
			}
			if tt.expectDeleted {
				if updated.DeletedAt == nil {
					t.Errorf("expected resource to be marked as deleted")
				}
			} else {
				if updated.DeletedAt != nil {
					t.Errorf("expected resource not to be marked as deleted")
				}
			}
		})
	}
}
