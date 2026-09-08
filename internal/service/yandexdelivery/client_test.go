package yandexdelivery

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/pkg/logger"
)

// TestClient_ListAllDeliveryPoints_LiveCredentials verifies that the configured
// Yandex Delivery OAuth token actually works by requesting the real pickup
// points list from the Yandex Delivery API. It makes a live network call, so it
// is skipped unless YANDEX_DELIVERY_OAUTH_TOKEN is set in the environment.
//
// Run it explicitly, e.g.:
//
//	YANDEX_DELIVERY_OAUTH_TOKEN=... \
//	  go test ./internal/service/yandexdelivery/... -run TestClient_ListAllDeliveryPoints_LiveCredentials -v
func TestClient_ListAllDeliveryPoints_LiveCredentials(t *testing.T) {
	token := os.Getenv("YANDEX_DELIVERY_OAUTH_TOKEN")
	if token == "" {
		t.Skip("YANDEX_DELIVERY_OAUTH_TOKEN not set; skipping live Yandex Delivery credentials check")
	}

	baseURL := os.Getenv("YANDEX_DELIVERY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://b2b-authproxy.taxi.yandex.net"
	}

	cfg := config.YandexDeliveryConfig{
		Enabled:        true,
		BaseURL:        baseURL,
		OauthToken:     token,
		RequestTimeout: 240 * time.Second,
	}

	c := newClient(cfg, logger.New())

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	points, err := c.listAllDeliveryPoints(ctx)
	if err != nil {
		t.Fatalf("yandex delivery pickup-points request failed — check YANDEX_DELIVERY_OAUTH_TOKEN: %v", err)
	}
	if len(points) == 0 {
		t.Fatal("yandex delivery pickup-points request returned no points")
	}
}
