package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/twigman/fshare/src/internal/apperror"
)

func (s *RESTService) CreateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONStatus(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	keyUUID, err := s.authorizeBearer(w, r)
	if err != nil {
		return
	}

	trusted, err := s.apiKeyService.IsAPIKeyHighlyTrusted(keyUUID)
	if err != nil {
		writeJSONStatus(w, http.StatusInternalServerError, "Auth error")
		return
	}
	if !trusted {
		writeJSONStatus(w, http.StatusForbidden, "Not authorized to create API keys")
		return
	}

	var req APIKeyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONStatus(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	key, err := s.apiKeyService.AddAPIKey(req.Key, req.Comment, req.HighlyTrusted, &keyUUID)
	if err != nil {
		if err == apperror.ErrCharsNotAllowed || err == apperror.ErrEmptyAPIKey {
			writeJSONStatus(w, http.StatusBadRequest, "Could not create API key")
			return
		}
		writeJSONStatus(w, http.StatusInternalServerError, "Could not create API key")
		return
	}

	_, err = s.resourceService.GetOrCreateHomeDir(key.HashedKey)
	if err != nil {
		writeJSONStatus(w, http.StatusInternalServerError, "Could not create API key")
		log.Printf("Could not create home dir for API key with UUID %s: %v", key.UUID, err)
		return
	}

	res := APIKeyResponse{
		UUID:          key.UUID,
		Comment:       key.Comment,
		HighlyTrusted: key.IsHighlyTrusted,
		CreatedAt:     key.CreatedAt,
	}

	writeJSONResponse(w, http.StatusCreated, res)
}
