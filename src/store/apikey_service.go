package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type APIKeyService struct {
	db *SQLite
}

func NewAPIKeyService(db *SQLite) *APIKeyService {
	return &APIKeyService{db: db}
}

func (a *APIKeyService) AddAPIKey(api_key string, comment string, isHighlyTrusted bool) (*APIKey, error) {
	key_uuid, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("UUID generation error: %v", err)
	}

	key := &APIKey{
		UUID:            key_uuid.String(),
		HashedKey:       HashAPIKey(api_key),
		Comment:         comment,
		IsHighlyTrusted: isHighlyTrusted,
		CreatedAt:       time.Now().UTC(),
	}

	err = a.db.insertAPIKey(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (a *APIKeyService) AnyAPIKeyExists() (bool, error) {
	count, err := a.db.countApiKeyEntries()
	if err != nil {
		return false, err
	}

	return count > 0, nil
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

func (a *APIKeyService) GetUUIDForAPIKey(api_key string) (string, error) {
	key, err := a.db.findAPIKeyByHash(HashAPIKey(api_key))
	if err != nil {
		return "", err
	}

	if key == nil {
		return "", nil
	}
	return key.UUID, nil
}

func HashAPIKey(api_key string) string {
	hash := sha256.Sum256([]byte(api_key))
	return hex.EncodeToString(hash[:])
}
