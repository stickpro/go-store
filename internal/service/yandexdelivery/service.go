package yandexdelivery

import (
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

// New builds the Yandex Delivery (Яндекс Доставка) provider. It serves pickup
// points via the shared cache engine but is not a shipping.RateProvider —
// Yandex cost calculation (Platform offers/create) needs a platform_station_id,
// a warehouse registered in the Yandex cabinet, which is not configured. So
// /v1/delivery/yandex_delivery/rates returns 404 and QuoteAll skips it.
func New(cfg *config.Config, l logger.Logger, kv key_value.IKeyValue) shipping.Provider {
	c := newClient(cfg.YandexDelivery, l)
	return shipping.NewCachedProvider(shipping.ProviderConfig{
		Code:         "yandex_delivery",
		Enabled:      cfg.YandexDelivery.Enabled,
		CacheKey:     constant.CacheKeyYandexDeliveryPoints,
		CacheTTL:     cfg.YandexDelivery.DeliveryPointsCacheTTL,
		RefreshEvery: cfg.YandexDelivery.DeliveryPointsRefreshEvery,
	}, kv, l, c.listAllDeliveryPoints, toDeliveryPoint)
}

func toDeliveryPoint(p dto.YandexDeliveryPointDTO) dto.DeliveryPoint {
	var phones []string
	if p.Phone != "" {
		phones = []string{p.Phone}
	}

	hasCard := false
	for _, m := range p.PaymentMethods {
		if m == "card_on_receipt" || m == "card" {
			hasCard = true
		}
	}

	return dto.DeliveryPoint{
		Provider:    "yandex_delivery",
		Code:        p.Code,
		Name:        p.Name,
		Type:        p.Type,
		PostalCode:  p.PostalCode,
		Country:     p.Country,
		Region:      p.Region,
		Locality:    p.Locality,
		Address:     p.FullAddress,
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Phones:      phones,
		Email:       p.Email,
		CardPayment: hasCard,
		Details: map[string]any{
			"operator_station_id":   p.OperatorStationID,
			"operator_id":           p.OperatorID,
			"sub_region":            p.SubRegion,
			"street":                p.Street,
			"house":                 p.House,
			"geo_id":                p.GeoID,
			"instruction":           p.Instruction,
			"payment_methods":       p.PaymentMethods,
			"time_zone":             p.TimeZone,
			"schedule":              p.Schedule,
			"is_yandex_branded":     p.IsYandexBranded,
			"is_market_partner":     p.IsMarketPartner,
			"is_post_office":        p.IsPostOffice,
			"available_for_dropoff": p.AvailableForDropoff,
			"deactivation_date":     p.DeactivationDate,
		},
	}
}
