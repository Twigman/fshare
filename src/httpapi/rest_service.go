package httpapi

import (
	"fmt"
	"net/http"
	"strings"

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

func (s *RESTService) authorizeBearer(w http.ResponseWriter, r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return "", fmt.Errorf("missing authorization")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		http.Error(w, "Invalid Authorization scheme", http.StatusUnauthorized)
		return "", fmt.Errorf("invalid scheme")
	}

	apiKey := strings.TrimPrefix(authHeader, prefix)
	keyUUID, err := s.apiKeyService.GetUUIDForAPIKey(apiKey)
	if err != nil {
		http.Error(w, "Authorization failed", http.StatusUnauthorized)
		return "", err
	}

	return keyUUID, nil
}
