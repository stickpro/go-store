package cdek

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/logger"
)

// deliveryPointsPageSize is the max page size accepted by the CDEK
// /v2/deliverypoints endpoint.
const deliveryPointsPageSize = 1000

// client talks to the CDEK API v2 (https://api.cdek.ru/v2 by default). It
// handles OAuth2 client_credentials token acquisition/renewal and paginated
// delivery point search. It has no cache of its own; caching is the
// responsibility of Service.
type client struct {
	cfg    config.CDEKConfig
	logger logger.Logger
	http   *http.Client

	tokenMu     sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func newClient(cfg config.CDEKConfig, l logger.Logger) *client {
	return &client{
		cfg:    cfg,
		logger: l,
		http:   &http.Client{Timeout: cfg.RequestTimeout},
	}
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// token returns a valid OAuth2 access token, requesting a new one when the
// cached one is missing or about to expire. Safe for concurrent use.
func (c *client) token(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.cfg.Account},
		"client_secret": {c.cfg.SecurePassword},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/oauth/token?"+form.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("build oauth request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("request oauth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cdek oauth token: unexpected status %d", resp.StatusCode)
	}

	var tok oauthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode oauth token response: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("cdek oauth token: empty access_token in response")
	}

	c.accessToken = tok.AccessToken
	// Renew a minute early so an in-flight request never races token expiry.
	c.tokenExpiry = time.Now().Add(time.Duration(tok.ExpiresIn)*time.Second - time.Minute)

	return c.accessToken, nil
}

type apiLocation struct {
	CountryCode string  `json:"country_code"`
	RegionCode  int     `json:"region_code"`
	Region      string  `json:"region"`
	CityCode    int     `json:"city_code"`
	City        string  `json:"city"`
	PostalCode  string  `json:"postal_code"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Address     string  `json:"address"`
	AddressFull string  `json:"address_full"`
}

type apiPhone struct {
	Number string `json:"number"`
}

type apiDeliveryPoint struct {
	Code           string      `json:"code"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	OwnerCode      string      `json:"owner_code"`
	Location       apiLocation `json:"location"`
	WorkTime       string      `json:"work_time"`
	Phones         []apiPhone  `json:"phones"`
	Email          string      `json:"email"`
	Note           string      `json:"note"`
	TakeOnly       bool        `json:"take_only"`
	IsHandout      bool        `json:"is_handout"`
	IsReception    bool        `json:"is_reception"`
	IsDressingRoom bool        `json:"is_dressing_room"`
	HaveCash       bool        `json:"have_cash"`
	HaveCashless   bool        `json:"have_cashless"`
	AllowedCod     bool        `json:"allowed_cod"`
	WeightMin      float64     `json:"weight_min"`
	WeightMax      float64     `json:"weight_max"`
}

func (p apiDeliveryPoint) toDTO() dto.CDEKDeliveryPointDTO {
	phones := make([]string, 0, len(p.Phones))
	for _, ph := range p.Phones {
		if ph.Number != "" {
			phones = append(phones, ph.Number)
		}
	}

	return dto.CDEKDeliveryPointDTO{
		Code:           p.Code,
		Name:           p.Name,
		Type:           p.Type,
		OwnerCode:      p.OwnerCode,
		CountryCode:    p.Location.CountryCode,
		RegionCode:     p.Location.RegionCode,
		Region:         p.Location.Region,
		CityCode:       p.Location.CityCode,
		City:           p.Location.City,
		PostalCode:     p.Location.PostalCode,
		Address:        p.Location.Address,
		AddressFull:    p.Location.AddressFull,
		Latitude:       p.Location.Latitude,
		Longitude:      p.Location.Longitude,
		WorkTime:       p.WorkTime,
		Phones:         phones,
		Email:          p.Email,
		Note:           p.Note,
		TakeOnly:       p.TakeOnly,
		IsHandout:      p.IsHandout,
		IsReception:    p.IsReception,
		IsDressingRoom: p.IsDressingRoom,
		HaveCash:       p.HaveCash,
		HaveCashless:   p.HaveCashless,
		AllowedCod:     p.AllowedCod,
		WeightMin:      p.WeightMin,
		WeightMax:      p.WeightMax,
	}
}

// fetchDeliveryPointsPage fetches one page of /v2/deliverypoints.
func (c *client) fetchDeliveryPointsPage(ctx context.Context, countryCode string, page int) ([]apiDeliveryPoint, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	q := url.Values{
		"size": {strconv.Itoa(deliveryPointsPageSize)},
		"page": {strconv.Itoa(page)},
	}
	if countryCode != "" {
		q.Set("country_code", countryCode)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+"/deliverypoints?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build deliverypoints request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request deliverypoints: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cdek deliverypoints: unexpected status %d", resp.StatusCode)
	}

	var points []apiDeliveryPoint
	if err := json.NewDecoder(resp.Body).Decode(&points); err != nil {
		return nil, fmt.Errorf("decode deliverypoints response: %w", err)
	}

	return points, nil
}

// maxDeliveryPointsPages caps how many pages searchAllDeliveryPoints will
// fetch, as a safety net against the CDEK API never returning an empty page
// (it does not always fill a page to deliveryPointsPageSize, so page count
// alone can't predict when the real data ends).
const maxDeliveryPointsPages = 200

// searchAllDeliveryPoints pages through /v2/deliverypoints until an empty
// page is returned and maps every point into the domain DTO. CDEK does not
// always fill a page to deliveryPointsPageSize even when more data follows,
// so a short page is not a reliable end-of-data signal — only an empty one
// is. countryCode is also re-checked locally since the API's own country_code
// filter has been observed to let other countries' points through.
func (c *client) searchAllDeliveryPoints(ctx context.Context, countryCode string) ([]dto.CDEKDeliveryPointDTO, error) {
	var all []dto.CDEKDeliveryPointDTO

	for page := 0; page < maxDeliveryPointsPages; page++ {
		points, err := c.fetchDeliveryPointsPage(ctx, countryCode, page)
		if err != nil {
			return nil, fmt.Errorf("fetch page %d: %w", page, err)
		}
		if len(points) == 0 {
			return all, nil
		}

		for _, p := range points {
			if countryCode != "" && p.Location.CountryCode != countryCode {
				continue
			}
			all = append(all, p.toDTO())
		}
	}

	return nil, fmt.Errorf("exceeded %d pages without an empty page from cdek deliverypoints", maxDeliveryPointsPages)
}
