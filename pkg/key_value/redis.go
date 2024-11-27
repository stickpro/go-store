package key_value

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisStorage struct {
	client *redis.Client
}

func NewRedisStorage(client *redis.Client) IKeyValue {
	return &redisStorage{
		client,
	}
}

func (o *redisStorage) Get(ctx context.Context, key string) (KeyValueResult, error) {
	if v := o.client.Get(ctx, key); v.Err() != nil && errors.Is(v.Err(), redis.Nil) {
		return nil, fmt.Errorf("key not found: %w", v.Err())
	}
	kType, err := o.client.Type(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get key type: %w", err)
	}
	switch kType {
	case "hash":
		res, err := o.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("get hash key from redis: %w", err)
		}
		return json.Marshal(res)
	case "string":
		return o.client.Get(ctx, key).Bytes()
	default:
		return nil, fmt.Errorf("unsupported key type: %s", kType)
	}
}

func (o *redisStorage) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var err error
	switch value.(type) {
	case string:
		_, err = o.client.Set(ctx, key, value, expiration).Result()
	case map[string]interface{}:
		_, err = o.client.HSet(ctx, key, value).Result()
		if err != nil {
			return fmt.Errorf("set hash: %w", err)
		}
		_, err = o.client.Expire(ctx, key, expiration).Result()
		if err != nil {
			return fmt.Errorf("set hash ttl: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("set key to redis: %w", err)
	}

	return err
}

func (o *redisStorage) Delete(ctx context.Context, key string) error {
	return o.client.Del(ctx, key).Err()
}

func (o *redisStorage) Close() error {
	return o.client.Close()
}
