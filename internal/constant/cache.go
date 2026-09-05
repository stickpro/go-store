package constant

// Cache keys shared between services (key_value store).
const (
	// CacheKeyFilterableAttributes caches the list of filterable attributes with their
	// metadata. Invalidated by the attribute service on any attribute mutation.
	CacheKeyFilterableAttributes = "attributes:filterable"

	// CacheKeyCDEKDeliveryPoints caches the full CDEK delivery points (offices +
	// postomats) list. Repopulated once a day by the CDEK cache refresher.
	CacheKeyCDEKDeliveryPoints = "cdek:delivery_points"
)
