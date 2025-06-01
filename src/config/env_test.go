package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twigman/fshare/src/config"
)

func TestLoadOrCreateEnv_CreateNew(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	env, err := config.LoadOrCreateEnv(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if env == nil {
		t.Fatalf("expected env, got nil")
	}
	if env.HMACSecret == "" {
		t.Error("expected HMACSecret to be set")
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("expected file to be created, got error %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != env.HMACSecret {
		t.Errorf("expected file content to match HMACSecret, got %s", content)
	}
}

func TestLoadOrCreateEnv_LoadExisting(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	secret := "my-secret"
	err := os.WriteFile(envPath, []byte(secret+"\n"), 0o600)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	env, err := config.LoadOrCreateEnv(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if env.HMACSecret != secret {
		t.Errorf("expected %s, got %s", secret, env.HMACSecret)
	}
}

func TestCreateInitDataEnv_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	content := "test-content"

	path, err := config.CreateInitDataEnv(dir, content)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected %s, got %s", content, string(data))
	}
}
