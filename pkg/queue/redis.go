package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisQueue struct {
	client *redis.Client
}

func NewRedisQueue(url string) (IQueue, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("queue: parse redis url: %w", err)
	}
	return &redisQueue{client: redis.NewClient(opts)}, nil
}

func (q *redisQueue) Push(ctx context.Context, name string, payload []byte) error {
	if err := q.client.LPush(ctx, name, payload).Err(); err != nil {
		return fmt.Errorf("queue push %s: %w", name, err)
	}
	return nil
}

func (q *redisQueue) Pop(ctx context.Context, name string) ([]byte, error) {
	result, err := q.client.BRPop(ctx, time.Second, name).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("queue pop %s: %w", name, err)
	}
	return []byte(result[1]), nil
}

func (q *redisQueue) Close() error {
	return q.client.Close()
}
