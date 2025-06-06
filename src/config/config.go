package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Port            int    `json:"port"`
	DataPath        string `json:"data_path"`
	UploadPath      string `json:"upload_path"`
	MaxFileSizeInMB int64  `json:"max_file_size_in_mb"` // 0 = no limit
	//ContinuousFileValidation bool   `json:"continuous_file_validation"`
	//SpacePerUserInMB         int    `json:"space_per_user_in_mb"`
	AutoDeleteIntervalInSec int `json:"autodelete_interval_in_sec"`
	//MaxFolderDepth           int    `json:"max_folder_depth"`
}

// LoadConfig loads the configuration from a given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks the initialized configuration values
func (c *Config) Validate() error {
	// port
	if c.Port == 0 {
		return errors.New("port value is probably not set (0 is not allowed)")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("port value is not valid")
	}

	// upload_path
	if c.UploadPath == "" {
		return errors.New("upload_path is required")
	}

	return nil
}

func (c *Config) MaxFileSizeBytes() int64 {
	return c.MaxFileSizeInMB << 20
}

func (c *Config) IsUploadLimited() bool {
	return c.MaxFileSizeInMB > 0
}
