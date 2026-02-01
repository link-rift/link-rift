package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NoFile(t *testing.T) {
	// Point config dir to temp location
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.APIBaseURL != "http://localhost:8080" {
		t.Errorf("APIBaseURL = %q, want %q", cfg.APIBaseURL, "http://localhost:8080")
	}
	if cfg.AccessToken != "" {
		t.Errorf("AccessToken = %q, want empty", cfg.AccessToken)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &CLIConfig{
		APIBaseURL:   "https://api.example.com",
		AccessToken:  "test-token-123",
		RefreshToken: "refresh-token-456",
		WorkspaceID:  "ws-789",
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify file exists with correct permissions
	info, err := os.Stat(ConfigPath())
	if err != nil {
		t.Fatalf("Config file not found: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Config file permissions = %o, want 0600", info.Mode().Perm())
	}

	// Load it back
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loaded.APIBaseURL != cfg.APIBaseURL {
		t.Errorf("APIBaseURL = %q, want %q", loaded.APIBaseURL, cfg.APIBaseURL)
	}
	if loaded.AccessToken != cfg.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, cfg.AccessToken)
	}
	if loaded.RefreshToken != cfg.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, cfg.RefreshToken)
	}
	if loaded.WorkspaceID != cfg.WorkspaceID {
		t.Errorf("WorkspaceID = %q, want %q", loaded.WorkspaceID, cfg.WorkspaceID)
	}
}

func TestClearAuth(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &CLIConfig{
		APIBaseURL:   "https://api.example.com",
		AccessToken:  "test-token-123",
		RefreshToken: "refresh-token-456",
		WorkspaceID:  "ws-789",
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	if err := ClearAuth(cfg); err != nil {
		t.Fatalf("ClearAuth() error = %v", err)
	}

	if cfg.AccessToken != "" {
		t.Errorf("AccessToken = %q, want empty", cfg.AccessToken)
	}
	if cfg.RefreshToken != "" {
		t.Errorf("RefreshToken = %q, want empty", cfg.RefreshToken)
	}
	// WorkspaceID should be preserved
	if cfg.WorkspaceID != "ws-789" {
		t.Errorf("WorkspaceID = %q, want %q", cfg.WorkspaceID, "ws-789")
	}

	// Verify it was saved to disk
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if loaded.AccessToken != "" {
		t.Errorf("Persisted AccessToken = %q, want empty", loaded.AccessToken)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".linkrift")
	os.MkdirAll(configDir, 0700)
	os.WriteFile(filepath.Join(configDir, "config.json"), []byte("not json"), 0600)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() expected error for invalid JSON")
	}
}

func TestLoadConfig_EmptyBaseURL(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".linkrift")
	os.MkdirAll(configDir, 0700)

	cfg := CLIConfig{WorkspaceID: "ws-123"}
	data, _ := json.Marshal(cfg)
	os.WriteFile(filepath.Join(configDir, "config.json"), data, 0600)

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loaded.APIBaseURL != "http://localhost:8080" {
		t.Errorf("APIBaseURL = %q, want default", loaded.APIBaseURL)
	}
}
