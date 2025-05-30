package store

import "time"

type Resource struct {
	UUID         string
	Name         string
	IsPrivate    bool
	IsFile       bool
	ParentUUID   *string
	APIKeyUUID   string
	AutoDeleteAt *time.Time
	CreatedAt    time.Time
	DeletedAt    *time.Time
	IsBroken     bool
}

type APIKey struct {
	UUID            string
	HashedKey       string
	Comment         string
	IsHighlyTrusted bool
	CreatedAt       time.Time
	CreatedBy       *string
}
