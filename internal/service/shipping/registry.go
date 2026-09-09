package shipping

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/goccy/go-json"
	"golang.org/x/sync/errgroup"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/key_value"
)

// Registry is the set of shipping providers this store wires up, indexed by
// code, plus the checkout delivery-method catalogue. It also owns the
// short-lived cache of carrier rate answers.
type Registry struct {
	all       []Provider
	byCode    map[string]Provider
	methods   []Method
	methodIdx map[string]Method
	kv        key_value.IKeyValue
	quoteTTL  time.Duration
	defaults  ParcelDefaults
}

// NewRegistry indexes providers by their Code and builds the delivery-method
// catalogue. kv + quoteTTL back the shipping rate cache (carriers rate-limit
// calculation calls); defaults is the fallback parcel for products with no
// weight/dimensions.
func NewRegistry(
	kv key_value.IKeyValue,
	quoteTTL time.Duration,
	defaults ParcelDefaults,
	methods []config.ShippingMethodConfig,
	providers ...Provider,
) *Registry {
	byCode := make(map[string]Provider, len(providers))
	for _, p := range providers {
		byCode[p.Code()] = p
	}

	ms := methodsFromConfig(methods)
	methodIdx := make(map[string]Method, len(ms))
	for _, m := range ms {
		methodIdx[m.Code] = m
	}

	return &Registry{
		all: providers, byCode: byCode,
		methods: ms, methodIdx: methodIdx,
		kv: kv, quoteTTL: quoteTTL, defaults: defaults,
	}
}

// ParcelDefaults returns the store's fallback parcel dimensions.
func (r *Registry) ParcelDefaults() ParcelDefaults {
	return r.defaults
}

// Method returns the catalogue entry for code.
func (r *Registry) Method(code string) (Method, bool) {
	m, ok := r.methodIdx[code]
	return m, ok
}

// Methods returns the delivery-method catalogue with per-method availability
// flags resolved.
func (r *Registry) Methods() []MethodInfo {
	out := make([]MethodInfo, 0, len(r.methods))
	for _, m := range r.methods {
		out = append(out, r.methodInfo(m))
	}
	return out
}

func (r *Registry) methodInfo(m Method) MethodInfo {
	mi := MethodInfo{Method: m, Enabled: true}
	if m.Kind == MethodSelfPickup || m.Provider == "" {
		return mi
	}

	p, ok := r.byCode[m.Provider]
	mi.Enabled = ok && p.Enabled()
	mi.HasPoints = mi.Enabled && m.Kind == MethodPickup
	if _, isRater := r.Rater(m.Provider); isRater {
		mi.HasRates = true
	}
	return mi
}

// All returns every registered provider, in registration order.
func (r *Registry) All() []Provider {
	return r.all
}

// Get returns the provider registered under code.
func (r *Registry) Get(code string) (Provider, bool) {
	p, ok := r.byCode[code]
	return p, ok
}

// Rater returns the provider under code if it is enabled and supports rate
// calculation.
func (r *Registry) Rater(code string) (RateProvider, bool) {
	p, ok := r.byCode[code]
	if !ok || !p.Enabled() {
		return nil, false
	}
	rp, ok := p.(RateProvider)
	return rp, ok
}

// Quote returns one carrier's shipping options, from cache when possible.
func (r *Registry) Quote(ctx context.Context, code string, q RateQuery) ([]dto.ShippingRate, error) {
	rp, ok := r.Rater(code)
	if !ok {
		return nil, ErrRatesNotSupported
	}

	key := quoteCacheKey(code, q)
	if rates, ok := r.loadCached(ctx, key); ok {
		return rates, nil
	}

	rates, err := rp.Quote(ctx, q)
	if err != nil {
		return nil, err
	}
	sortRates(rates)
	r.storeCached(ctx, key, rates)
	return rates, nil
}

// QuoteAll asks every enabled rate provider concurrently and merges the
// options, cheapest first. Carriers without a calculator (ErrRatesNotSupported)
// are skipped; other per-carrier errors are collected and returned alongside
// whatever rates did come back.
func (r *Registry) QuoteAll(ctx context.Context, q RateQuery) ([]dto.ShippingRate, []error) {
	type result struct {
		rates []dto.ShippingRate
		err   error
	}
	results := make([]result, len(r.all))

	g, gctx := errgroup.WithContext(ctx)
	for i, p := range r.all {
		i, code := i, p.Code()
		if _, ok := r.Rater(code); !ok {
			continue
		}
		g.Go(func() error {
			rates, err := r.Quote(gctx, code, q)
			results[i] = result{rates: rates, err: err}
			return nil
		})
	}
	_ = g.Wait()

	var all []dto.ShippingRate
	var errs []error
	for _, res := range results {
		if res.err != nil {
			if !errors.Is(res.err, ErrRatesNotSupported) {
				errs = append(errs, res.err)
			}
			continue
		}
		all = append(all, res.rates...)
	}
	sortRates(all)
	return all, errs
}

func sortRates(rates []dto.ShippingRate) {
	sort.SliceStable(rates, func(i, j int) bool {
		return rates[i].Cost.LessThan(rates[j].Cost)
	})
}

func quoteCacheKey(code string, q RateQuery) string {
	dest := q.ToPointCode
	if dest == "" {
		dest = q.ToPostalCode
	}
	// bucket weight to 500 g so nearby carts reuse one answer
	weightBucket := (q.Parcel.WeightGrams + 499) / 500 * 500
	return fmt.Sprintf("shipping:rate:%s:%s:%s:%d:%s",
		code, q.FromPostalCode, dest, weightBucket, q.DeliveryType)
}

func (r *Registry) loadCached(ctx context.Context, key string) ([]dto.ShippingRate, bool) {
	if r.kv == nil {
		return nil, false
	}
	data, err := r.kv.Get(ctx, key)
	if err != nil {
		return nil, false
	}
	var rates []dto.ShippingRate
	if err := json.Unmarshal(data.Bytes(), &rates); err != nil {
		return nil, false
	}
	return rates, true
}

func (r *Registry) storeCached(ctx context.Context, key string, rates []dto.ShippingRate) {
	if r.kv == nil || len(rates) == 0 {
		return
	}
	data, err := json.Marshal(rates)
	if err != nil {
		return
	}
	_ = r.kv.Set(ctx, key, string(data), r.quoteTTL)
}
