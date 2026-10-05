package constant

// Cache keys shared between services (key_value store).
const (
	// CacheKeyFilterableAttributes caches the list of filterable attributes with their
	// metadata. Invalidated by the attribute service on any attribute mutation.
	CacheKeyFilterableAttributes = "attributes:filterable"

	// CacheKeyProductVariantsFilterableAttributesPushed holds a fingerprint of
	// the filterable attribute slugs last pushed to the product_variants
	// index's settings (see attribute.Service.RefreshFilterableAttributes).
	// Meili recomputes facets over the whole collection on a settings update,
	// so this guards against pushing on every attribute write when the set
	// hasn't actually changed.
	CacheKeyProductVariantsFilterableAttributesPushed = "product_variants:filterable_attributes_pushed"

	// CacheKeyCDEKDeliveryPoints caches the full CDEK delivery points (offices +
	// postomats) list. Repopulated once a day by the CDEK cache refresher.
	CacheKeyCDEKDeliveryPoints = "cdek:delivery_points"

	// CacheKeyYandexDeliveryPoints holds a persisted copy of the full Yandex
	// Delivery pickup points (ПВЗ + postomats) list, used only to warm-start the
	// in-memory list after a restart. The read path is process memory, not this
	// key. Repopulated once a day by the Yandex Delivery cache refresher.
	CacheKeyYandexDeliveryPoints = "yandex_delivery:delivery_points"

	// CacheKeyPochtaDeliveryPoints holds a persisted copy of the full Russian
	// Post pickup points (ОПС + ПВЗ + postomats) list, exported from the "ОПС
	// passport" endpoint. Used only to warm-start the in-memory list after a
	// restart; the read path is process memory, not this key. Repopulated once a
	// day by the Russian Post cache refresher.
	CacheKeyPochtaDeliveryPoints = "pochta:delivery_points"
)
