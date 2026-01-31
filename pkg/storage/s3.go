package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/link-rift/link-rift/internal/config"
)

// S3Storage implements ObjectStorage using an S3-compatible backend.
// This is currently a stub implementation. When the AWS SDK v2 dependency is
// available, replace the method bodies with real PutObject / GetObject /
// DeleteObject calls.
type S3Storage struct {
	cfg config.S3Config
}

// Compile-time check that S3Storage satisfies ObjectStorage.
var _ ObjectStorage = (*S3Storage)(nil)

// NewS3Storage creates a new S3Storage instance.
// Returns an error if required configuration fields are missing.
func NewS3Storage(cfg config.S3Config) (*S3Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3: bucket name is required")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("s3: region is required")
	}
	return &S3Storage{cfg: cfg}, nil
}

// Upload stores data under the given key in S3.
// Stub: returns an error until the AWS SDK is integrated.
func (s *S3Storage) Upload(_ context.Context, _ string, _ []byte, _ string) (string, error) {
	return "", fmt.Errorf("s3: upload not implemented — AWS SDK is not yet integrated")
}

// Get retrieves the object stored under the given key from S3.
// Stub: returns an error until the AWS SDK is integrated.
func (s *S3Storage) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, fmt.Errorf("s3: get not implemented — AWS SDK is not yet integrated")
}

// Delete removes the object stored under the given key from S3.
// Stub: returns an error until the AWS SDK is integrated.
func (s *S3Storage) Delete(_ context.Context, _ string) error {
	return fmt.Errorf("s3: delete not implemented — AWS SDK is not yet integrated")
}

// GetURL returns the public URL for the given key.
// If a custom endpoint is configured (e.g. MinIO), the URL uses
// {endpoint}/{bucket}/{key}. Otherwise it uses the standard S3 virtual-hosted
// style: https://{bucket}.s3.{region}.amazonaws.com/{key}.
func (s *S3Storage) GetURL(key string) string {
	if s.cfg.Endpoint != "" {
		endpoint := strings.TrimRight(s.cfg.Endpoint, "/")
		return fmt.Sprintf("%s/%s/%s", endpoint, s.cfg.Bucket, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.Bucket, s.cfg.Region, key)
}
