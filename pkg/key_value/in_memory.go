package key_value

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type inMemory struct {
	client *ttlcache.Cache[string, []byte]
	// mu guards the read-modify-write compound operations (Incr, SetNX) so they
	// behave atomically within this process.
	mu sync.Mutex
}

func NewInMemory() IKeyValue {
	return &inMemory{
		client: ttlcache.New[string, []byte](),
	}
}

func (im *inMemory) Get(_ context.Context, key string) (KeyValueResult, error) {
	entry := im.client.Get(key)
	if entry == nil {
		return nil, ErrEntryNotFound
	}

	return entry.Value(), nil
}

func (im *inMemory) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	if strVal, ok := value.([]byte); ok {
		_ = im.client.Set(key, strVal, expiration)
		return nil
	}
	if strVal, ok := value.(string); ok {
		_ = im.client.Set(key, []byte(strVal), expiration)
		return nil
	}

	return fmt.Errorf("unsupported value type: %T", value)
}

func (im *inMemory) Incr(_ context.Context, key string) (int64, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	var n int64
	if entry := im.client.Get(key); entry != nil {
		parsed, err := strconv.ParseInt(string(entry.Value()), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("incr: value at %q is not an integer: %w", key, err)
		}
		n = parsed
	}
	n++

	ttl := ttlcache.NoTTL
	if entry := im.client.Get(key); entry != nil {
		if remaining := time.Until(entry.ExpiresAt()); remaining > 0 {
			ttl = remaining
		}
	}
	im.client.Set(key, []byte(strconv.FormatInt(n, 10)), ttl)
	return n, nil
}

func (im *inMemory) SetNX(_ context.Context, key, value string, expiration time.Duration) (bool, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if entry := im.client.Get(key); entry != nil {
		return false, nil
	}
	im.client.Set(key, []byte(value), expiration)
	return true, nil
}

func (im *inMemory) Expire(_ context.Context, key string, expiration time.Duration) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry := im.client.Get(key)
	if entry == nil {
		return nil
	}
	im.client.Set(key, entry.Value(), expiration)
	return nil
}

func (im *inMemory) Delete(_ context.Context, key string) error {
	im.client.Delete(key)
	return nil
}

func (im *inMemory) Close() error {
	return nil
}
