package config

import (
	"os"
	"path/filepath"
	"testing"
)

var testConfigMap = map[string]string{
	"valid": `{
		"port": 8080,
		"upload_path": "/tmp",
		"max_file_size_in_mb": 5
	}`,
	"noMaxFileSize": `{
		"port": 8080,
		"upload_path": "/tmp"
	}`,
	"noUploadPath": `{
		"port": 8080,
		"max_file_size_in_mb": 5
	}`,
	"noPort": `{
		"upload_path": "/tmp",
		"max_file_size_in_mb": 5
	}`,
	"invalidJSON": `{
		"port": 8080,
		"upload_path": "/tmp",
		"max_file_size_in_mb": 5`,
	"wrongFieldType_maxFileSizeIsString": `{
		"port": 8080,
		"upload_path": "/tmp",
		"max_file_size_in_mb": "five"
	}`,
	"wrongFieldType_portIsString": `{
		"port": "8080",
		"upload_path": "/tmp",
		"max_file_size_in_mb": 5
	}`,
	"valid_2": `{
		"port": 6100,
		"upload_path": "/tmp",
		"max_file_size_in_mb": 0
	}`,
	"invalidPort": `{
		"port": -1,
		"upload_path": "/tmp",
		"max_file_size_in_mb": 1
	}`,
	"emptyUploadPath": `{
		"port": 8080,
		"upload_path": "",
		"max_file_size_in_mb": 1
	}`,
}

func writeTestConfigToTempFile(t *testing.T, json string) string {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(tmpFile, []byte(json), 0644); err != nil {
		t.Fatalf("could not write temp config file: %v", err)
	}
	return tmpFile
}

func TestLoadConfig(t *testing.T) {
	const uploadLimitBytes5MB = 5 << 20
	const uploadLimitBytes1MB = 1 << 20

	tests := []struct {
		name               string
		key                string
		expectErr          bool
		expectedUploadPath string
		expectedPort       int
		limitedUpload      bool
		MaxFileSizeInBytes int64
	}{
		{"valid config", "valid", false, "/tmp", 8080, true, uploadLimitBytes5MB},
		{"no max file size", "noMaxFileSize", false, "/tmp", 8080, false, 0},
		{"no upload path", "noUploadPath", false, "", 8080, true, uploadLimitBytes5MB},
		{"no port", "noPort", false, "/tmp", 0, true, uploadLimitBytes5MB},
		{"invalid JSON", "invalidJSON", true, "/tmp", 8080, false, 0},
		{"wrong field type - max file size", "wrongFieldType_maxFileSizeIsString", true, "/tmp", 8080, false, 0},
		{"wrong field type - port", "wrongFieldType_portIsString", true, "/tmp", 0, true, uploadLimitBytes5MB},
		{"valid config 2", "valid_2", false, "/tmp", 6100, false, 0},
		{"invalid port", "invalidPort", false, "/tmp", -1, true, uploadLimitBytes1MB},
		{"empty upload path", "emptyUploadPath", false, "", 8080, true, uploadLimitBytes1MB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json, ok := testConfigMap[tt.key]
			if !ok {
				t.Fatalf("test %q not found", tt.key)
			}

			path := writeTestConfigToTempFile(t, json)
			cfg, err := LoadConfig(path)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.UploadPath != tt.expectedUploadPath {
				t.Errorf("Unexpected upload_path: %s", cfg.UploadPath)
			}

			if cfg.Port != tt.expectedPort {
				t.Errorf("expected port %d, got %d", tt.expectedPort, cfg.Port)
			}

			if cfg.IsUploadLimited() != tt.limitedUpload {
				t.Error("Expected upload to be limited")
			}

			if cfg.MaxFileSizeBytes() != tt.MaxFileSizeInBytes {
				t.Errorf("Expected %d bytes upload limit, got %d", tt.MaxFileSizeInBytes, cfg.MaxFileSizeBytes())
			}
		})
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

func TestValidate(t *testing.T) {
	// only tests that can be loaded successfully
	tests := []struct {
		name           string
		key            string
		expectValidErr bool
		expectErrMsg   string
	}{
		{"valid config", "valid", false, ""},
		{"no max file size", "noMaxFileSize", false, ""},
		{"no upload path", "noUploadPath", true, "upload_path is required"},
		{"no port", "noPort", true, "port value is probably not set (0 is not allowed)"},
		{"valid config 2", "valid_2", false, ""},
		{"invalid port", "invalidPort", true, "port value is not valid"},
		{"empty upload path", "emptyUploadPath", true, "upload_path is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json, ok := testConfigMap[tt.key]
			if !ok {
				t.Fatalf("fixture %q not found", tt.key)
			}

			path := writeTestConfigToTempFile(t, json)
			cfg, err := LoadConfig(path)

			if err != nil {
				t.Errorf("Error loading config for validation: %v", err)
			}

			err = cfg.Validate()

			if tt.expectValidErr {
				if err == nil {
					t.Errorf("Expected validation error, but got nil")
				}

				if err.Error() != tt.expectErrMsg {
					t.Errorf("Expected error msg %q, but got %q", tt.expectErrMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, but got: %v", err)
				}
			}
		})
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
