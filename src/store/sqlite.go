package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db *sql.DB
}

type Resource struct {
	uuid              string
	name              string
	isPrivate         bool
	isFile            bool
	parentUUID        *string
	ownerHashedKey    string
	createdAt         time.Time
	autoDeleteInHours int
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

func (s *SQLite) SaveResource(r Resource) (string, error) {
	_, err := s.db.Exec(`
		INSERT INTO resource (
			uuid, name, is_private, is_file, parent_uuid, owner_hashed_key, created_at, autodelete_in_hours
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, r.uuid, r.name, r.isPrivate, r.isFile, r.parentUUID, r.ownerHashedKey, r.createdAt, r.autoDeleteInHours)

	if err != nil {
		return "", err
	}
	return r.uuid, nil
}

func (s *SQLite) SaveFile(name string, isPrivate bool, ownerHashedKey string) (string, error) {
	r_uuid, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("UUID generation error: %v", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO resource (
			uuid, name, is_private, is_file, parent_uuid, owner_hashed_key, created_at, autodelete_in_hours
		) VALUES (?, ?, ?, TRUE, NULL, ?, ?, 0)
	`, r_uuid.String(), name, isPrivate, ownerHashedKey, time.Now().UTC().Format(time.RFC3339))

	if err != nil {
		return "", err
	}
	return r_uuid.String(), nil
}

// SaveAPIKey hashes an API key, adds it to database and returns it
func (s *SQLite) SaveAPIKey(apikey string, comment string) (string, error) {
	hash := sha256.Sum256([]byte(apikey))
	keyHash := hex.EncodeToString(hash[:])

	_, err := s.db.Exec(`
		INSERT INTO api_key (hashed_key, owner, created_at)
		VALUES (?, ?, ?)
	`, keyHash, comment, time.Now().UTC().Format(time.RFC3339))

	if err != nil {
		return "", fmt.Errorf("error adding API key: %v", err)
	}

	return keyHash, nil
}

// AnyAPIKeyExists checks whether an entry is stored in table api_key
func (s *SQLite) AnyAPIKeyExists() (bool, error) {
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
