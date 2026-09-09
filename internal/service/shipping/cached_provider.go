package shipping

import (
	"context"
	"time"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/kvcache"
	"github.com/stickpro/go-store/pkg/logger"
)

// ProviderConfig is the per-carrier wiring for NewCachedProvider.
type ProviderConfig struct {
	Code         string // carrier id, also the log prefix
	Enabled      bool
	CacheKey     string // key/value store key for the persisted copy
	CacheTTL     time.Duration
	RefreshEvery time.Duration
}

// NewCachedProvider builds the pickup-points half of a carrier: it keeps the
// carrier's full directory warm in memory (see pkg/kvcache). fetch loads the
// carrier's own rows, mapFn turns each into the common dto.DeliveryPoint. A
// carrier package embeds the returned *CachedProvider and adds Quote.
func NewCachedProvider[S any](
	pc ProviderConfig,
	kv key_value.IKeyValue,
	l logger.Logger,
	fetch func(context.Context) ([]S, error),
	mapFn func(S) dto.DeliveryPoint,
) *CachedProvider {
	p := &CachedProvider{code: pc.Code, enabled: pc.Enabled}
	p.cache = kvcache.New(kvcache.Config{
		Name:         pc.Code,
		Key:          pc.CacheKey,
		TTL:          pc.CacheTTL,
		RefreshEvery: pc.RefreshEvery,
	}, kv, l, func(ctx context.Context) ([]dto.DeliveryPoint, error) {
		src, err := fetch(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]dto.DeliveryPoint, len(src))
		for i, v := range src {
			out[i] = mapFn(v)
		}
		return out, nil
	})
	return p
}

// CachedProvider is the shared pickup-points implementation of Provider. Carrier
// packages embed it and add Quote to become a RateProvider.
type CachedProvider struct {
	code    string
	enabled bool
	cache   *kvcache.List[dto.DeliveryPoint]
}

func (p *CachedProvider) Code() string { return p.code }

func (p *CachedProvider) Enabled() bool { return p.enabled }

func (p *CachedProvider) Points() []dto.DeliveryPoint {
	if !p.enabled {
		return nil
	}
	return p.cache.Items()
}

func (p *CachedProvider) RunCacheRefresher(ctx context.Context) {
	if !p.enabled {
		return
	}
	p.cache.Run(ctx)
}

// FindPoint returns the cached pickup point with the given Code.
func (p *CachedProvider) FindPoint(code string) (dto.DeliveryPoint, bool) {
	for _, pt := range p.cache.Items() {
		if pt.Code == code {
			return pt, true
		}
	}
	return dto.DeliveryPoint{}, false
}
