package httpapi

import (
	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/store"
)

type RESTService struct {
	config        *config.Config
	apiKeyService *store.APIKeyService
	fileService   *store.FileService
}

func NewRESTService(config *config.Config, apiKeyService *store.APIKeyService, fileService *store.FileService) *RESTService {
	return &RESTService{
		config:        config,
		apiKeyService: apiKeyService,
		fileService:   fileService,
	}
}
