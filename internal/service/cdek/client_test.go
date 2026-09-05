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
