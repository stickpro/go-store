package object_storage

import "context"

type IObjectStorage interface {
	Save(ctx context.Context, path string, data []byte) (string, error)
	Get(ctx context.Context, path string) (string, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	URL(ctx context.Context, path string) (string, error)
}
