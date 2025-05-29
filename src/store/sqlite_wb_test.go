package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestFindFilesForDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	as := NewAPIKeyService(db)

	// prepare testdata
	// add apikey
	key, err := as.AddAPIKey("123", "test", false)
	if err != nil {
		t.Fatalf("error creating apikey: %v", err)
	}

	past := time.Now().Add(-1 * time.Hour).UTC()
	future := time.Now().Add(1 * time.Hour).UTC()

	res1 := &Resource{
		UUID:         "res1",
		Name:         "oldfile.txt",
		IsFile:       true,
		APIKeyUUID:   key.UUID,
		AutoDeleteAt: &past,
		CreatedAt:    past, // delete
	}

	res2 := &Resource{
		UUID:         "res2",
		Name:         "newfile.txt",
		IsFile:       true,
		APIKeyUUID:   key.UUID,
		AutoDeleteAt: &future,
		CreatedAt:    past,
	}

	res3 := &Resource{
		UUID:         "res3",
		Name:         "deletedfile.txt",
		IsFile:       true,
		APIKeyUUID:   key.UUID,
		AutoDeleteAt: &past,
		CreatedAt:    past,
		DeletedAt:    &past, // already deleted
	}

	res4 := &Resource{
		UUID:         "res4",
		Name:         "newfile2.txt",
		IsFile:       true,
		APIKeyUUID:   key.UUID,
		AutoDeleteAt: &past,
		CreatedAt:    past, // delete
	}

	for _, r := range []*Resource{res1, res2, res3, res4} {
		if err := db.insertResource(r); err != nil {
			t.Fatalf("failed to insert resource: %v", err)
		}
	}

	// Test: should return res1 and res4
	results, err := db.findFilesForDeletion(time.Now().UTC())
	if err != nil {
		t.Fatalf("findFilesForDeletion error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 resources for deletion, got %d", len(results))
	}

	if results[0].UUID != "res1" && results[1].UUID != "res4" {
		if results[0].UUID != "res4" && results[1].UUID != "res1" {
			t.Errorf("expected res1 and res4, got %s and %s", results[0].UUID, results[1].UUID)
		}
	}
}
