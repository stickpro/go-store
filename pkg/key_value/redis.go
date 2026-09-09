package key_value

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
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
	kType, err := o.client.Type(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get key type: %w", err)
	}
	switch kType {
	case "none":
		return nil, ErrEntryNotFound
	case "hash":
		res, err := o.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("get hash key from redis: %w", err)
		}
		return json.Marshal(res)
	case "string":
		b, err := o.client.Get(ctx, key).Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil, ErrEntryNotFound
			}
			return nil, fmt.Errorf("get key from redis: %w", err)
		}
		return b, nil
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

func (o *redisStorage) Incr(ctx context.Context, key string) (int64, error) {
	n, err := o.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("incr key in redis: %w", err)
	}
	return n, nil
}

func (o *redisStorage) SetNX(ctx context.Context, key, value string, expiration time.Duration) (bool, error) {
	ok, err := o.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("setnx key in redis: %w", err)
	}
	return ok, nil
}

func (o *redisStorage) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := o.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("expire key in redis: %w", err)
	}
	return nil
}

func (o *redisStorage) Delete(ctx context.Context, key string) error {
	return o.client.Del(ctx, key).Err()
}

func (o *redisStorage) Close() error {
	return o.client.Close()
}
