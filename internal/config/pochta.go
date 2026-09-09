package config

import "time"

// PochtaConfig configures the Russian Post (Почта России) pickup points
// integration. The full directory of post offices, ПВЗ and postomats is
// exported once from the Otpravka "ОПС passport" endpoint
// (https://otpravka-api.pochta.ru/1.0/unloading-passport/zip), kept in process
// memory and refreshed in the background — exactly like the CDEK and Yandex
// Delivery integrations. Callers never hit the Russian Post API per request.
//
// Two credentials are needed, both issued in the otpravka.pochta.ru cabinet:
//   - AccessToken — the application token (sent as "Authorization: AccessToken …")
//   - UserLogin / UserPassword — the pochta.ru account, sent as
//     "X-User-Authorization: Basic base64(login:password)"
//
// When Enabled is false the list is never populated and the Pochta endpoints
// return an empty list.
type PochtaConfig struct {
	Enabled        bool          `yaml:"enabled" default:"false" usage:"enable the Russian Post (Почта России) pickup points integration"`
	BaseURL        string        `yaml:"base_url" default:"https://otpravka-api.pochta.ru" usage:"Russian Post Otpravka API base URL"`
	AccessToken    string        `yaml:"access_token" secret:"true" usage:"Otpravka API application token (Authorization: AccessToken …) from the otpravka.pochta.ru cabinet"`
	UserLogin      string        `yaml:"user_login" usage:"pochta.ru account login (email) for the X-User-Authorization header"`
	UserPassword   string        `yaml:"user_password" secret:"true" usage:"pochta.ru account password for the X-User-Authorization header"`
	RequestTimeout time.Duration `yaml:"request_timeout" default:"180s" usage:"HTTP timeout for the ОПС passport export; the archive unpacks to ~60 MB, so keep it generous"`

	// UnloadType selects which objects the ОПС passport export returns:
	// "ALL", "OPS" (post offices), "PVZ" (pickup points) or "APS" (postomats).
	UnloadType string `yaml:"unload_type" default:"ALL" usage:"ОПС passport export object type: ALL | OPS | PVZ | APS"`

	// DeliveryPointsCacheTTL is how long the persisted copy of the list stays
	// valid in the key/value store; keep it comfortably longer than
	// DeliveryPointsRefreshEvery so a slow or failed refresh doesn't drop the cache.
	DeliveryPointsCacheTTL time.Duration `yaml:"delivery_points_cache_ttl" default:"72h" usage:"TTL of the persisted delivery points list"`
	// DeliveryPointsRefreshEvery is how often the background job re-exports the
	// full delivery points list from the ОПС passport and repopulates memory.
	DeliveryPointsRefreshEvery time.Duration `yaml:"delivery_points_refresh_every" default:"24h" usage:"how often the background job refreshes the delivery points list"`

	// TariffURL is the public Russian Post tariff calculator base URL. It needs
	// no authentication.
	TariffURL string `yaml:"tariff_url" default:"https://tariff.pochta.ru" usage:"Russian Post tariff calculator base URL (public, no auth)"`
	// TariffObjectCode is the "object of calculation" code passed to the tariff
	// calculator: 27030 (Посылка стандарт), 27020 (Посылка стандарт с
	// объявленной ценностью), 23030 (Посылка нестандартная), …
	TariffObjectCode int `yaml:"tariff_object_code" default:"27030" usage:"Russian Post tariff calculator object code (27030 = Посылка стандарт)"`
	// TariffPackType is the packaging code (Appendix 3): 10/20/30/40 = box
	// S/M/L/XL, 11/21/31 = poly bag S/M/L. Required by the "посылка" objects.
	TariffPackType int `yaml:"tariff_pack_type" default:"30" usage:"Russian Post packaging code (10/20/30/40 = box S/M/L/XL)"`
}
