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
		created_at DATETIME,
		autodelete_in_hours INTEGER,
		FOREIGN KEY (owner_hashed_key) REFERENCES api_key(hashed_key) ON DELETE CASCADE,
		FOREIGN KEY (parent_uuid) REFERENCES resource(uuid) ON DELETE SET NULL
	);
	`)
	return err
}

func (s *SQLite) saveResource(r *Resource) error {
	_, err := s.db.Exec(`
		INSERT INTO resource (
			uuid, name, is_private, is_file, parent_uuid, owner_hashed_key, created_at, autodelete_in_hours
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, r.UUID, r.Name, r.IsPrivate, r.IsFile, r.ParentUUID, r.OwnerHashedKey, r.CreatedAt, r.AutoDeleteInHours)

	if err != nil {
		return err
	}
	return nil
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

func (s *SQLite) countApiKeyEntries() (int, error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM api_key`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

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
