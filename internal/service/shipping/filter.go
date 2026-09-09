package shipping

import (
	"math"
	"sort"
	"strings"

	"github.com/stickpro/go-store/internal/dto"
)

// defaultSearchRadiusKM is the nearby-search radius used when a request gives
// coordinates but no radius.
const defaultSearchRadiusKM = 20

// Filter applies a carrier-agnostic filter to a provider's points. It is the
// single place delivery-point filtering happens.
func Filter(points []dto.DeliveryPoint, f dto.DeliveryPointsFilter) []dto.DeliveryPoint {
	nearby := f.BBox == nil && f.Latitude != nil && f.Longitude != nil
	radiusKM := float64(defaultSearchRadiusKM)
	if f.RadiusKM != nil && *f.RadiusKM > 0 {
		radiusKM = *f.RadiusKM
	}

	result := make([]dto.DeliveryPoint, 0, len(points))
	for _, p := range points {
		if f.Type != "" && !strings.EqualFold(p.Type, f.Type) {
			continue
		}
		if f.Index != "" && p.Code != f.Index && p.PostalCode != f.Index {
			continue
		}
		if f.Region != "" && !containsFold(p.Region, f.Region) {
			continue
		}
		if f.Locality != "" && !containsFold(p.Locality, f.Locality) {
			continue
		}
		if f.BBox != nil && !f.BBox.Contains(p.Latitude, p.Longitude) {
			continue
		}
		if nearby && haversineKM(*f.Latitude, *f.Longitude, p.Latitude, p.Longitude) > radiusKM {
			continue
		}
		result = append(result, p)
	}

	if nearby {
		lat, lon := *f.Latitude, *f.Longitude
		sort.Slice(result, func(i, j int) bool {
			return haversineKM(lat, lon, result[i].Latitude, result[i].Longitude) <
				haversineKM(lat, lon, result[j].Latitude, result[j].Longitude)
		})
	}

	return result
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// haversineKM is the great-circle distance between two lat/lon points in km.
func haversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	const earthKM = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return earthKM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
