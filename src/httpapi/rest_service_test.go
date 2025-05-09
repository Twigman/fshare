package httpapi_test

import (
	"bytes"
	"path/filepath"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
	"github.com/twigman/fshare/src/store"
	"github.com/twigman/fshare/src/testutil/fake"
)

func initTestServices(cfg *config.Config) (*store.APIKeyService, *store.ResourceService, *httpapi.RESTService, error) {
	db, err := store.NewDB(cfg.SQLitePath)
	if err != nil {
		return nil, nil, nil, err
	}

	rs := store.NewResourceService(cfg, db)
	as := store.NewAPIKeyService(db)

	restService := httpapi.NewRESTService(cfg, as, rs)

	return as, rs, restService, nil
}

func setupExistingTestUpload(uploadDir string, apiKey string, filename string, isPrivate bool) (*httpapi.RESTService, *store.ResourceService, *store.APIKey, string, error) {
	cfg := &config.Config{
		UploadPath:      uploadDir,
		MaxFileSizeInMB: 5,
		Port:            8080,
		SQLitePath:      filepath.Join(uploadDir, "test_db.sqlite"),
	}

	as, rs, restService, err := initTestServices(cfg)
	if err != nil {
		return nil, nil, nil, "", err
	}

	key, err := as.AddAPIKey(apiKey, "test key")
	if err != nil {
		return nil, nil, nil, "", err
	}

	_, err = rs.GetOrCreateHomeDir(key.HashedKey)
	if err != nil {
		return nil, nil, nil, "", err
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
		return nil, nil, nil, "", err
	}

	return restService, rs, key, fileUUID, nil
}
