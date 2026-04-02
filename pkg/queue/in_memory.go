package queue

import (
	"context"
	"sync"
)

type inMemoryQueue struct {
	mu     sync.Mutex
	queues map[string]chan []byte
}

func NewInMemoryQueue() IQueue {
	return &inMemoryQueue{queues: make(map[string]chan []byte)}
}

func (q *inMemoryQueue) ch(name string) chan []byte {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.queues[name]; !ok {
		q.queues[name] = make(chan []byte, 10000)
	}
	return q.queues[name]
}

func (q *inMemoryQueue) Push(_ context.Context, name string, payload []byte) error {
	q.ch(name) <- payload
	return nil
}

func (q *inMemoryQueue) Pop(ctx context.Context, name string) ([]byte, error) {
	select {
	case v := <-q.ch(name):
		return v, nil
	case <-ctx.Done():
		return nil, nil
	}
}

func (q *inMemoryQueue) Close() error { return nil }
