package cdek

import (
	"context"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

// provider is the CDEK (СДЭК) carrier: pickup points via the shared cache
// engine (embedded *shipping.CachedProvider) plus shipping-cost calculation
// (see rater.go).
type provider struct {
	*shipping.CachedProvider
	client   *client
	cfg      config.CDEKConfig
	defaults shipping.ParcelDefaults
	origin   string
}

// New builds the CDEK delivery provider.
func New(cfg *config.Config, l logger.Logger, kv key_value.IKeyValue) shipping.Provider {
	c := newClient(cfg.CDEK, l)
	return &provider{
		CachedProvider: shipping.NewCachedProvider(shipping.ProviderConfig{
			Code:         "cdek",
			Enabled:      cfg.CDEK.Enabled,
			CacheKey:     constant.CacheKeyCDEKDeliveryPoints,
			CacheTTL:     cfg.CDEK.DeliveryPointsCacheTTL,
			RefreshEvery: cfg.CDEK.DeliveryPointsRefreshEvery,
		}, kv, l, func(ctx context.Context) ([]dto.CDEKDeliveryPointDTO, error) {
			return c.searchAllDeliveryPoints(ctx, cfg.CDEK.DeliveryPointsCountryCode)
		}, toDeliveryPoint),
		client:   c,
		cfg:      cfg.CDEK,
		defaults: shipping.ParcelDefaultsFromConfig(cfg.Shipping),
		origin:   cfg.Shipping.OriginPostalCode,
	}
}

func toDeliveryPoint(p dto.CDEKDeliveryPointDTO) dto.DeliveryPoint {
	var workTime []string
	if p.WorkTime != "" {
		workTime = []string{p.WorkTime}
	}

	return dto.DeliveryPoint{
		Provider:    "cdek",
		Code:        p.Code,
		Name:        p.Name,
		Type:        p.Type,
		PostalCode:  p.PostalCode,
		Country:     p.CountryCode,
		Region:      p.Region,
		Locality:    p.City,
		Address:     firstNonEmpty(p.AddressFull, p.Address),
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Phones:      p.Phones,
		Email:       p.Email,
		WorkTime:    workTime,
		CardPayment: p.HaveCashless,
		CashPayment: p.HaveCash,
		Details: map[string]any{
			"owner_code":       p.OwnerCode,
			"region_code":      p.RegionCode,
			"city_code":        p.CityCode,
			"note":             p.Note,
			"take_only":        p.TakeOnly,
			"is_handout":       p.IsHandout,
			"is_reception":     p.IsReception,
			"is_dressing_room": p.IsDressingRoom,
			"allowed_cod":      p.AllowedCod,
			"weight_min":       p.WeightMin,
			"weight_max":       p.WeightMax,
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
