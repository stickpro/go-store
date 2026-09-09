package config

import "time"

// CDEKConfig configures the CDEK (СДЭК) delivery integration. Account and
// SecurePassword are the OAuth2 client_credentials pair issued in the CDEK
// partner cabinet (https://lk.cdek.ru). When Enabled is false the delivery
// points cache is never populated and CDEK endpoints return an empty list.
type CDEKConfig struct {
	Enabled        bool          `yaml:"enabled" default:"false" usage:"enable the CDEK delivery integration"`
	BaseURL        string        `yaml:"base_url" default:"https://api.cdek.ru/v2" usage:"CDEK API base URL; use https://api.edu.cdek.ru/v2 for the test contour"`
	Account        string        `yaml:"account" usage:"CDEK API client_id (account)"`
	SecurePassword string        `yaml:"secure_password" secret:"true" usage:"CDEK API client_secret (secure password)"`
	RequestTimeout time.Duration `yaml:"request_timeout" default:"15s" usage:"HTTP timeout for a single CDEK API request"`

	// DeliveryPointsCacheTTL is how long the cached delivery points list stays
	// valid in the key/value store; keep it comfortably longer than
	// DeliveryPointsRefreshEvery so a slow or failed refresh doesn't drop the cache.
	DeliveryPointsCacheTTL time.Duration `yaml:"delivery_points_cache_ttl" default:"48h" usage:"TTL of the cached delivery points list"`
	// DeliveryPointsRefreshEvery is how often the background job re-fetches the
	// full delivery points list from CDEK and repopulates the cache.
	DeliveryPointsRefreshEvery time.Duration `yaml:"delivery_points_refresh_every" default:"24h" usage:"how often the background job refreshes the delivery points cache"`
	// DeliveryPointsCountryCode limits the cached list to one country (ISO 3166-1
	// alpha-2, e.g. "RU"). Empty fetches every CDEK delivery point worldwide.
	DeliveryPointsCountryCode string `yaml:"delivery_points_country_code" default:"RU" usage:"ISO 3166-1 alpha-2 country code to limit cached delivery points to; empty = worldwide"`

	// TariffCodes are the CDEK tariff codes offered when calculating shipping
	// cost, e.g. 136 (warehouse-warehouse parcel), 137 (warehouse-door),
	// 139 (door-door), 366/368 (E-commerce express). Each is quoted separately.
	TariffCodes []int `yaml:"tariff_codes" usage:"CDEK tariff codes to quote for shipping cost (e.g. 136, 137, 139)"`
}
