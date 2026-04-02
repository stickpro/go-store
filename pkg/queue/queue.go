package queue

import "context"

type IQueue interface {
	Push(ctx context.Context, name string, payload []byte) error
	// Pop blocks until a message is available or ctx is cancelled.
	Pop(ctx context.Context, name string) ([]byte, error)
	Close() error
}
