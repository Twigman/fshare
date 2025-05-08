package store

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db *sql.DB
}

func NewDB(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	s := &SQLite{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

// init initializes the database and explicitly allows foreign keys
func (s *SQLite) init() error {
	_, err := s.db.Exec(`PRAGMA foreign_keys = ON`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
	CREATE TABLE IF NOT EXISTS api_key (
		uuid TEXT PRIMARY KEY,
		hashed_key TEXT UNIQUE,
		comment TEXT,
		created_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS resource (
		uuid TEXT PRIMARY KEY,
		name TEXT,
		is_private BOOLEAN,
		is_file BOOLEAN,
		parent_uuid TEXT,
		api_key_uuid TEXT,
		autodelete_in_hours INTEGER,
		created_at DATETIME,
		deleted_at DATETIME,
		FOREIGN KEY (api_key_uuid) REFERENCES api_key(uuid) ON DELETE CASCADE,
		FOREIGN KEY (parent_uuid) REFERENCES resource(uuid) ON DELETE SET NULL
	);
	`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_resource
		ON resource(name, parent_uuid, api_key_uuid)
		WHERE deleted_at IS NULL;
	`)
	return err
}

// insertResource saves a resource
func (s *SQLite) insertResource(r *Resource) error {
	_, err := s.db.Exec(`
		INSERT INTO resource (
			uuid, name, is_private, is_file, parent_uuid, api_key_uuid, autodelete_in_hours, created_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.UUID, r.Name, r.IsPrivate, r.IsFile, r.ParentUUID, r.APIKeyUUID, r.AutoDeleteInHours, r.CreatedAt, r.DeletedAt)

	if err != nil {
		return err
	}
	return nil
}

func (s *SQLite) findResourceByUUID(uuid string) (*Resource, error) {
	row := s.db.QueryRow(`SELECT uuid, name, is_private, is_file, parent_uuid, api_key_uuid, autodelete_in_hours, created_at, deleted_at FROM resource WHERE uuid = ?`, uuid)
	var r Resource
	if err := row.Scan(&r.UUID, &r.Name, &r.IsPrivate, &r.IsFile, &r.ParentUUID, &r.APIKeyUUID, &r.AutoDeleteInHours, &r.CreatedAt, &r.DeletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *SQLite) findActiveResource(name string, apiKeyUUID string, parentDir *string) (*Resource, error) {
	var row *sql.Row
	if parentDir == nil {
		row = s.db.QueryRow(`
			SELECT uuid, name, is_private, is_file, parent_uuid, api_key_uuid,
			       autodelete_in_hours, created_at, deleted_at
			FROM resource
			WHERE name = ?
			  AND api_key_uuid = ?
			  AND deleted_at IS NULL
			  AND parent_uuid IS NULL
		`, name, apiKeyUUID)
	} else {
		row = s.db.QueryRow(`
			SELECT uuid, name, is_private, is_file, parent_uuid, api_key_uuid,
			       autodelete_in_hours, created_at, deleted_at
			FROM resource
			WHERE name = ?
			  AND api_key_uuid = ?
			  AND deleted_at IS NULL
			  AND parent_uuid = (
			    SELECT uuid FROM resource
			    WHERE name = ?
			      AND api_key_uuid = ?
			      AND deleted_at IS NULL
			  )
		`, name, apiKeyUUID, *parentDir, apiKeyUUID)
	}

	var r Resource
	if err := row.Scan(&r.UUID, &r.Name, &r.IsPrivate, &r.IsFile, &r.ParentUUID,
		&r.APIKeyUUID, &r.AutoDeleteInHours, &r.CreatedAt, &r.DeletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *SQLite) updateResource(r *Resource) error {
	_, err := s.db.Exec(`
		UPDATE resource
		SET name = ?,
		    is_private = ?,
		    is_file = ?,
		    parent_uuid = ?,
		    api_key_uuid = ?,
		    autodelete_in_hours = ?,
		    created_at = ?,
		    deleted_at = ?
		WHERE uuid = ?
	`, r.Name, r.IsPrivate, r.IsFile, r.ParentUUID, r.APIKeyUUID, r.AutoDeleteInHours, r.CreatedAt, r.DeletedAt, r.UUID)
	return err
}

// insertAPIKey saves a hashed API key, a comment and the timestamp
func (s *SQLite) insertAPIKey(key *APIKey) error {
	_, err := s.db.Exec(`
		INSERT INTO api_key (uuid, hashed_key, comment, created_at)
		VALUES (?, ?, ?, ?)
	`, key.UUID, key.HashedKey, key.Comment, key.CreatedAt)

	if err != nil {
		return fmt.Errorf("error adding API key: %v", err)
	}

	return nil
}

// countApiKeyEntries counts the entries in table api_key
func (s *SQLite) countApiKeyEntries() (int, error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM api_key`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// findAPIKeyByHash finds and returns the api_key entry containing the hashed key
func (s *SQLite) findAPIKeyByHash(hash string) (*APIKey, error) {
	row := s.db.QueryRow(`SELECT uuid, hashed_key, comment, created_at FROM api_key WHERE hashed_key = ?`, hash)
	var k APIKey
	if err := row.Scan(&k.UUID, &k.HashedKey, &k.Comment, &k.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}
