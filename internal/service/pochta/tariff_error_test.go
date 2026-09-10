package pochta

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

// tariff.pochta.ru answers 400 with a full calculation body that carries an
// "errors" array when the chosen route/point can't be served. The client must
// surface that as a TariffError, and the rater as shipping.ErrRateUnavailable —
// not a raw JSON dump / 500.
func TestCalculateTariff_StructuredErrorOn400(t *testing.T) {
	const body = `{"id":27030,"name":"Посылка стандарт","errors":[{"msg":"Доставка в 983033 \"ЯНИНО-1 ППВЗ\" не осуществляется","type":1,"code":2006}]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := newClient(config.PochtaConfig{TariffURL: srv.URL, TariffObjectCode: 27030, RequestTimeout: 5 * time.Second}, logger.New())

	_, err := c.calculateTariff(context.Background(), pochtaTariffQuery{Object: 27030, From: "115551", To: "983033", WeightGrams: 600})

	var te *TariffError
	if !errors.As(err, &te) {
		t.Fatalf("want *TariffError, got %[1]T: %[1]v", err)
	}
	if te.Code != 2006 {
		t.Fatalf("TariffError.Code = %d, want 2006", te.Code)
	}

	cfg := &config.Config{}
	cfg.Pochta = config.PochtaConfig{Enabled: true, TariffURL: srv.URL, TariffObjectCode: 27030, RequestTimeout: 5 * time.Second}
	cfg.Shipping.OriginPostalCode = "115551"
	rp, ok := New(cfg, logger.New(), key_value.NewInMemory()).(shipping.RateProvider)
	if !ok {
		t.Fatal("pochta provider must implement shipping.RateProvider")
	}

	_, rerr := rp.Quote(context.Background(), shipping.RateQuery{ToPostalCode: "983033", Parcel: shipping.Parcel{WeightGrams: 600}})
	if !errors.Is(rerr, shipping.ErrRateUnavailable) {
		t.Fatalf("rater error = %v, want shipping.ErrRateUnavailable", rerr)
	}
}

func TestCalculateTariff_RawErrorWhenBodyUnparseable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>502 Bad Gateway</html>"))
	}))
	defer srv.Close()

	c := newClient(config.PochtaConfig{TariffURL: srv.URL, TariffObjectCode: 27030, RequestTimeout: 5 * time.Second}, logger.New())

	_, err := c.calculateTariff(context.Background(), pochtaTariffQuery{Object: 27030, From: "115551", To: "190000", WeightGrams: 600})
	var te *TariffError
	if errors.As(err, &te) {
		t.Fatalf("a non-JSON 502 must not become a TariffError: %v", err)
	}
	if err == nil {
		t.Fatal("expected an error for a 502 response")
	}
}
