package storage

import "context"

// ObjectStorage abstracts file storage operations (S3, local filesystem, etc.).
type ObjectStorage interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	GetURL(key string) string
}
