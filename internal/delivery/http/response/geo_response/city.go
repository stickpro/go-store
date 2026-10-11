package geo_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// CityResponse is a settlement from the cities table: a city (city set,
// settlement null), a settlement (settlement set, city set when it lies inside
// a city) or a federal city (only region set). Fields GAR doesn't provide —
// coordinates, timezone, foundation year — are null for most settlements.
type CityResponse struct {
	ID              uuid.UUID        `json:"id"`
	Address         string           `json:"address"`
	PostalCode      *string          `json:"postal_code"`
	Country         string           `json:"country"`
	FederalDistrict string           `json:"federal_district"`
	RegionType      string           `json:"region_type"`
	Region          string           `json:"region"`
	AreaType        *string          `json:"area_type"`
	Area            *string          `json:"area"`
	CityType        *string          `json:"city_type"`
	City            *string          `json:"city"`
	SettlementType  *string          `json:"settlement_type"`
	Settlement      *string          `json:"settlement"`
	KladrID         *string          `json:"kladr_id"`
	FiasID          uuid.UUID        `json:"fias_id"`
	FiasLevel       int16            `json:"fias_level"`
	CapitalMarker   int16            `json:"capital_marker"`
	Okato           *string          `json:"okato"`
	Oktmo           *string          `json:"oktmo"`
	TaxOffice       *string          `json:"tax_office"`
	Timezone        *string          `json:"timezone"`
	GeoLat          *decimal.Decimal `json:"geo_lat"`
	GeoLon          *decimal.Decimal `json:"geo_lon"`
	Population      int64            `json:"population"`
	FoundationYear  *int16           `json:"foundation_year"`
} //	@name	CityResponse

func NewFromModel(city *models.City) *CityResponse {
	return &CityResponse{
		ID:              city.ID,
		Address:         city.Address,
		PostalCode:      pgtypeutils.DecodeText(city.PostalCode),
		Country:         city.Country,
		FederalDistrict: city.FederalDistrict,
		RegionType:      city.RegionType,
		Region:          city.Region,
		AreaType:        pgtypeutils.DecodeText(city.AreaType),
		Area:            pgtypeutils.DecodeText(city.Area),
		CityType:        pgtypeutils.DecodeText(city.CityType),
		City:            pgtypeutils.DecodeText(city.City),
		SettlementType:  pgtypeutils.DecodeText(city.SettlementType),
		Settlement:      pgtypeutils.DecodeText(city.Settlement),
		KladrID:         pgtypeutils.DecodeText(city.KladrID),
		FiasID:          city.FiasID,
		FiasLevel:       city.FiasLevel,
		CapitalMarker:   city.CapitalMarker,
		Okato:           pgtypeutils.DecodeText(city.Okato),
		Oktmo:           pgtypeutils.DecodeText(city.Oktmo),
		TaxOffice:       pgtypeutils.DecodeText(city.TaxOffice),
		Timezone:        pgtypeutils.DecodeText(city.Timezone),
		GeoLat:          decodeDecimal(city.GeoLat),
		GeoLon:          decodeDecimal(city.GeoLon),
		Population:      city.Population,
		FoundationYear:  pgtypeutils.DecodeInt2(city.FoundationYear),
	}
}

func NewFromModels(city []*models.City) []*CityResponse {
	res := make([]*CityResponse, 0, len(city))
	for _, c := range city {
		res = append(res, NewFromModel(c))
	}
	return res
}

func decodeDecimal(d decimal.NullDecimal) *decimal.Decimal {
	if !d.Valid {
		return nil
	}
	return &d.Decimal
}
