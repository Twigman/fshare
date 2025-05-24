package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

type Env struct {
	HMACSecret string
}

func LoadOrCreateEnv(path string) (*Env, error) {
	if _, err := os.Stat(path); err == nil {
		// read
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not read secret file: %w", err)
		}
		return &Env{HMACSecret: strings.TrimSpace(string(data))}, nil
	}

	// does not exist -> generate
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("could not generate secret: %w", err)
	}

	env := &Env{HMACSecret: hex.EncodeToString(secretBytes)}

	// Datei sicher schreiben (z. B. 0600)
	if err := os.WriteFile(path, []byte(env.HMACSecret+"\n"), 0600); err != nil {
		return env, fmt.Errorf("could not write secret file to: %w", err)
	}
	return env, nil
}
