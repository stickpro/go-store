package yandexdelivery

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/logger"
)

// client talks to the Yandex Delivery logistics platform API
// (https://b2b-authproxy.taxi.yandex.net by default). It authenticates with a
// static Yandex OAuth token and has no cache of its own; caching is the
// responsibility of Service.
type client struct {
	cfg    config.YandexDeliveryConfig
	logger logger.Logger
	http   *http.Client
}

func newClient(cfg config.YandexDeliveryConfig, l logger.Logger) *client {
	return &client{
		cfg:    cfg,
		logger: l,
		http:   &http.Client{Timeout: cfg.RequestTimeout},
	}
}

// pickupPointsListRequest is the /api/b2b/platform/pickup-points/list body. An
// empty body returns every self-pickup point, ПВЗ and postomat; warehouses are
// dropped locally while mapping.
type pickupPointsListRequest struct{}

type apiPosition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type apiAddress struct {
	GeoID       int    `json:"geoId"`
	Country     string `json:"country"`
	Region      string `json:"region"`
	SubRegion   string `json:"subRegion"`
	Locality    string `json:"locality"`
	Street      string `json:"street"`
	House       string `json:"house"`
	Comment     string `json:"comment"`
	FullAddress string `json:"full_address"`
	PostalCode  string `json:"postal_code"`
}

type apiContact struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type apiTime struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
}

func (t apiTime) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hours, t.Minutes)
}

type apiScheduleRestriction struct {
	Days     []int   `json:"days"`
	TimeFrom apiTime `json:"time_from"`
	TimeTo   apiTime `json:"time_to"`
}

type apiSchedule struct {
	TimeZone     int                      `json:"time_zone"`
	Restrictions []apiScheduleRestriction `json:"restrictions"`
}

type apiPickupPoint struct {
	ID                string      `json:"id"`
	OperatorStationID string      `json:"operator_station_id"`
	OperatorID        string      `json:"operator_id"`
	Name              string      `json:"name"`
	Type              string      `json:"type"`
	Position          apiPosition `json:"position"`
	Address           apiAddress  `json:"address"`
	Instruction       string      `json:"instruction"`
	PaymentMethods    []string    `json:"payment_methods"`
	Contact           apiContact  `json:"contact"`
	Schedule          apiSchedule `json:"schedule"`

	IsYandexBranded     bool    `json:"is_yandex_branded"`
	IsMarketPartner     bool    `json:"is_market_partner"`
	IsPostOffice        bool    `json:"is_post_office"`
	AvailableForDropoff bool    `json:"available_for_dropoff"`
	DeactivationDate    *string `json:"deactivation_date"`
}

type pickupPointsListResponse struct {
	Points []apiPickupPoint `json:"points"`
}

func (p apiPickupPoint) toDTO() dto.YandexDeliveryPointDTO {
	schedule := make([]dto.YandexDeliveryScheduleDTO, 0, len(p.Schedule.Restrictions))
	for _, r := range p.Schedule.Restrictions {
		schedule = append(schedule, dto.YandexDeliveryScheduleDTO{
			Days:     r.Days,
			TimeFrom: r.TimeFrom.String(),
			TimeTo:   r.TimeTo.String(),
		})
	}

	return dto.YandexDeliveryPointDTO{
		Code:                p.ID,
		OperatorStationID:   p.OperatorStationID,
		OperatorID:          p.OperatorID,
		Name:                p.Name,
		Type:                p.Type,
		Country:             p.Address.Country,
		Region:              p.Address.Region,
		SubRegion:           p.Address.SubRegion,
		Locality:            p.Address.Locality,
		Street:              p.Address.Street,
		House:               p.Address.House,
		PostalCode:          p.Address.PostalCode,
		FullAddress:         p.Address.FullAddress,
		GeoID:               p.Address.GeoID,
		Latitude:            p.Position.Latitude,
		Longitude:           p.Position.Longitude,
		Instruction:         p.Instruction,
		Phone:               p.Contact.Phone,
		Email:               p.Contact.Email,
		PaymentMethods:      p.PaymentMethods,
		TimeZone:            p.Schedule.TimeZone,
		Schedule:            schedule,
		IsYandexBranded:     p.IsYandexBranded,
		IsMarketPartner:     p.IsMarketPartner,
		IsPostOffice:        p.IsPostOffice,
		AvailableForDropoff: p.AvailableForDropoff,
		DeactivationDate:    p.DeactivationDate,
	}
}

// listAllDeliveryPoints fetches the full pickup points list in a single call
// (the endpoint has no pagination) and maps every ПВЗ and postomat into the
// domain DTO. Warehouses (type "warehouse") are skipped.
func (c *client) listAllDeliveryPoints(ctx context.Context) ([]dto.YandexDeliveryPointDTO, error) {
	body, err := json.Marshal(pickupPointsListRequest{})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.cfg.BaseURL+"/api/b2b/platform/pickup-points/list", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build pickup-points request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.OauthToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request pickup-points: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yandex delivery pickup-points: unexpected status %d", resp.StatusCode)
	}

	// Drain the whole (tens-of-MB) body into memory first, then parse. Feeding
	// the streaming decoder straight off the network makes it read the body in
	// tiny increments while it works, dragging one request out to minutes — long
	// enough for the server to drop the connection mid-JSON ("unexpected end of
	// JSON input"). A plain ReadAll finishes the transfer in a few seconds.
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read pickup-points response: %w", err)
	}

	var parsed pickupPointsListResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode pickup-points response: %w", err)
	}

	points := make([]dto.YandexDeliveryPointDTO, 0, len(parsed.Points))
	for _, p := range parsed.Points {
		if p.Type == "warehouse" {
			continue
		}
		points = append(points, p.toDTO())
	}

	return points, nil
}
