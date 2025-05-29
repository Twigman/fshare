package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/twigman/fshare/src/internal/apperror"
)

var apiKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_.]+$`)

type APIKeyService struct {
	db *SQLite
}

func NewAPIKeyService(db *SQLite) *APIKeyService {
	return &APIKeyService{db: db}
}

func (a *APIKeyService) AddAPIKey(apiKey string, comment string, isHighlyTrusted bool, createdBy *string) (*APIKey, error) {
	key_uuid, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("UUID generation error: %v", err)
	}

	hashedKey, err := hashAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	key := &APIKey{
		UUID:            key_uuid.String(),
		HashedKey:       hashedKey,
		Comment:         comment,
		IsHighlyTrusted: isHighlyTrusted,
		CreatedAt:       time.Now().UTC(),
		CreatedBy:       createdBy,
	}

	err = a.db.insertAPIKey(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (a *APIKeyService) AnyAPIKeyExists() bool {
	count, err := a.db.countApiKeyEntries()
	if err != nil {
		return false
	}

	return count > 0
}

func (a *APIKeyService) IsAPIKeyHighlyTrusted(keyUUID string) (bool, error) {
	key, err := a.db.findAPIKeyByUUID(keyUUID)
	if err != nil {
		return false, err
	}

	if key == nil {
		return false, nil
	}

	if key.IsHighlyTrusted {
		return true, nil
	}
	return false, nil
}

func (a *APIKeyService) GetUUIDForAPIKey(apiKey string) (string, error) {
	hash, err := hashAPIKey(apiKey)
	if err != nil {
		return "", nil
	}

	key, err := a.db.findAPIKeyByHash(hash)
	if err != nil {
		return "", err
	}

	if key == nil {
		return "", nil
	}
	return key.UUID, nil
}

func hashAPIKey(apiKey string) (string, error) {
	if !apiKeyRegex.MatchString(apiKey) {
		err := *apperror.ErrCharsNotAllowed
		err.Msg = fmt.Sprintf("%s (allowed: a-z, A-Z, 0-9, -, _, .)", err.Msg)
		return "", apperror.ErrCharsNotAllowed
	}
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:]), nil
}
