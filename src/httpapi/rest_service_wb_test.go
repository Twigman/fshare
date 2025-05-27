package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/store"
	"github.com/twigman/fshare/src/testutil/fake"
)

func InitTestServices(cfg *config.Config) (*store.APIKeyService, *store.ResourceService, *RESTService, error) {
	db, err := store.NewDB(cfg.SQLitePath)
	if err != nil {
		return nil, nil, nil, err
	}

	rs := store.NewResourceService(cfg, db)
	as := store.NewAPIKeyService(db)

	restService := NewRESTService(cfg, as, rs)

	return as, rs, restService, nil
}

func SetupExistingTestUpload(uploadDir string, apiKey string, filename string, isPrivate bool, keyHighlyTrusted bool) (*RESTService, *store.ResourceService, *store.APIKeyService, *store.APIKey, string, error) {
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
		EnvPath:         filepath.Join(uploadDir, "test.env"),
	}

	as, rs, restService, err := InitTestServices(cfg)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	key, err := as.AddAPIKey(apiKey, "test key", keyHighlyTrusted)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	_, err = rs.GetOrCreateHomeDir(key.HashedKey)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	content := []byte("Hello World")
	file := &fake.FakeMultipartFile{Reader: bytes.NewReader(content)}

	r := &store.Resource{
		Name:              filename,
		IsPrivate:         isPrivate,
		APIKeyUUID:        key.UUID,
		AutoDeleteInHours: 0,
	}

	fileUUID, err := rs.SaveUploadedFile(file, r)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	return restService, rs, as, key, fileUUID, nil
}

// Tests valid url
// manipulated uuid
// manipulated expires
func TestValidateSignedRequest(t *testing.T) {
	// setup
	uploadDir := t.TempDir()
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
		EnvPath:         filepath.Join(uploadDir, "test.env"),
	}
	_, _, s, err := InitTestServices(cfg)
	if err != nil {
		t.Errorf("could not init test services %v", err)
	}

	uuid := "test-uuid"
	expiry := time.Now().Add(5 * time.Minute)
	signedURL, err := s.generateSignedURL("/test", uuid, expiry)
	if err != nil {
		t.Errorf("could not create url: %v", err)
	}

	// create request
	req := httptest.NewRequest(http.MethodGet, signedURL, nil)

	if !s.isValidSignedRequest(req, uuid) {
		t.Errorf("expected valid signature, got invalid")
	}

	// manipulate URL (change UUID)
	badReq := httptest.NewRequest(http.MethodGet, signedURL, nil)
	q := badReq.URL.Query()
	badReq.URL.Path = "/test/other-uuid"
	badReq.URL.RawQuery = q.Encode()

	if s.isValidSignedRequest(badReq, "other-uuid") {
		t.Errorf("expected invalid signature, but got valid")
	}

	// manipulate expires
	badReq2 := httptest.NewRequest(http.MethodGet, signedURL, nil)
	q2 := badReq2.URL.Query()
	q2.Set("expires", "9999999999") // Fake future
	badReq2.URL.RawQuery = q2.Encode()

	if s.isValidSignedRequest(badReq2, uuid) {
		t.Errorf("expected invalid signature due to manipulated expires, but got valid")
	}
}
