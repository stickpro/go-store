package shipping

import (
	"context"
	"errors"

	"github.com/stickpro/go-store/internal/dto"
)

// ErrRatesNotSupported is returned by a RateProvider that has no shipping-cost
// calculation wired up. QuoteAll skips such carriers silently.
var ErrRatesNotSupported = errors.New("shipping: rate calculation not supported by this provider")

// ErrRateUnavailable means the carrier answered but cannot serve this particular
// route / parcel / pickup point (e.g. the tariff does not deliver to the chosen
// point). It is a business condition, not an infrastructure failure: QuoteAll
// omits the carrier from the options and checkout turns it into a 422, not a 500.
var ErrRateUnavailable = errors.New("shipping: carrier has no rate for this route/parcel")

// RateProvider is a Provider that can also quote shipping cost. Every carrier
// package implements it; a carrier with no calculator returns
// ErrRatesNotSupported from Quote.
type RateProvider interface {
	Provider
	// Quote returns this carrier's delivery options for the parcel and route in
	// q, cheapest first is not required (the caller sorts).
	Quote(ctx context.Context, q RateQuery) ([]dto.ShippingRate, error)
}

// RateQuery is the normalised input a RateProvider works with. Exactly one of
// ToPointCode / ToPostalCode identifies the destination.
type RateQuery struct {
	FromPostalCode string
	ToPostalCode   string
	ToPointCode    string // a pickup point Code from Points(); carrier resolves its location
	DeliveryType   string // "pickup" | "courier" | "" (carrier default)
	Parcel         Parcel
}
