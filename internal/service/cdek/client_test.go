package cdek

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/pkg/logger"
)

// TestClient_Token_LiveCredentials verifies that the configured CDEK API
// credentials (account/secure_password) actually work by requesting a real
// OAuth2 access token from the CDEK API. It makes a live network call, so it
// is skipped unless CDEK_ACCOUNT and CDEK_SECURE_PASSWORD are set in the
// environment.
//
// Run it explicitly against the CDEK test contour, e.g.:
//
//	CDEK_ACCOUNT=... CDEK_SECURE_PASSWORD=... CDEK_BASE_URL=https://api.edu.cdek.ru/v2 \
//	  go test ./internal/service/cdek/... -run TestClient_Token_LiveCredentials -v

func TestClient_Token_LiveCredentials(t *testing.T) {
	account := os.Getenv("CDEK_ACCOUNT")
	secret := os.Getenv("CDEK_SECURE_PASSWORD")
	if account == "" || secret == "" {
		t.Skip("CDEK_ACCOUNT / CDEK_SECURE_PASSWORD not set; skipping live CDEK credentials check")
	}

	baseURL := os.Getenv("CDEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.cdek.ru/v2"
	}

	cfg := config.CDEKConfig{
		Enabled:        true,
		BaseURL:        baseURL,
		Account:        account,
		SecurePassword: secret,
		RequestTimeout: 15 * time.Second,
	}

	c := newClient(cfg, logger.New())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	token, err := c.token(ctx)
	if err != nil {
		t.Fatalf("cdek oauth token request failed — check CDEK_ACCOUNT/CDEK_SECURE_PASSWORD: %v", err)
	}
	if token == "" {
		t.Fatal("cdek oauth token request returned an empty access_token")
	}
}

// TestClient_CalculateTariff_LiveCredentials calculates a real CDEK tariff.
// Skipped unless CDEK_ACCOUNT / CDEK_SECURE_PASSWORD are set.
func TestClient_CalculateTariff_LiveCredentials(t *testing.T) {
	account := os.Getenv("CDEK_ACCOUNT")
	secret := os.Getenv("CDEK_SECURE_PASSWORD")
	if account == "" || secret == "" {
		t.Skip("CDEK_ACCOUNT / CDEK_SECURE_PASSWORD not set")
	}
	baseURL := os.Getenv("CDEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.cdek.ru/v2"
	}

	c := newClient(config.CDEKConfig{
		Enabled: true, BaseURL: baseURL, Account: account, SecurePassword: secret,
		RequestTimeout: 20 * time.Second,
	}, logger.New())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := c.calculateTariff(ctx, cdekTariffRequest{
		Type: 1, TariffCode: 136,
		FromLocation: cdekTariffLocation{PostalCode: "115551"},
		ToLocation:   cdekTariffLocation{PostalCode: "190000"},
		Packages:     []cdekTariffPackage{{Weight: 1000, Length: 20, Width: 15, Height: 10}},
	})
	if err != nil {
		t.Fatalf("cdek tariff calc failed: %v", err)
	}
	if !res.TotalSum.IsPositive() {
		t.Fatalf("expected a positive total_sum, got %s", res.TotalSum)
	}
	t.Logf("CDEK 136 Мск→СПб 1kg: %s %s, %d-%d days", res.TotalSum, res.Currency, res.PeriodMin, res.PeriodMax)
}
