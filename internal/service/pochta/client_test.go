package pochta

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/pkg/logger"
)

// buildPassportZip wraps a JSON string in a one-file ZIP archive shaped like the
// Russian Post ОПС passport export.
func buildPassportZip(t *testing.T, jsonBody string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("ALL_test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(jsonBody)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParsePassportArchive(t *testing.T) {
	zipBytes := buildPassportZip(t, `{"passportElements":[
		{"address":{"index":"115551","region":"Москва г","place":"Москва"},"addressFias":{"ads":"115551 Москва Домодедовская ул 20"},"ecom":"1","ecomOptions":{"weightLimit":20.0,"cardPayment":true},"latitude":"55.612772","longitude":"37.704862","type":"ГОПС","workTime":["пн, открыто: 08:00 - 20:00"]},
		{"address":{"index":""},"type":"ГОПС"}
	]}`)

	points, err := parsePassportArchive(zipBytes)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 {
		t.Fatalf("want 1 point (blank index dropped), got %d", len(points))
	}
	p := points[0]
	if p.Code != "115551" || p.Type != "ГОПС" || !p.Ecom || !p.CardPayment || p.WeightLimitKg != 20.0 {
		t.Fatalf("unexpected mapping: %+v", p)
	}
	if p.Latitude == 0 || p.Longitude == 0 {
		t.Fatalf("coordinates not parsed: %+v", p)
	}
}

// TestClient_ListAllDeliveryPoints_LiveCredentials verifies that the configured
// Russian Post Otpravka credentials work by exporting the real ОПС passport and
// checking a non-empty pickup points list comes back. It makes a live network
// call (a multi-MB download), so it is skipped unless POCHTA_ACCESS_TOKEN,
// POCHTA_USER_LOGIN and POCHTA_USER_PASSWORD are set in the environment.
//
// Run it explicitly, e.g.:
//
//	POCHTA_ACCESS_TOKEN=... POCHTA_USER_LOGIN=... POCHTA_USER_PASSWORD=... \
//	  go test ./internal/service/pochta/... -run TestClient_ListAllDeliveryPoints_LiveCredentials -v
func TestClient_ListAllDeliveryPoints_LiveCredentials(t *testing.T) {
	token := os.Getenv("POCHTA_ACCESS_TOKEN")
	login := os.Getenv("POCHTA_USER_LOGIN")
	password := os.Getenv("POCHTA_USER_PASSWORD")
	if token == "" || login == "" || password == "" {
		t.Skip("POCHTA_ACCESS_TOKEN / POCHTA_USER_LOGIN / POCHTA_USER_PASSWORD not set; skipping live Russian Post credentials check")
	}

	baseURL := os.Getenv("POCHTA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://otpravka-api.pochta.ru"
	}

	cfg := config.PochtaConfig{
		Enabled:        true,
		BaseURL:        baseURL,
		AccessToken:    token,
		UserLogin:      login,
		UserPassword:   password,
		UnloadType:     "ALL",
		RequestTimeout: 180 * time.Second,
	}

	c := newClient(cfg, logger.New())

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	points, err := c.listAllDeliveryPoints(ctx)
	if err != nil {
		t.Fatalf("pochta passport export failed — check POCHTA_ACCESS_TOKEN/POCHTA_USER_LOGIN/POCHTA_USER_PASSWORD: %v", err)
	}
	if len(points) < 1000 {
		t.Fatalf("pochta passport export returned only %d points, expected the full directory", len(points))
	}
}

// TestClient_CalculateTariff_Live hits the public Russian Post tariff
// calculator (no credentials needed). Skipped in -short mode.
func TestClient_CalculateTariff_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("-short")
	}
	c := newClient(config.PochtaConfig{
		TariffURL: "https://tariff.pochta.ru", TariffObjectCode: 27030,
		RequestTimeout: 20 * time.Second,
	}, logger.New())

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	res, err := c.calculateTariff(ctx, pochtaTariffQuery{
		Object: 27030, From: "115551", To: "190000",
		WeightGrams: 1000, LengthCM: 20, WidthCM: 15, HeightCM: 10, PackType: 30,
	})
	if err != nil {
		t.Fatalf("pochta tariff calc failed: %v", err)
	}
	if res.TotalKopecks <= 0 {
		t.Fatalf("expected a positive tariff, got %d kopecks", res.TotalKopecks)
	}
	t.Logf("Pochta 27030 Мск→СПб 1kg: %d kopecks, %d-%d days", res.TotalKopecks, res.MinDays, res.MaxDays)
}
