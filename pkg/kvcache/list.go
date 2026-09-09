// Package kvcache provides an in-memory list that is periodically refreshed
// from a source function and mirrored into a key/value store, so it survives a
// process restart with a warm cache. It is the shared engine behind the CDEK,
// Yandex Delivery and Russian Post pickup-points caches.
package kvcache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

// Config parametrises a List.
type Config struct {
	// Name prefixes log lines, e.g. "cdek".
	Name string
	// Key is the key/value store key that holds the persisted copy.
	Key string
	// TTL is how long the persisted copy stays valid; keep it comfortably longer
	// than RefreshEvery so a slow or failed refresh doesn't drop the warm cache.
	TTL time.Duration
	// RefreshEvery is the background refresh interval (defaults to 24h when <= 0).
	RefreshEvery time.Duration
	// WarmupBackoff is the first back-off between initial-refresh retries on
	// startup; it doubles up to a 5m cap over initialRefreshAttempts tries
	// (defaults to 30s when <= 0).
	WarmupBackoff time.Duration
}

// List holds []T in process memory, refreshes it from fetch on an interval and
// persists a copy to kv for a warm start. Create it with New; the zero value is
// not usable. Safe for concurrent use.
type List[T any] struct {
	cfg    Config
	kv     key_value.IKeyValue
	logger logger.Logger
	fetch  func(context.Context) ([]T, error)

	mu     sync.RWMutex
	items  []T
	loaded bool
}

// New builds a List. fetch loads the authoritative list from its source (an API,
// a file, …); it is called by Refresh and by Run.
func New[T any](cfg Config, kv key_value.IKeyValue, l logger.Logger, fetch func(context.Context) ([]T, error)) *List[T] {
	return &List[T]{cfg: cfg, kv: kv, logger: l, fetch: fetch}
}

// Items returns the current in-memory snapshot. The slice is swapped wholesale
// on every refresh and never mutated in place, so callers may read it directly.
func (l *List[T]) Items() []T {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.items
}

// Loaded reports whether the list has been populated at least once.
func (l *List[T]) Loaded() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loaded
}

// Seed replaces the in-memory list without fetching or persisting. Handy for
// tests and for priming the cache from a known-good source.
func (l *List[T]) Seed(items []T) {
	l.store(items)
}

// Refresh fetches a fresh list, swaps it into memory and persists a copy. A
// failed persist is logged, not returned: memory is the source of truth.
func (l *List[T]) Refresh(ctx context.Context) error {
	items, err := l.fetch(ctx)
	if err != nil {
		return fmt.Errorf("%s: refresh: %w", l.cfg.Name, err)
	}

	l.store(items)

	if err := l.persist(ctx, items); err != nil {
		l.logger.Errorw(l.cfg.Name+": persist cache", "error", err)
	}
	l.logger.Infow(l.cfg.Name+": list refreshed", "count", len(items))
	return nil
}

// initial-refresh retry schedule: a transient error on startup shouldn't leave
// the cache cold until the next scheduled refresh (which can be a day away).
const (
	initialRefreshAttempts   = 5
	defaultWarmupBackoff     = 30 * time.Second
	initialRefreshBackoffMax = 5 * time.Minute
)

// Run blocks: it warms memory from the persisted copy (or a fetch, with a few
// backoff retries), then refreshes on the configured interval until ctx is
// cancelled.
func (l *List[T]) Run(ctx context.Context) {
	l.warmUp(ctx)

	interval := l.cfg.RefreshEvery
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := l.Refresh(ctx); err != nil {
				l.logger.Errorw(l.cfg.Name+": refresh failed", "error", err)
			}
		}
	}
}

// warmUp populates memory once: from the persisted copy if there is one, else
// from a fetch retried with exponential backoff.
func (l *List[T]) warmUp(ctx context.Context) {
	if items, err := l.loadPersisted(ctx); err != nil {
		l.logger.Errorw(l.cfg.Name+": read persisted cache", "error", err)
	} else if items != nil {
		l.store(items)
		l.logger.Infow(l.cfg.Name+": list loaded from cache", "count", len(items))
		return
	}

	backoff := l.cfg.WarmupBackoff
	if backoff <= 0 {
		backoff = defaultWarmupBackoff
	}
	for attempt := 1; ; attempt++ {
		if err := l.Refresh(ctx); err == nil {
			return
		} else {
			l.logger.Errorw(l.cfg.Name+": initial refresh failed", "attempt", attempt, "error", err)
		}
		if attempt >= initialRefreshAttempts || ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > initialRefreshBackoffMax {
			backoff = initialRefreshBackoffMax
		}
	}
}

func (l *List[T]) store(items []T) {
	l.mu.Lock()
	l.items = items
	l.loaded = true
	l.mu.Unlock()
}

func (l *List[T]) persist(ctx context.Context, items []T) error {
	data, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := l.kv.Set(ctx, l.cfg.Key, string(data), l.cfg.TTL); err != nil {
		return fmt.Errorf("kv set: %w", err)
	}
	return nil
}

// loadPersisted returns nil (no error) when there is no persisted copy yet.
func (l *List[T]) loadPersisted(ctx context.Context) ([]T, error) {
	data, err := l.kv.Get(ctx, l.cfg.Key)
	if err != nil {
		if errors.Is(err, key_value.ErrEntryNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var items []T
	if err := json.Unmarshal(data.Bytes(), &items); err != nil {
		return nil, fmt.Errorf("unmarshal persisted: %w", err)
	}
	return items, nil
}
