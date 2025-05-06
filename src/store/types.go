package store

import "time"

type Resource struct {
	UUID              string
	Name              string
	IsPrivate         bool
	IsFile            bool
	ParentUUID        *string
	OwnerHashedKey    string
	AutoDeleteInHours int
	CreatedAt         time.Time
	DeletedAt         *time.Time
}

type APIKey struct {
	HashedKey string
	Comment   string
	CreatedAt time.Time
}
