package shipping

import (
	"context"
	"testing"
	"time"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

func newTestProvider(t *testing.T, enabled bool) *CachedProvider {
	t.Helper()
	return NewCachedProvider(
		ProviderConfig{Code: "test", Enabled: enabled, CacheKey: "test:dp", CacheTTL: time.Minute},
		key_value.NewInMemory(), logger.New(),
		func(context.Context) ([]int, error) { return []int{1, 2}, nil },
		func(n int) dto.DeliveryPoint { return dto.DeliveryPoint{Code: string(rune('a' + n))} },
	)
}

func TestCachedProviderPointsRespectsEnabled(t *testing.T) {
	off := newTestProvider(t, false)
	off.cache.Seed([]dto.DeliveryPoint{{Code: "x"}})
	if got := off.Points(); got != nil {
		t.Fatalf("disabled provider must return nil, got %v", got)
	}

	on := newTestProvider(t, true)
	on.cache.Seed([]dto.DeliveryPoint{{Code: "x"}})
	if got := on.Points(); len(got) != 1 {
		t.Fatalf("enabled provider must return seeded points, got %v", got)
	}
}

func TestCachedProviderMapsFetchedRows(t *testing.T) {
	p := newTestProvider(t, true)
	if err := p.cache.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := p.Points()
	if len(got) != 2 || got[0].Code != "b" || got[1].Code != "c" {
		t.Fatalf("fetch+map: got %+v", got)
	}
}

func TestCachedProviderFindPoint(t *testing.T) {
	p := newTestProvider(t, true)
	p.cache.Seed([]dto.DeliveryPoint{{Code: "p1", PostalCode: "115551"}})
	if pt, ok := p.FindPoint("p1"); !ok || pt.PostalCode != "115551" {
		t.Fatalf("FindPoint(p1): %+v ok=%v", pt, ok)
	}
	if _, ok := p.FindPoint("nope"); ok {
		t.Fatal("FindPoint(nope) should miss")
	}
}
