package store

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIKeyService_AddAPIKey(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		comment         string
		trusted         bool
		expectCreateErr bool
		ErrStr          string
	}{
		{"not trusted key", "123abcDEF", "test", false, false, ""},
		{"trusted key", "test", "test", true, false, ""},
		{"all symbols", "azAZ09-_.", "test", false, false, ""},
		{"samed key", "123abcDEF", "test", false, true, "UNIQUE constraint failed"},
		{"not allowed special chars", "azAZ09-_.=", "test", false, true, "One or more characters are not permitted"},
	}
	uploadDir := t.TempDir()
	dbPath := filepath.Join(uploadDir, "test_db.sqlite")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("could not init test db %v", err)
	}

	as := NewAPIKeyService(db)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := as.AddAPIKey(tt.key, tt.comment, tt.trusted, nil)
			if err != nil {
				if !tt.expectCreateErr {
					t.Fatalf("error adding apikey %v", err)
				}
				if !strings.Contains(err.Error(), tt.ErrStr) {
					t.Fatalf("unexpected error: %v", err)
				} else {
					// expected error; end test
					return
				}
			}

			testKeyHash, err := hashAPIKey(tt.key)
			if err != nil {
				t.Fatalf("Could not hash API key for testing")
			}

			if key.Comment != tt.comment {
				t.Fatalf("wrong comment set for api key")
			}
			if key.IsHighlyTrusted != tt.trusted {
				t.Fatalf("wrong API key trust level")
			}
			if key.CreatedAt.IsZero() {
				t.Fatalf("Expected CreatedAt to be set, got zero value")
			}
			if key.UUID == "" {
				t.Fatalf("UUID is empty")
			}
			if key.HashedKey != testKeyHash {
				t.Fatalf("API key has not been hashed correctly")
			}

			// check api key in db
			dbKey, err := db.findAPIKeyByUUID(key.UUID)
			if err != nil {
				t.Fatalf("API key not found in db")
			}

			if dbKey.HashedKey != testKeyHash {
				t.Fatalf("stored API key hash is incorrect")
			}
			if dbKey.Comment != tt.comment {
				t.Fatalf("stored API key comment is incorrect")
			}
			if dbKey.IsHighlyTrusted != tt.trusted {
				t.Fatalf("stored API key has the an incorrect trust level")
			}
			if !dbKey.CreatedAt.Equal(key.CreatedAt) {
				t.Fatalf("stored API key timestamp is different from created API key timestamp")
			}
			if dbKey.UUID != key.UUID {
				t.Fatalf("stored API key UUID is different from created API key UUID")
			}
		})
	}
}

func TestAPIKeyService_AnyAPIKeyExists(t *testing.T) {
	uploadDir := t.TempDir()
	dbPath := filepath.Join(uploadDir, "test_db.sqlite")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("could not init test db %v", err)
	}

	as := NewAPIKeyService(db)
	if as.AnyAPIKeyExists() {
		t.Fatalf("no API key should exist at this time")
	}

	as.AddAPIKey("123", "test", false, nil)
	if !as.AnyAPIKeyExists() {
		t.Fatalf("API key should exist")
	}

}

func TestAPIKeyService_IsAPIKeyHighlyTrusted(t *testing.T) {
	uploadDir := t.TempDir()
	dbPath := filepath.Join(uploadDir, "test_db.sqlite")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("could not init test db %v", err)
	}

	as := NewAPIKeyService(db)

	// No Key → false
	trusted, err := as.IsAPIKeyHighlyTrusted("non-existent-uuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trusted {
		t.Fatalf("expected key to not be trusted (does not exist)")
	}

	// add trusted key
	k, err := as.AddAPIKey("trustedkey", "test", true, nil)
	if err != nil {
		t.Fatalf("could not add API key: %v", err)
	}

	trusted, err = as.IsAPIKeyHighlyTrusted(k.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !trusted {
		t.Fatalf("expected key to be trusted")
	}

	// add untrusted key
	k2, err := as.AddAPIKey("untrustedkey", "test", false, nil)
	if err != nil {
		t.Fatalf("could not add API key: %v", err)
	}

	trusted, err = as.IsAPIKeyHighlyTrusted(k2.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trusted {
		t.Fatalf("expected key to not be trusted")
	}
}

func TestAPIKeyService_GetUUIDForAPIKey(t *testing.T) {
	uploadDir := t.TempDir()
	dbPath := filepath.Join(uploadDir, "test_db.sqlite")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("could not init test db %v", err)
	}

	as := NewAPIKeyService(db)

	// no key
	uuid, err := as.GetUUIDForAPIKey("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid != "" {
		t.Fatalf("expected no UUID for unknown key, got %v", uuid)
	}

	// add key
	k, err := as.AddAPIKey("myapikey", "test", false, nil)
	if err != nil {
		t.Fatalf("could not add API key: %v", err)
	}

	// find key
	uuid, err = as.GetUUIDForAPIKey("myapikey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid != k.UUID {
		t.Fatalf("expected UUID %v, got %v", k.UUID, uuid)
	}

	// invalid key
	k2, err := as.GetUUIDForAPIKey("invalid===")
	if err != nil && k2 != "" {
		t.Fatalf("expected empty uuid and no error, got %v and uuid %s", err, k2)
	}
}

func TestHashAPIKey(t *testing.T) {
	valid := "abcDEF123-_." // valid
	hash1, err := hashAPIKey(valid)
	if err != nil {
		t.Fatalf("unexpected error for valid key: %v", err)
	}
	hash2, err := hashAPIKey(valid)
	if err != nil {
		t.Fatalf("unexpected error for valid key: %v", err)
	}
	if hash1 != hash2 {
		t.Fatalf("expected consistent hash, got %v and %v", hash1, hash2)
	}

	// invalid Key
	_, err = hashAPIKey("invalid===")
	if err == nil || !strings.Contains(err.Error(), "One or more characters are not permitted") {
		t.Fatalf("expected error for invalid key, got %v", err)
	}
}
