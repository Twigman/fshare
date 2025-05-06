package store

import "time"

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
