package store

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type APIKeyService struct {
	db *SQLite
}

func NewAPIKeyService(db *SQLite) *APIKeyService {
	return &APIKeyService{db: db}
}

func (a *APIKeyService) AddAPIKey(api_key string, comment string) (*APIKey, error) {
	key := &APIKey{
		HashedKey: HashAPIKey(api_key),
		Comment:   comment,
		CreatedAt: time.Now().UTC(),
	}

	err := a.db.insertAPIKey(key)
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

func (a *APIKeyService) IsValidAPIKey(api_key string) (bool, error) {
	key, err := a.db.findAPIKeyByHash(HashAPIKey(api_key))
	if err != nil {
		return false, err
	}

	if key != nil {
		return true, nil
	}
	return false, nil
}

func HashAPIKey(api_key string) string {
	hash := sha256.Sum256([]byte(api_key))
	return hex.EncodeToString(hash[:])
}
