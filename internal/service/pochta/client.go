package pochta

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/logger"
)

// errBodyLimit caps how much of an error response body is read into the error
// message.
const errBodyLimit = 2 << 10

// client talks to the Russian Post Otpravka API. It authenticates with the
// application AccessToken plus the Basic-encoded pochta.ru account and has no
// cache of its own; caching is the responsibility of Service. Its only job is
// the one-shot "ОПС passport" export that yields the full pickup points list.
type client struct {
	cfg    config.PochtaConfig
	logger logger.Logger
	http   *http.Client

	userAuth string // precomputed base64(login:password) for X-User-Authorization
}

func newClient(cfg config.PochtaConfig, l logger.Logger) *client {
	return &client{
		cfg:      cfg,
		logger:   l,
		http:     &http.Client{Timeout: cfg.RequestTimeout},
		userAuth: base64.StdEncoding.EncodeToString([]byte(cfg.UserLogin + ":" + cfg.UserPassword)),
	}
}

// --- ОПС passport JSON shape ---------------------------------------------------

type apiPassport struct {
	Elements []apiPassportElement `json:"passportElements"`
}

type apiPassportElement struct {
	Address struct {
		Index  string `json:"index"`
		Region string `json:"region"`
		Area   string `json:"area"`
		Place  string `json:"place"`
		Street string `json:"street"`
		House  string `json:"house"`
		Office string `json:"office"`
	} `json:"address"`
	AddressFias struct {
		Ads string `json:"ads"`
	} `json:"addressFias"`
	BrandName   string `json:"brandName"`
	Ecom        string `json:"ecom"`
	EcomOptions struct {
		BrandName         string  `json:"brandName"`
		Getto             string  `json:"getto"`
		CardPayment       bool    `json:"cardPayment"`
		CashPayment       bool    `json:"cashPayment"`
		WeightLimit       float64 `json:"weightLimit"`
		WithFitting       bool    `json:"withFitting"`
		ContentsChecking  bool    `json:"contentsChecking"`
		PartialRedemption bool    `json:"partialRedemption"`
		ReturnAvailable   bool    `json:"returnAvailable"`
	} `json:"ecomOptions"`
	Latitude  string   `json:"latitude"`
	Longitude string   `json:"longitude"`
	Type      string   `json:"type"`
	WorkTime  []string `json:"workTime"`
}

func (e apiPassportElement) toDTO() dto.PochtaDeliveryPointDTO {
	lat, _ := strconv.ParseFloat(e.Latitude, 64)
	lon, _ := strconv.ParseFloat(e.Longitude, 64)

	brand := e.BrandName
	if brand == "" {
		brand = e.EcomOptions.BrandName
	}
	name := brand
	if name == "" {
		name = "Отделение " + e.Address.Index
	}

	address := e.AddressFias.Ads
	if address == "" {
		address = strings.TrimSpace(strings.Join(nonEmpty(
			e.Address.Index, e.Address.Region, e.Address.Area, e.Address.Place,
			e.Address.Street, e.Address.House, e.Address.Office,
		), ", "))
	}

	return dto.PochtaDeliveryPointDTO{
		Code:              e.Address.Index,
		PostalCode:        e.Address.Index,
		Name:              name,
		Type:              e.Type,
		BrandName:         brand,
		Region:            e.Address.Region,
		Area:              e.Address.Area,
		Place:             e.Address.Place,
		Street:            e.Address.Street,
		House:             e.Address.House,
		Office:            e.Address.Office,
		Address:           address,
		Description:       e.EcomOptions.Getto,
		Latitude:          lat,
		Longitude:         lon,
		Ecom:              e.Ecom == "1",
		CardPayment:       e.EcomOptions.CardPayment,
		CashPayment:       e.EcomOptions.CashPayment,
		WeightLimitKg:     e.EcomOptions.WeightLimit,
		WithFitting:       e.EcomOptions.WithFitting,
		ContentsChecking:  e.EcomOptions.ContentsChecking,
		PartialRedemption: e.EcomOptions.PartialRedemption,
		ReturnAvailable:   e.EcomOptions.ReturnAvailable,
		WorkTime:          e.WorkTime,
	}
}

func nonEmpty(vals ...string) []string {
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// listAllDeliveryPoints downloads the "ОПС passport" ZIP export, unpacks the
// single JSON file it contains and maps every element into the domain DTO.
func (c *client) listAllDeliveryPoints(ctx context.Context) ([]dto.PochtaDeliveryPointDTO, error) {
	q := url.Values{"type": {c.unloadType()}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.cfg.BaseURL+"/1.0/unloading-passport/zip?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build passport request: %w", err)
	}
	req.Header.Set("Authorization", "AccessToken "+c.cfg.AccessToken)
	req.Header.Set("X-User-Authorization", "Basic "+c.userAuth)
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request passport: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, errBodyLimit))
		return nil, fmt.Errorf("pochta passport: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	archive, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read passport archive: %w", err)
	}

	return parsePassportArchive(archive)
}

func (c *client) unloadType() string {
	if c.cfg.UnloadType == "" {
		return "ALL"
	}
	return c.cfg.UnloadType
}

// parsePassportArchive unpacks the ОПС passport ZIP (a single JSON file with a
// top-level "passportElements" array) into domain DTOs.
func parsePassportArchive(archive []byte) ([]dto.PochtaDeliveryPointDTO, error) {
	zr, err := zip.NewReader(bytesReaderAt(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open passport zip: %w", err)
	}
	if len(zr.File) == 0 {
		return nil, fmt.Errorf("pochta passport zip is empty")
	}

	f, err := zr.File[0].Open()
	if err != nil {
		return nil, fmt.Errorf("open %s in passport zip: %w", zr.File[0].Name, err)
	}
	defer f.Close()

	var passport apiPassport
	if err := json.NewDecoder(f).Decode(&passport); err != nil {
		return nil, fmt.Errorf("decode passport json: %w", err)
	}

	points := make([]dto.PochtaDeliveryPointDTO, 0, len(passport.Elements))
	for _, e := range passport.Elements {
		if e.Address.Index == "" {
			continue
		}
		points = append(points, e.toDTO())
	}
	return points, nil
}

// bytesReaderAt adapts a byte slice to io.ReaderAt for archive/zip.
type bytesReaderAt []byte

func (b bytesReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= int64(len(b)) {
		return 0, io.EOF
	}
	n := copy(p, b[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// --- shipping cost calculation (public tariff.pochta.ru, no auth) -------------

type pochtaTariffQuery struct {
	Object       int
	From         string
	To           string
	WeightGrams  int
	LengthCM     int
	WidthCM      int
	HeightCM     int
	PackType     int
	SumOCKopecks int64
}

type pochtaTariffResult struct {
	Name         string
	TotalKopecks int64
	NoVATKopecks int64
	WeightBilled int
	MinDays      int
	MaxDays      int
	Deadline     string
}

// TariffError is a structured error from the tariff calculator: the route,
// parcel or destination the customer picked cannot be served on this tariff
// (e.g. the point does not accept this mail category). It is a business
// condition, not an infrastructure failure — the rater maps it to
// shipping.ErrRateUnavailable.
type TariffError struct {
	Code int
	Msg  string
}

func (e *TariffError) Error() string {
	return fmt.Sprintf("pochta tariff %d: %s", e.Code, e.Msg)
}

// pochtaTariffResponse is the tariff.pochta.ru JSON shape. Amounts are kopecks.
type pochtaTariffResponse struct {
	Name      string `json:"name"`
	Pay       int64  `json:"pay"`
	PayNDS    int64  `json:"paynds"`
	WeightPay int    `json:"weightpay"`
	Delivery  struct {
		Min      int    `json:"min"`
		Max      int    `json:"max"`
		Deadline string `json:"deadline"`
	} `json:"delivery"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors"`
}

// calculateTariff calls GET tariff.pochta.ru/v2/calculate/tariff/delivery.
func (c *client) calculateTariff(ctx context.Context, q pochtaTariffQuery) (*pochtaTariffResult, error) {
	base := c.cfg.TariffURL
	if base == "" {
		base = "https://tariff.pochta.ru"
	}

	params := url.Values{}
	params.Set("json", "")
	params.Set("object", strconv.Itoa(q.Object))
	params.Set("from", q.From)
	params.Set("to", q.To)
	params.Set("weight", strconv.Itoa(q.WeightGrams))
	if q.LengthCM > 0 && q.WidthCM > 0 && q.HeightCM > 0 {
		params.Set("size", fmt.Sprintf("%dx%dx%d", q.LengthCM, q.WidthCM, q.HeightCM))
	}
	if q.PackType > 0 {
		params.Set("pack", strconv.Itoa(q.PackType))
	}
	if q.SumOCKopecks > 0 {
		params.Set("sumoc", strconv.FormatInt(q.SumOCKopecks, 10))
	}

	reqURL := base + "/v2/calculate/tariff/delivery?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build tariff request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request tariff: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	// tariff.pochta.ru answers 400 with a structured {"errors":[...]} body when
	// the route / parcel / pickup point cannot be served. Parse it first so that
	// turns into a clean TariffError, not a raw JSON dump.
	var parsed pochtaTariffResponse
	jsonErr := json.Unmarshal(raw, &parsed)
	if jsonErr == nil && len(parsed.Errors) > 0 {
		return nil, &TariffError{Code: parsed.Errors[0].Code, Msg: parsed.Errors[0].Msg}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pochta tariff: status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if jsonErr != nil {
		return nil, fmt.Errorf("decode tariff response: %w", jsonErr)
	}

	total := parsed.PayNDS
	if total == 0 {
		total = parsed.Pay
	}
	return &pochtaTariffResult{
		Name:         parsed.Name,
		TotalKopecks: total,
		NoVATKopecks: parsed.Pay,
		WeightBilled: parsed.WeightPay,
		MinDays:      parsed.Delivery.Min,
		MaxDays:      parsed.Delivery.Max,
		Deadline:     parsed.Delivery.Deadline,
	}, nil
}
