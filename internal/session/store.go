package session

import "context"

type Store interface {
	Create(ctx context.Context, metadata Metadata) (string, error)
	Get(ctx context.Context, token string) (*Metadata, error)
	Delete(ctx context.Context, token string) error
	Extend(ctx context.Context, token string) error
}
