package storage

import (
	"context"
	"io"
	"time"
)

const (
	TypeS3    = "s3"
	TypeLocal = "local"
)

// FileStorage abstracts object storage (S3-compatible, local filesystem, etc.).
type FileStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, string, error)
	URL(key string) string
	SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
