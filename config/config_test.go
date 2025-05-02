package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.json")
	json := `{
		"port": 8080,
		"upload_path": "/tmp",
		"max_file_size_in_mb": 5
	}`
	if err := os.WriteFile(tmpFile, []byte(json), 0644); err != nil {
		t.Fatalf("Could not write temp config file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}
	if cfg.UploadPath != "/tmp" {
		t.Errorf("Unexpected upload_path: %s", cfg.UploadPath)
	}
	if !cfg.IsUploadLimited() {
		t.Error("Expected upload to be limited")
	}
	if cfg.MaxFileSizeBytes() != 5<<20 {
		t.Errorf("Expected 5 MB in bytes, got %d", cfg.MaxFileSizeBytes())
	}
}

func TestLoadConfig_FileNotExist(t *testing.T) {
	noTmpFile := filepath.Join(t.TempDir(), "no_config.json")
	cfg, err := LoadConfig(noTmpFile)

	if err == nil {
		t.Error("Expected error when loading non-existent config file, got nil:", noTmpFile)
	}

	if cfg != nil {
		t.Errorf("Expected nil config, got: %+v", cfg)
	}
}

func TestLoadConfig_Default_NoMaxFileSize(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.json")

	partialJSON := `{
		"port": 8080,
		"upload_path": "/tmp"
	}`

	if err := os.WriteFile(tmpFile, []byte(partialJSON), 0644); err != nil {
		t.Fatalf("Could not write test file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.MaxFileSizeInMB != 0 {
		t.Errorf("Expected default MaxFileSizeInMB == 0, got: %d", cfg.MaxFileSizeInMB)
	}
	if cfg.IsUploadLimited() {
		t.Error("Expected IsUploadLimited() to be false for default value")
	}
}

func TestLoadConfig_Default_NoUploadPath(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.json")

	partialJSON := `{
		"port": 8080,
		"max_file_size_in_mb": 5
	}`

	if err := os.WriteFile(tmpFile, []byte(partialJSON), 0644); err != nil {
		t.Fatalf("Could not write test file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.UploadPath != "" {
		t.Errorf("Expected default upload_path == \"\", got: %s", cfg.UploadPath)
	}

	err = cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for upload_path, got: %v", err)
	}
}

func TestLoadConfig_Default_NoPort(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.json")

	partialJSON := `{
		"upload_path": "/tmp",
		"max_file_size_in_mb": 5
	}`

	if err := os.WriteFile(tmpFile, []byte(partialJSON), 0644); err != nil {
		t.Fatalf("Could not write test file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Port != 0 {
		t.Errorf("Expected default Port == 0, got: %d", cfg.Port)
	}

	err = cfg.Validate()
	if err == nil {
		t.Errorf("Expected port value validation error, got: %v", err)
	}
}

func TestLoadConfig_InvalidJSONFormat(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad_config.json")

	badJSON := `{
		"port": 8080,
		"upload_path": "/tmp",
		"max_file_size_in_mb": 5`

	if err := os.WriteFile(tmpFile, []byte(badJSON), 0644); err != nil {
		t.Fatalf("Could not write test file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got: %v", err)
	}
	if cfg != nil {
		t.Errorf("Expected nil config, got: %+v", cfg)
	}
}

func TestLoadConfig_WrongFieldType_MaxFileSize(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad_config.json")

	badJSON := `{
		"port": 8080,
		"upload_path": "/tmp",
		"max_file_size_in_mb": "five"
		}`

	if err := os.WriteFile(tmpFile, []byte(badJSON), 0644); err != nil {
		t.Fatalf("Could not write test file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got: %v", err)
	}
	if cfg != nil {
		t.Errorf("Expected nil config, got: %+v", cfg)
	}
}

func TestLoadConfig_WrongFieldType_PortAsString(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad_config.json")

	badJSON := `{
		"port": "8080",
		"upload_path": "/tmp",
		"max_file_size_in_mb": 5
		}`

	if err := os.WriteFile(tmpFile, []byte(badJSON), 0644); err != nil {
		t.Fatalf("Could not write test file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got: %v", err)
	}
	if cfg != nil {
		t.Errorf("Expected nil config, got: %+v", cfg)
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := &Config{
		Port:            8080,
		UploadPath:      t.TempDir(),
		MaxFileSizeInMB: 0,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := &Config{
		Port:            -1,
		UploadPath:      t.TempDir(),
		MaxFileSizeInMB: 1,
	}
	err := cfg.Validate()
	if err == nil || err.Error() != "port value is not valid" {
		t.Errorf("Expected invalid port error, got: %v", err)
	}
}

func TestValidate_EmptyUploadPath(t *testing.T) {
	cfg := &Config{
		Port:            8080,
		UploadPath:      "",
		MaxFileSizeInMB: 1,
	}
	err := cfg.Validate()
	if err == nil || err.Error() != "upload_path is required" {
		t.Errorf("Expected upload_path error, got: %v", err)
	}
}

func TestValidate_NonExistentUploadPath(t *testing.T) {
	cfg := &Config{
		Port:            8080,
		UploadPath:      filepath.Join(t.TempDir(), "nonexistent"),
		MaxFileSizeInMB: 1,
	}
	err := cfg.Validate()
	if err == nil || err.Error() == "" {
		t.Errorf("Expected invalid upload path error, got: %v", err)
	}
}
