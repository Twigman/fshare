package store

import (
	"database/sql"
	"fmt"
	"time"

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

func (s *SQLite) saveResource(r Resource) error {
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

func (s *SQLite) saveFile(uuid string, name string, isPrivate bool, ownerHashedKey string, autoDel int) error {
	_, err := s.db.Exec(`
		INSERT INTO resource (
			uuid, name, is_private, is_file, parent_uuid, owner_hashed_key, created_at, autodelete_in_hours
		) VALUES (?, ?, ?, TRUE, NULL, ?, ?, ?)
	`, uuid, name, isPrivate, ownerHashedKey, time.Now().UTC().Format(time.RFC3339), autoDel)

	if err != nil {
		return err
	}
	return nil
}

// SaveAPIKey saves a hashed API key and a comment
func (s *SQLite) saveAPIKey(hash string, comment string) error {
	_, err := s.db.Exec(`
		INSERT INTO api_key (hashed_key, comment, created_at)
		VALUES (?, ?, ?)
	`, hash, comment, time.Now().UTC().Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("error adding API key: %v", err)
	}

	return nil
}

// AnyAPIKeyExists checks whether an entry is stored in table api_key
func (s *SQLite) anyAPIKeyExists() (bool, error) {
	row := s.db.QueryRow(`SELECT 1 FROM api_key LIMIT 1`)

	var dummy int
	err := row.Scan(&dummy)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
