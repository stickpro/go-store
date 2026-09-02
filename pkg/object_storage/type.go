package object_storage

import (
	"context"
	"errors"
)

// ErrNotFound is returned by Read when the object does not exist.
var ErrNotFound = errors.New("object_storage: not found")

type IObjectStorage interface {
	Save(ctx context.Context, path string, data []byte) (string, error)
	Get(ctx context.Context, path string) (string, error)
	Read(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	URL(ctx context.Context, path string) (string, error)
}
