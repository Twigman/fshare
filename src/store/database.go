package store

import "time"

type Database interface {
	init() error
	saveResource(r Resource) error
	saveFile(uuid string, name string, isPrivate bool, ownerHashedKey string, autoDel int) error
	saveAPIKey(hash string, comment string) error
}

type Resource struct {
	UUID              string
	Name              string
	IsPrivate         bool
	IsFile            bool
	ParentUUID        *string
	OwnerHashedKey    string
	CreatedAt         time.Time
	AutoDeleteInHours int
}
