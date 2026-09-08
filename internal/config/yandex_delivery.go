package config

import "time"

// YandexDeliveryConfig configures the Yandex Delivery (Яндекс Доставка) pickup
// points integration. OauthToken is the Yandex OAuth token issued for the
// logistics platform (b2b) API. When Enabled is false the delivery points cache
// is never populated and the Yandex Delivery endpoints return an empty list.
type YandexDeliveryConfig struct {
	Enabled        bool          `yaml:"enabled" default:"false" usage:"enable the Yandex Delivery pickup points integration"`
	BaseURL        string        `yaml:"base_url" default:"https://b2b-authproxy.taxi.yandex.net" usage:"Yandex Delivery API base URL; use https://b2b.taxi.tst.yandex.net for the test contour"`
	OauthToken     string        `yaml:"oauth_token" secret:"true" usage:"Yandex OAuth token for the logistics platform (b2b) API"`
	RequestTimeout time.Duration `yaml:"request_timeout" default:"180s" usage:"HTTP timeout for a single Yandex Delivery API request; the full pickup points list is tens of MB, so keep it generous"`

	// DeliveryPointsCacheTTL is how long the cached delivery points list stays
	// valid in the key/value store; keep it comfortably longer than
	// DeliveryPointsRefreshEvery so a slow or failed refresh doesn't drop the cache.
	DeliveryPointsCacheTTL time.Duration `yaml:"delivery_points_cache_ttl" default:"48h" usage:"TTL of the cached delivery points list"`
	// DeliveryPointsRefreshEvery is how often the background job re-fetches the
	// full delivery points list from Yandex Delivery and repopulates the cache.
	DeliveryPointsRefreshEvery time.Duration `yaml:"delivery_points_refresh_every" default:"24h" usage:"how often the background job refreshes the delivery points cache"`
}
