package geo_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type GeoResponse struct {
	City string `json:"city"`
} //@name GeoResponse

type CityResponse struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	Address         string          `db:"address" json:"address"`
	PostalCode      string          `db:"postal_code" json:"postal_code"`
	Country         string          `db:"country" json:"country"`
	FederalDistrict string          `db:"federal_district" json:"federal_district"`
	RegionType      string          `db:"region_type" json:"region_type"`
	Region          string          `db:"region" json:"region"`
	AreaType        *string         `db:"area_type" json:"area_type"`
	Area            *string         `db:"area" json:"area"`
	CityType        string          `db:"city_type" json:"city_type"`
	City            string          `db:"city" json:"city"`
	SettlementType  *string         `db:"settlement_type" json:"settlement_type"`
	Settlement      *string         `db:"settlement" json:"settlement"`
	KladrID         string          `db:"kladr_id" json:"kladr_id"`
	FiasID          uuid.UUID       `db:"fias_id" json:"fias_id"`
	FiasLevel       int16           `db:"fias_level" json:"fias_level"`
	CapitalMarker   int16           `db:"capital_marker" json:"capital_marker"`
	Okato           string          `db:"okato" json:"okato"`
	Oktmo           string          `db:"oktmo" json:"oktmo"`
	TaxOffice       string          `db:"tax_office" json:"tax_office"`
	Timezone        string          `db:"timezone" json:"timezone"`
	GeoLat          decimal.Decimal `db:"geo_lat" json:"geo_lat"`
	GeoLon          decimal.Decimal `db:"geo_lon" json:"geo_lon"`
	Population      int64           `db:"population" json:"population"`
	FoundationYear  int16           `db:"foundation_year" json:"foundation_year"`
}

func NewFromModel(city *models.City) *CityResponse {
	return &CityResponse{
		ID:              city.ID,
		Address:         city.Address,
		PostalCode:      city.PostalCode,
		Country:         city.Country,
		FederalDistrict: city.FederalDistrict,
		RegionType:      city.RegionType,
		Region:          city.Region,
		AreaType:        pgtypeutils.DecodeText(city.AreaType),
		Area:            pgtypeutils.DecodeText(city.Area),
		CityType:        city.CityType,
		City:            city.City,
		SettlementType:  pgtypeutils.DecodeText(city.SettlementType),
		Settlement:      pgtypeutils.DecodeText(city.Settlement),
		KladrID:         city.KladrID,
		FiasID:          city.FiasID,
		FiasLevel:       city.FiasLevel,
		CapitalMarker:   city.CapitalMarker,
		Okato:           city.Okato,
		Oktmo:           city.Oktmo,
		TaxOffice:       city.TaxOffice,
		Timezone:        city.Timezone,
		GeoLat:          city.GeoLat,
		GeoLon:          city.GeoLon,
		Population:      city.Population,
		FoundationYear:  city.FoundationYear,
	}
}

func NewFromModels(city []*models.City) []*CityResponse {
	res := make([]*CityResponse, 0, len(city))
	for _, c := range city {
		res = append(res, NewFromModel(c))
	}
	return res
}
