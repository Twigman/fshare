package httpapi_test

import (
	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
	"github.com/twigman/fshare/src/store"
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
