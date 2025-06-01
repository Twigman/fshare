package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/twigman/fshare/src/internal/apperror"
	"github.com/twigman/fshare/src/utils"
)

type Env struct {
	HMACSecret string
}

func LoadOrCreateEnv(path string) (*Env, error) {
	filePath := filepath.Join(path, ".env")
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, apperror.ErrResourceResolvePath
	}

	if _, err := os.Stat(absPath); err == nil {
		// read
		data, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("could not read secret file: %w", err)
		}
		return &Env{HMACSecret: strings.TrimSpace(string(data))}, nil
	}

	secret, err := utils.GenerateSecret(32)
	if err != nil {
		return nil, err
	}

	env := &Env{HMACSecret: secret}

	// Datei sicher schreiben (z. B. 0600)
	if err := os.WriteFile(absPath, []byte(env.HMACSecret+"\n"), 0600); err != nil {
		return env, fmt.Errorf("could not write secret file to: %w", err)
	}
	return env, nil
}

func CreateInitDataEnv(path string, data string) (string, error) {
	filePath := filepath.Join(path, "init_data.env")
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", apperror.ErrResourceResolvePath
	}

	err = os.WriteFile(absPath, []byte(data), 0o600)
	if err != nil {
		return "", err
	}
	return absPath, nil
}
