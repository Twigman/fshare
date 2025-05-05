package store

import (
	"crypto/sha256"
	"encoding/hex"
)

type APIKeyService struct {
	db *SQLite
}

func NewAPIKeyService(db *SQLite) *APIKeyService {
	return &APIKeyService{db: db}
}

func (a *APIKeyService) AddAPIKey(api_key string, comment string) (string, error) {
	hashedKey := HashAPIKey(api_key)
	err := a.db.saveAPIKey(hashedKey, comment)
	if err != nil {
		return "", err
	}
	return hashedKey, nil
}

func (a *APIKeyService) AnyAPIKeyExists() (bool, error) {
	return a.db.anyAPIKeyExists()
}

func IsValidAPIKey(api_key string) bool {
	HashAPIKey(api_key)

	return true
}

func HashAPIKey(api_key string) string {
	hash := sha256.Sum256([]byte(api_key))
	return hex.EncodeToString(hash[:])
}
