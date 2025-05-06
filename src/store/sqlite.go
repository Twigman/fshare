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
		hashed_key TEXT PRIMARY KEY,
		comment TEXT,
		created_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS resource (
		uuid TEXT PRIMARY KEY,
		name TEXT,
		is_private BOOLEAN,
		is_file BOOLEAN,
		parent_uuid TEXT,
		owner_hashed_key TEXT,
		autodelete_in_hours INTEGER,
		created_at DATETIME,
		deleted_at DATETIME,
		FOREIGN KEY (owner_hashed_key) REFERENCES api_key(hashed_key) ON DELETE CASCADE,
		FOREIGN KEY (parent_uuid) REFERENCES resource(uuid) ON DELETE SET NULL
	);
	`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_resource
		ON resource(name, parent_uuid, owner_hashed_key)
		WHERE deleted_at IS NULL;
	`)
	return err
}

// insertResource saves a resource
func (s *SQLite) insertResource(r *Resource) error {
	_, err := s.db.Exec(`
		INSERT INTO resource (
			uuid, name, is_private, is_file, parent_uuid, owner_hashed_key, autodelete_in_hours, created_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.UUID, r.Name, r.IsPrivate, r.IsFile, r.ParentUUID, r.OwnerHashedKey, r.AutoDeleteInHours, r.CreatedAt, r.DeletedAt)

	if err != nil {
		return err
	}
	return nil
}

func (s *SQLite) findResourceByUUID(uuid string) (*Resource, error) {
	row := s.db.QueryRow(`SELECT uuid, name, is_private, is_file, parent_uuid, owner_hashed_key, autodelete_in_hours, created_at, deleted_at FROM resource WHERE uuid = ?`, uuid)
	var r Resource
	if err := row.Scan(&r.UUID, &r.Name, &r.IsPrivate, &r.IsFile, &r.ParentUUID, &r.OwnerHashedKey, &r.AutoDeleteInHours, &r.CreatedAt, &r.DeletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *SQLite) findActiveResourceByNameAndOwner(name string, ownerHash string) (*Resource, error) {
	row := s.db.QueryRow(`SELECT uuid, name, is_private, is_file, parent_uuid, owner_hashed_key, autodelete_in_hours, created_at, deleted_at FROM resource WHERE name = ? AND owner_hashed_key = ? AND deleted_at IS NULL`, name, ownerHash)
	var r Resource
	if err := row.Scan(&r.UUID, &r.Name, &r.IsPrivate, &r.IsFile, &r.ParentUUID, &r.OwnerHashedKey, &r.AutoDeleteInHours, &r.CreatedAt, &r.DeletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// insertAPIKey saves a hashed API key, a comment and the timestamp
func (s *SQLite) insertAPIKey(key *APIKey) error {
	_, err := s.db.Exec(`
		INSERT INTO api_key (hashed_key, comment, created_at)
		VALUES (?, ?, ?)
	`, key.HashedKey, key.Comment, key.CreatedAt)

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
	row := s.db.QueryRow(`SELECT hashed_key, owner, created_at FROM api_key WHERE hashed_key = ?`, hash)
	var k APIKey
	if err := row.Scan(&k.HashedKey, &k.Comment, &k.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}
