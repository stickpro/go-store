package key_value

import (
	"context"
	"fmt"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type inMemory struct {
	client *ttlcache.Cache[string, []byte]
}

func NewInMemory() IKeyValue {
	return &inMemory{
		ttlcache.New[string, []byte](),
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

func (im *inMemory) Delete(_ context.Context, key string) error {
	im.client.Delete(key)
	return nil
}

func (im *inMemory) Close() error {
	return nil
}
