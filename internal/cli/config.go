package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// CLIConfig holds CLI-specific configuration persisted between sessions.
type CLIConfig struct {
	APIBaseURL   string `json:"api_base_url"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	WorkspaceID  string `json:"workspace_id,omitempty"`
}

// DefaultConfigDir returns ~/.linkrift
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".linkrift"
	}
	return filepath.Join(home, ".linkrift")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.json")
}

// LoadConfig reads the config from disk.
func LoadConfig() (*CLIConfig, error) {
	path := ConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CLIConfig{
				APIBaseURL: "http://localhost:8080",
			}, nil
		}
		return nil, err
	}

	var cfg CLIConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = "http://localhost:8080"
	}

	return &cfg, nil
}

// SaveConfig writes the config to disk.
func SaveConfig(cfg *CLIConfig) error {
	dir := DefaultConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0600)
}

// ClearAuth removes tokens from the config.
func ClearAuth(cfg *CLIConfig) error {
	cfg.AccessToken = ""
	cfg.RefreshToken = ""
	return SaveConfig(cfg)
}
