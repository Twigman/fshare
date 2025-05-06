package store

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type FileService struct {
	db         *SQLite
	uploadPath string
}

func NewFileService(uploadPath string, db *SQLite) *FileService {
	return &FileService{uploadPath: uploadPath, db: db}
}

func (f *FileService) SaveUploadedFile(file multipart.File, r *Resource) (string, error) {
	safeName := filepath.Base(r.Name)
	dstPath := filepath.Join(f.uploadPath, safeName)

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

	err = f.db.insertResource(r)
	if err != nil {
		return "", err
	}

	return file_uuid.String(), nil
}
