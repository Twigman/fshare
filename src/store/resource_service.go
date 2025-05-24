package store

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/internal/apperror"
)

type ResourceService struct {
	db  *SQLite
	cfg *config.Config
}

func NewResourceService(cfg *config.Config, db *SQLite) *ResourceService {
	return &ResourceService{cfg: cfg, db: db}
}

func (s *ResourceService) BuildResourcePath(r *Resource) (string, error) {

	// make sure target path is in upload folder
	dstPath := filepath.Join(s.cfg.UploadPath, r.APIKeyUUID, r.Name)

	absBase, err := filepath.Abs(s.cfg.UploadPath)
	if err != nil {
		return "", apperror.ErrResolvePath
	}

	absDst, err := filepath.Abs(dstPath)
	if err != nil || !strings.HasPrefix(absDst, absBase) {
		return "", apperror.ErrInvalidFilepath
	}
	return dstPath, nil
}

func (s *ResourceService) SaveUploadedFile(file multipart.File, r *Resource) (string, error) {
	if strings.Contains(r.Name, "..") ||
		strings.Contains(r.Name, "/") ||
		strings.Contains(r.Name, "\\") ||
		strings.HasPrefix(r.Name, ".") {
		return "", apperror.ErrInvalidFilename
	}

	r.Name = strings.TrimSpace(r.Name)
	r.Name = filepath.Base(r.Name)

	absDst, err := s.BuildResourcePath(r)
	if err != nil {
		return "", err
	}

	fileUUID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("UUID generation error: %v", err)
	}

	tmpDir := filepath.Dir(absDst)
	tmpFile, err := os.CreateTemp(tmpDir, "upload-*")
	if err != nil {
		return "", fmt.Errorf("temp file creation error: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err = io.Copy(tmpFile, file); err != nil {
		return "", fmt.Errorf("file copy error: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("file close error: %v", err)
	}

	if err := os.Rename(tmpFile.Name(), absDst); err != nil {
		return "", fmt.Errorf("rename error: %v", err)
	}

	r.UUID = fileUUID.String()
	r.IsFile = true
	r.ParentUUID = nil
	r.CreatedAt = time.Now().UTC()
	r.DeletedAt = nil

	if err := s.db.insertResource(r); err != nil {
		return "", err
	}

	return fileUUID.String(), nil
}

func (s *ResourceService) GetOrCreateHomeDir(hashed_key string) (*Resource, error) {
	key, err := s.db.findAPIKeyByHash(hashed_key)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, fmt.Errorf("API key does not exist")
	}

	// name of user home dir = api key uuid
	r, err := s.db.findActiveResource(key.UUID, key.UUID, nil)
	if err != nil {
		return nil, err
	}

	if r == nil {
		// home should not exist
		homePath := filepath.Join(s.cfg.UploadPath, key.UUID)
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

		err = s.db.insertResource(r)
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

func (s *ResourceService) GetResourceByUUID(uuid string) (*Resource, error) {
	r, err := s.db.findResourceByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("resource not found")
	}
	return r, nil
}

func (s *ResourceService) DeleteResourceByUUID(rUUID string, keyUUID string) error {
	res, err := s.GetResourceByUUID(rUUID)
	if err != nil || res == nil || !res.IsFile || res.DeletedAt != nil {
		return err
	}

	// needs to be owner
	if res.APIKeyUUID != keyUUID {
		return fmt.Errorf("authorization error")
	}

	resPath := filepath.Join(s.cfg.UploadPath, res.APIKeyUUID, res.Name)

	// remove resource
	if err := os.Remove(resPath); err != nil {
		return err
	}

	t := time.Now().UTC()
	res.DeletedAt = &t

	err = s.db.updateResource(res)
	if err != nil {
		return err
	}
	return nil
}

func (s *ResourceService) MarkResourceAsBroken(rUUID string) error {
	res, err := s.GetResourceByUUID(rUUID)
	if err != nil || res == nil || res.DeletedAt != nil {
		return err
	}

	res.IsBroken = true
	err = s.db.updateResource(res)
	if err != nil {
		return err
	}
	return nil
}
