package pochta

import (
	"strings"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

// provider is the Russian Post (Почта России) carrier: pickup points via the
// shared cache engine (embedded *shipping.CachedProvider) plus shipping-cost
// calculation through the public tariff calculator (see rater.go).
type provider struct {
	*shipping.CachedProvider
	client *client
	cfg    config.PochtaConfig
	origin string
}

// New builds the Russian Post delivery provider.
func New(cfg *config.Config, l logger.Logger, kv key_value.IKeyValue) shipping.Provider {
	c := newClient(cfg.Pochta, l)
	return &provider{
		CachedProvider: shipping.NewCachedProvider(shipping.ProviderConfig{
			Code:         "pochta",
			Enabled:      cfg.Pochta.Enabled,
			CacheKey:     constant.CacheKeyPochtaDeliveryPoints,
			CacheTTL:     cfg.Pochta.DeliveryPointsCacheTTL,
			RefreshEvery: cfg.Pochta.DeliveryPointsRefreshEvery,
		}, kv, l, c.listAllDeliveryPoints, toDeliveryPoint),
		client: c,
		cfg:    cfg.Pochta,
		origin: cfg.Shipping.OriginPostalCode,
	}
}

func toDeliveryPoint(p dto.PochtaDeliveryPointDTO) dto.DeliveryPoint {
	return dto.DeliveryPoint{
		Provider:    "pochta",
		Code:        p.Code,
		Name:        p.Name,
		Type:        p.Type,
		PostalCode:  p.PostalCode,
		Country:     "RU",
		Region:      p.Region,
		Locality:    p.Place,
		Address:     firstNonEmpty(p.Address, strings.TrimSpace(p.Street+" "+p.House)),
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		WorkTime:    p.WorkTime,
		CardPayment: p.CardPayment,
		CashPayment: p.CashPayment,
		Details: map[string]any{
			"brand_name":         p.BrandName,
			"area":               p.Area,
			"office":             p.Office,
			"description":        p.Description,
			"ecom":               p.Ecom,
			"weight_limit_kg":    p.WeightLimitKg,
			"with_fitting":       p.WithFitting,
			"contents_checking":  p.ContentsChecking,
			"partial_redemption": p.PartialRedemption,
			"return_available":   p.ReturnAvailable,
		},
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
