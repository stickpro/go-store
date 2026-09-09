// Package shipping is the carrier-agnostic delivery-points layer shared by the
// CDEK, Yandex Delivery and Russian Post integrations. Each carrier's service
// implements Provider; this package holds the provider registry and the one
// filtering implementation used by the /v1/delivery HTTP handler.
package shipping

import (
	"context"

	"github.com/stickpro/go-store/internal/dto"
)

// Provider is one cached delivery-points integration (CDEK, Yandex Delivery,
// Russian Post). Each provider fetches its carrier's full pickup points
// directory, keeps it warm in memory (see pkg/kvcache) and maps it to the
// carrier-agnostic dto.DeliveryPoint. Filtering and the HTTP layer are shared —
// see Filter and the /v1/delivery handler.
type Provider interface {
	// Code identifies the provider, e.g. "cdek", "yandex_delivery", "pochta".
	Code() string
	// Enabled reports whether the integration is turned on in config.
	Enabled() bool
	// RunCacheRefresher blocks, keeping the pickup-points cache warm until ctx is
	// cancelled.
	RunCacheRefresher(ctx context.Context)
	// Points returns every cached pickup point in the common shape. Returns an
	// empty slice when the integration is disabled or the cache is cold.
	Points() []dto.DeliveryPoint
}
