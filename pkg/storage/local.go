package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements ObjectStorage using the local filesystem.
// This is intended for development and testing environments.
type LocalStorage struct {
	basePath string
	baseURL  string
}

// Compile-time check that LocalStorage satisfies ObjectStorage.
var _ ObjectStorage = (*LocalStorage)(nil)

// NewLocalStorage creates a new LocalStorage instance.
// basePath is the root directory for stored files (defaults to "./data/uploads/"
// when empty). baseURL is the public URL prefix used to construct download URLs
// (e.g. "http://localhost:8080/uploads/").
func NewLocalStorage(basePath string, baseURL string) *LocalStorage {
	if basePath == "" {
		basePath = "./data/uploads/"
	}
	// Ensure trailing separators for clean path joining.
	if !strings.HasSuffix(basePath, string(os.PathSeparator)) {
		basePath += string(os.PathSeparator)
	}
	if baseURL != "" && !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// Upload writes data to basePath/key, creating intermediate directories as
// needed. It returns the public URL for the newly stored file.
func (l *LocalStorage) Upload(_ context.Context, key string, data []byte, _ string) (string, error) {
	fullPath := filepath.Join(l.basePath, key)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("local storage: creating directories %s: %w", dir, err)
	}

	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", fmt.Errorf("local storage: writing file %s: %w", fullPath, err)
	}

	return l.GetURL(key), nil
}

// Get reads the file at basePath/key and returns its contents.
func (l *LocalStorage) Get(_ context.Context, key string) ([]byte, error) {
	fullPath := filepath.Join(l.basePath, key)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local storage: file not found: %s", key)
		}
		return nil, fmt.Errorf("local storage: reading file %s: %w", fullPath, err)
	}

	return data, nil
}

// Delete removes the file at basePath/key.
func (l *LocalStorage) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(l.basePath, key)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already gone; treat as success.
		}
		return fmt.Errorf("local storage: deleting file %s: %w", fullPath, err)
	}

	return nil
}

// GetURL returns the public URL for the given key by joining baseURL and key.
func (l *LocalStorage) GetURL(key string) string {
	return l.baseURL + key
}
