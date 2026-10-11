package cdek

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/logger"
)

const (
	// citiesPageSize is the page size used for /v2/location/cities.
	citiesPageSize = 1000
	// maxCitiesPages is a safety net against the API never returning an
	// empty page; Russia is ~145 pages.
	maxCitiesPages = 1000
)

type apiCity struct {
	Code        int     `json:"code"`
	City        string  `json:"city"`
	FiasGUID    string  `json:"fias_guid"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	TimeZone    string  `json:"time_zone"`
}

// FetchCities downloads the whole CDEK locations directory for one country
// (ISO 3166-1 alpha-2, e.g. "RU"). It is meant for one-off console imports,
// not request handling: the directory is ~143k rows for Russia.
func FetchCities(ctx context.Context, cfg config.CDEKConfig, l logger.Logger, countryCode string) ([]dto.CDEKCityDTO, error) {
	c := newClient(cfg, l)

	var all []dto.CDEKCityDTO
	for page := range maxCitiesPages {
		cities, err := c.fetchCitiesPage(ctx, countryCode, page)
		if err != nil {
			return nil, fmt.Errorf("fetch page %d: %w", page, err)
		}
		if len(cities) == 0 {
			return all, nil
		}
		for _, city := range cities {
			if city.CountryCode != countryCode {
				continue
			}
			all = append(all, dto.CDEKCityDTO{
				Code:      city.Code,
				City:      city.City,
				FiasGUID:  city.FiasGUID,
				Region:    city.Region,
				Latitude:  city.Latitude,
				Longitude: city.Longitude,
				TimeZone:  city.TimeZone,
			})
		}
	}

	return nil, fmt.Errorf("exceeded %d pages without an empty page from cdek location/cities", maxCitiesPages)
}

// fetchCitiesPage fetches one page of /v2/location/cities.
func (c *client) fetchCitiesPage(ctx context.Context, countryCode string, page int) ([]apiCity, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	q := url.Values{
		"country_codes": {countryCode},
		"size":          {strconv.Itoa(citiesPageSize)},
		"page":          {strconv.Itoa(page)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+"/location/cities?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build location/cities request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request location/cities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cdek location/cities: unexpected status %d", resp.StatusCode)
	}

	var cities []apiCity
	if err := json.NewDecoder(resp.Body).Decode(&cities); err != nil {
		return nil, fmt.Errorf("decode location/cities response: %w", err)
	}
	return cities, nil
}
