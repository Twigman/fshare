package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/store"
)

type RESTService struct {
	config          *config.Config
	apiKeyService   *store.APIKeyService
	resourceService *store.ResourceService
}

func NewRESTService(config *config.Config, as *store.APIKeyService, rs *store.ResourceService) *RESTService {
	return &RESTService{
		config:          config,
		apiKeyService:   as,
		resourceService: rs,
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
	if err != nil || keyUUID == "" {
		http.Error(w, "Authorization failed", http.StatusUnauthorized)
		return "", fmt.Errorf("invalid api key")
	}
	return keyUUID, nil
}
