package kvcache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

func newList(t *testing.T, kv key_value.IKeyValue, fetch func(context.Context) ([]int, error)) *List[int] {
	t.Helper()
	return New(Config{Name: "test", Key: "test:list", TTL: time.Minute, RefreshEvery: time.Hour}, kv, logger.New(), fetch)
}

func TestRefreshStoresAndPersists(t *testing.T) {
	kv := key_value.NewInMemory()
	calls := 0
	l := newList(t, kv, func(context.Context) ([]int, error) {
		calls++
		return []int{1, 2, 3}, nil
	})

	if err := l.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !l.Loaded() || len(l.Items()) != 3 {
		t.Fatalf("after refresh: calls=%d loaded=%v items=%v", calls, l.Loaded(), l.Items())
	}

	// A fresh List over the same kv warm-starts from the persisted copy without
	// calling fetch.
	l2 := newList(t, kv, func(context.Context) ([]int, error) {
		t.Fatal("fetch must not run when a persisted copy exists and is non-empty")
		return nil, nil
	})
	got, err := l2.loadPersisted(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("persisted copy: got %v, want 3 items", got)
	}
}

func TestRefreshFetchErrorKeepsMemory(t *testing.T) {
	l := newList(t, key_value.NewInMemory(), func(context.Context) ([]int, error) {
		return nil, errors.New("boom")
	})
	l.Seed([]int{9})

	if err := l.Refresh(context.Background()); err == nil {
		t.Fatal("expected error from failing fetch")
	}
	if len(l.Items()) != 1 || l.Items()[0] != 9 {
		t.Fatalf("memory must be untouched on fetch error, got %v", l.Items())
	}
}

func TestWarmUpRetriesUntilFetchSucceeds(t *testing.T) {
	calls := 0
	l := New(Config{Name: "test", Key: "test:list", WarmupBackoff: time.Millisecond}, key_value.NewInMemory(), logger.New(),
		func(context.Context) ([]int, error) {
			calls++
			if calls < 3 {
				return nil, errors.New("transient")
			}
			return []int{1, 2}, nil
		})

	l.warmUp(context.Background())

	if calls != 3 || !l.Loaded() || len(l.Items()) != 2 {
		t.Fatalf("warmUp: calls=%d loaded=%v items=%v", calls, l.Loaded(), l.Items())
	}
}

func TestWarmUpUsesPersistedWithoutFetch(t *testing.T) {
	kv := key_value.NewInMemory()
	seed := New(Config{Name: "test", Key: "test:list"}, kv, logger.New(),
		func(context.Context) ([]int, error) { return []int{7, 8, 9}, nil })
	if err := seed.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}

	l := New(Config{Name: "test", Key: "test:list"}, kv, logger.New(),
		func(context.Context) ([]int, error) {
			t.Fatal("fetch must not run when a persisted copy exists")
			return nil, nil
		})
	l.warmUp(context.Background())
	if len(l.Items()) != 3 {
		t.Fatalf("warmUp should load persisted copy, got %v", l.Items())
	}
}

func TestLoadPersistedMissing(t *testing.T) {
	l := newList(t, key_value.NewInMemory(), nil)
	got, err := l.loadPersisted(context.Background())
	if err != nil {
		t.Fatalf("missing key must not error, got %v", err)
	}
	if got != nil {
		t.Fatalf("missing key must return nil, got %v", got)
	}
}
