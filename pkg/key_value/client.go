package key_value

import (
	"context"
	"time"
)

type KeyValueResult []byte

func (o KeyValueResult) String() string {
	return string(o)
}

func (o KeyValueResult) Bytes() []byte {
	return o
}

type IKeyValue interface {
	Get(ctx context.Context, key string) (KeyValueResult, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	// Incr atomically increments the integer stored at key and returns the new
	// value. A missing key is treated as 0, so the first call returns 1. The key
	// is created without a TTL; call Expire to bound it.
	Incr(ctx context.Context, key string) (int64, error)
	// SetNX sets key to value with the given expiration only if key does not
	// already exist. It reports whether the value was set.
	SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error)
	// Expire (re)sets the TTL of an existing key. A non-existent key is a no-op.
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Close() error
}

var (
	_ IKeyValue = (*redisStorage)(nil)
	_ IKeyValue = (*inMemory)(nil)
)
