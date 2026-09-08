package dto

// GeoBBox narrows a list of geo-located entities to a rectangular geographic
// area (typically a map viewport), inclusive on all bounds. It does not handle
// boxes that cross the antimeridian.
type GeoBBox struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// Contains reports whether (lat, lon) falls inside b, bounds inclusive.
func (b GeoBBox) Contains(lat, lon float64) bool {
	return lat >= b.MinLat && lat <= b.MaxLat &&
		lon >= b.MinLon && lon <= b.MaxLon
}
