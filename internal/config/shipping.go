package config

import "time"

// ShippingConfig holds settings shared by every carrier's shipping-cost
// calculation: the origin the store ships from, the fallback parcel used for
// products with no weight/dimensions, and the delivery methods offered at
// checkout. The Default* amounts are decimal strings (weight in kilograms,
// dimensions in centimetres — matching the products table) parsed once at
// service init.
type ShippingConfig struct {
	// OriginPostalCode is the postal index the store ships parcels from; used as
	// the "from" of every rate request that doesn't set one explicitly.
	OriginPostalCode string `yaml:"origin_postal_code" usage:"postal index the store ships from (rate calc origin)"`

	DefaultWeightKG string `yaml:"default_weight_kg" default:"0.5" usage:"fallback parcel weight in kg for products without a weight"`
	DefaultLengthCM string `yaml:"default_length_cm" default:"20" usage:"fallback parcel length in cm"`
	DefaultWidthCM  string `yaml:"default_width_cm" default:"15" usage:"fallback parcel width in cm"`
	DefaultHeightCM string `yaml:"default_height_cm" default:"10" usage:"fallback parcel height in cm"`

	// QuoteCacheTTL is how long a carrier's rate answer is cached in the
	// key/value store (keyed by carrier + route + weight bucket + type). Carriers
	// rate-limit calculation calls, so keep this generous.
	QuoteCacheTTL time.Duration `yaml:"quote_cache_ttl" default:"6h" usage:"TTL of a cached shipping-rate answer"`

	// Methods is the checkout delivery-method catalogue served by
	// GET /v1/delivery/methods. Each maps a UI choice to a carrier + tariff so
	// the frontend never touches tariff codes. Empty falls back to
	// DefaultShippingMethods.
	Methods []ShippingMethodConfig `yaml:"methods"`
}

// ShippingMethodConfig is one checkout delivery method.
type ShippingMethodConfig struct {
	// Code is the stable id the frontend and checkout use (delivery_method_code).
	Code string `yaml:"code"`
	// Title is the human label shown on the method tab.
	Title string `yaml:"title"`
	// Kind: "self_pickup" (from the store), "pickup" (a carrier pickup point) or
	// "courier" (a carrier delivers to an address).
	Kind string `yaml:"kind"`
	// Provider is the carrier code ("cdek", "pochta", …); empty for self_pickup.
	Provider string `yaml:"provider"`
	// TariffCode is the carrier tariff this method quotes; empty for self_pickup
	// or for carriers with no rate calculation.
	Tariff string `yaml:"tariff"`
	// Free forces a zero shipping cost regardless of the carrier quote.
	Free bool `yaml:"free"`
}

// DefaultShippingMethods is used when shipping.methods is empty.
func DefaultShippingMethods() []ShippingMethodConfig {
	return []ShippingMethodConfig{
		{Code: "pickup", Title: "Самовывоз", Kind: "self_pickup", Free: true},
		{Code: "cdek", Title: "СДЭК", Kind: "pickup", Provider: "cdek", Tariff: "136"},
		{Code: "pochta", Title: "Почта России", Kind: "pickup", Provider: "pochta", Tariff: "27030"},
		{Code: "yandex_delivery", Title: "Яндекс Доставка", Kind: "pickup", Provider: "yandex_delivery"},
	}
}

// ResolvedMethods returns the configured methods, or the built-in defaults when
// none are set.
func (c ShippingConfig) ResolvedMethods() []ShippingMethodConfig {
	if len(c.Methods) == 0 {
		return DefaultShippingMethods()
	}
	return c.Methods
}
