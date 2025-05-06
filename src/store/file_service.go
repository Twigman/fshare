package store

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/twigman/fshare/src/config"
)

type FileService struct {
	db  *SQLite
	cfg *config.Config
}

func NewFileService(cfg *config.Config, db *SQLite) *FileService {
	return &FileService{cfg: cfg, db: db}
}

func (f *FileService) SaveUploadedFile(file multipart.File, r *Resource) (string, error) {
	safeName := filepath.Base(r.Name)
	dstPath := filepath.Join(f.cfg.UploadPath, safeName)

	file_uuid, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("UUID generation error: %v", err)
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	r.UUID = file_uuid.String()
	r.IsFile = true
	r.ParentUUID = nil
	r.Name = safeName
	r.CreatedAt = time.Now().UTC()
	r.DeletedAt = nil

	err = f.db.insertResource(r)
	if err != nil {
		return "", err
	}

	return file_uuid.String(), nil
}

func (s *FileService) buildPathFromResource(r *Resource) (string, error) {
	parts := []string{r.Name}

	current := r
	for current.ParentUUID != nil {
		parent, err := s.db.findResourceByUUID(*current.ParentUUID)
		if err != nil {
			return "", err
		}
		if parent == nil {
			return "", fmt.Errorf("parent with UUID %s not found", *current.ParentUUID)
		}
		parts = append([]string{parent.Name}, parts...)
		current = parent
	}

	return filepath.Join(s.cfg.UploadPath, r.APIKeyUUID, filepath.Join(parts...)), nil
}

/*
func (f *FileService) GetOrCreateNewFolder(name string, owner_hashed_key string, path string) (*Resource, string, error) {
	parts := strings.Split(path, "/")
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}

	if (len(cleaned) + 1) > f.cfg.MaxFolderDepth {
		return nil, "", fmt.Errorf("new path is too long")
	}

	r, err := f.db.findActiveResource(name, hashed_key, parentDir)
	if err != nil {
		return nil, "", err
	}

	if r == nil {
		// folder does not exist

	}

	path, err := f.buildPathFromResource(r)
	if err != nil {
		return nil, "", err
	}
	return r, path, nil
}*/

func (f *FileService) GetOrCreateHomeDir(owner_hashed_key string) (*Resource, error) {
	key, err := f.db.findAPIKeyByHash(owner_hashed_key)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, fmt.Errorf("API key does not exist")
	}

	// name of user home dir = owner/user hash
	r, err := f.db.findActiveResource(key.UUID, key.UUID, nil)
	if err != nil {
		return nil, err
	}

	if r == nil {
		// home should not exist
		homePath := filepath.Join(f.cfg.UploadPath, key.UUID)
		err := os.Mkdir(homePath, 0o700)
		if err != nil {
			if os.IsExist(err) {
				return nil, fmt.Errorf("user directory already exists: %s", homePath)
			}
			return nil, err
		}

		home_uuid, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("UUID generation error: %v", err)
		}

		r = &Resource{
			UUID:              home_uuid.String(),
			Name:              key.UUID,
			IsPrivate:         true,
			IsFile:            false,
			ParentUUID:        nil,
			APIKeyUUID:        key.UUID,
			AutoDeleteInHours: 0,
			CreatedAt:         time.Now().UTC(),
			DeletedAt:         nil,
		}

		err = f.db.insertResource(r)
		if err != nil {
			// delete home dir
			err2 := os.Remove(homePath)
			if err2 != nil {
				return nil, fmt.Errorf("error creating home dir db for user + could not delete created home dir (%s):\ndb error: %v\nos error: %v", homePath, err, err2)
			}
			return nil, err
		}
	}
	return r, nil
}
